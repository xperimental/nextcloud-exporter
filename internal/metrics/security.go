package metrics

// Проверка того, содержит ли доступное обновление Nextcloud исправления безопасности.
//
// Nextcloud сообщает только "есть более новая версия", но не говорит, чинит ли она
// уязвимости. Единственный источник, способный ответить на этот вопрос, — список
// security advisories репозитория nextcloud/security-advisories на GitHub. Глобальный
// GitHub Advisory Database и OSV.dev по Nextcloud пусты, а release notes не упоминают
// исправления безопасности вовсе, потому что такие правки мержатся из приватных форков
// и в changelog не попадают.
//
// Список выкачивается в фоне раз в 12 часов и держится в памяти. Скрейп Prometheus
// никогда не ждёт GitHub: он читает уже готовый снимок.
//
// Логика намеренно смещена в сторону "сообщить об обновлении": если данных нет или они
// непригодны, метрика остаётся равной единице, то есть ведёт себя как до появления этой
// проверки.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// advisoriesURL — точка входа в список advisories.
	//
	// Сортировка по дате публикации по возрастанию выбрана не случайно: пагинация здесь
	// курсорная, и при таком порядке новые записи дописываются в конец, а уже отданные
	// страницы не сдвигаются. При сортировке по updated запись, отредактированная в
	// середине обхода, может быть пропущена.
	advisoriesURL = "https://api.github.com/repos/nextcloud/security-advisories/security-advisories" +
		"?per_page=100&sort=published&direction=asc"

	// advisoriesRefreshInterval — как часто обновлять список.
	//
	// GitHub без токена разрешает 60 запросов в час на IP-адрес. Полная выкачка стоит
	// три запроса, поэтому раз в 12 часов это шесть запросов в сутки — с большим запасом.
	advisoriesRefreshInterval = 12 * time.Hour

	// advisoriesRetryInterval — пауза перед повтором после неудачной выкачки.
	advisoriesRetryInterval = 15 * time.Minute

	// advisoriesMaxAge — после какого возраста снимок считается непригодным.
	//
	// Список advisories только пополняется, поэтому опасность устаревшего снимка в том,
	// что в нём не будет свежей записи, и метрика покажет ноль при реально нужном
	// обновлении безопасности. Это худшее из возможных поведений, поэтому у снимка есть
	// жёсткий срок годности: четыре пропущенных цикла обновления, и он перестаёт
	// использоваться.
	advisoriesMaxAge = 48 * time.Hour

	// advisoriesHTTPTimeout — таймаут на один запрос к GitHub.
	//
	// Это отдельный таймаут, не связанный с --timeout: тот задаёт ожидание Nextcloud
	// внутри скрейпа и обычно измеряется секундами, а здесь запрос идёт в фоне и
	// торопиться некуда.
	advisoriesHTTPTimeout = 30 * time.Second

	// advisoriesMaxPages — предохранитель от бесконечного обхода страниц.
	// Сейчас страниц три, так что запас десятикратный.
	advisoriesMaxPages = 10
)

// version — версия, разобранная на числовые компоненты.
//
// Отдельный тип нужен потому, что версии здесь приходят с разным числом компонентов:
// Nextcloud сообщает свою версию четырьмя числами ("34.0.2.1"), а advisories указывают
// исправленные версии тремя ("34.0.1"). Сравнивать их как строки нельзя: "34.0.10"
// меньше "34.0.9" в лексикографическом порядке, хотя по смыслу больше.
type version []int

// advisory — нужное подмножество полей одной записи из ответа GitHub.
// Остальные поля ответа отбрасываются: encoding/json игнорирует всё, чего нет в структуре.
type advisory struct {
	GHSAID string `json:"ghsa_id"`

	// WithdrawnAt — указатель, а не строка, чтобы отличить "поле отсутствует или null"
	// (nil) от заполненного значения. Заполненное означает, что запись отозвана и её
	// нужно игнорировать.
	WithdrawnAt *string `json:"withdrawn_at"`

	Vulnerabilities []advisoryVulnerability `json:"vulnerabilities"`
}

// advisoryVulnerability — сведения об одном затронутом продукте внутри записи.
type advisoryVulnerability struct {
	Package advisoryPackage `json:"package"`

	// VulnerableVersionRange и PatchedVersions — свободный текст, заполняемый вручную,
	// поэтому разбор должен быть терпимым к мусору. Подробности в parsePatchedVersions.
	VulnerableVersionRange string `json:"vulnerable_version_range"`
	PatchedVersions        string `json:"patched_versions"`
}

type advisoryPackage struct {
	Ecosystem string `json:"ecosystem"`
	Name      string `json:"name"`
}

// advisoryFix — итог разбора одной записи: в каких версиях уязвимость исправлена.
// Разбор делается один раз при выкачке, а не на каждом скрейпе.
type advisoryFix struct {
	ghsaID          string
	patchedVersions []version
}

// advisorySnapshot — неизменяемый результат одной успешной выкачки.
//
// После публикации снимок никогда не меняется: следующая выкачка создаёт новый и
// заменяет указатель. Благодаря этому читателям достаточно скопировать указатель под
// блокировкой, а дальше работать без неё.
type advisorySnapshot struct {
	fixes     []advisoryFix
	fetchedAt time.Time
}

// SecurityChecker хранит список advisories и отвечает на вопрос, несёт ли обновление
// исправления безопасности. Нулевое значение непригодно, создавайте через NewSecurityChecker.
type SecurityChecker struct {
	log       logrus.FieldLogger
	userAgent string
	client    *http.Client

	// url — адрес первой страницы списка. Отдельным полем, а не константой напрямую,
	// чтобы тесты могли подставить локальный сервер вместо GitHub.
	url string

	// mu защищает snapshot. RWMutex, а не Mutex, потому что читателей много (каждый
	// скрейп) и они друг друга не блокируют, а писатель один и появляется раз в 12 часов.
	mu       sync.RWMutex
	snapshot *advisorySnapshot
}

// NewSecurityChecker создаёт проверяльщик. Обращений к сети не делает — за это отвечает Start.
func NewSecurityChecker(log logrus.FieldLogger, userAgent string) *SecurityChecker {
	return &SecurityChecker{
		log:       log,
		userAgent: userAgent,
		url:       advisoriesURL,
		client: &http.Client{
			Timeout: advisoriesHTTPTimeout,
			// Transport намеренно не задан, используется стандартный, то есть проверка
			// TLS-сертификата включена. Флаг --tls-skip-verify относится к серверу
			// Nextcloud пользователя и на GitHub распространяться не должен.
		},
	}
}

// Start запускает фоновое обновление списка и сразу возвращает управление.
//
// Обновление живёт в отдельной горутине, поэтому скрейп Prometheus никогда не ждёт
// GitHub. Пока первая выкачка не закончилась, снимка нет, Check отвечает "не знаю", и
// метрика ведёт себя как раньше.
//
// Горутина завершается при отмене ctx. Экспортёр не имеет процедуры graceful shutdown,
// поэтому на практике горутина живёт до конца процесса, но ctx нужен для http-запросов
// и делает поведение предсказуемым в тестах.
func (c *SecurityChecker) Start(ctx context.Context) {
	go c.refreshLoop(ctx)
}

// refreshLoop обновляет снимок, пока не отменят ctx.
func (c *SecurityChecker) refreshLoop(ctx context.Context) {
	for {
		// Пауза до следующей попытки зависит от того, удалась ли текущая.
		wait := advisoriesRefreshInterval
		if err := c.refresh(ctx); err != nil {
			c.log.Warnf("Security check: can not refresh advisories: %s", err)
			wait = advisoriesRetryInterval
		}

		timer := time.NewTimer(wait)

		// select ждёт того из двух событий, которое произойдёт раньше: истечения таймера
		// или отмены контекста. Без этого отмена не сработала бы до конца паузы, и
		// процесс не смог бы завершиться быстро.
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return
		}
	}
}

// refresh выкачивает список заново и заменяет снимок.
func (c *SecurityChecker) refresh(ctx context.Context) error {
	advisories, err := c.fetchAll(ctx)
	if err != nil {
		return err
	}

	fixes := parseFixes(advisories)
	if len(fixes) == 0 {
		// Ни одной пригодной записи про Nextcloud Server. Так выглядит, например,
		// переименование ecosystem на стороне GitHub. Старый снимок при этом сохраняется,
		// а если его нет, проверка остаётся отключённой и метрика ведёт себя как раньше.
		// Считать, что уязвимостей нет, здесь нельзя — это молча спрятало бы обновления
		// безопасности.
		return fmt.Errorf("no usable advisories in %d records", len(advisories))
	}

	c.mu.Lock()
	c.snapshot = &advisorySnapshot{fixes: fixes, fetchedAt: time.Now()}
	c.mu.Unlock()

	c.log.Infof("Security check: loaded %d advisories for Nextcloud Server.", len(fixes))

	return nil
}

// Check сообщает, содержит ли обновление current -> available исправления безопасности.
//
// Второе возвращаемое значение — признак того, что ответ получен. Если оно false, ответа
// нет (нет данных, версии не разобрались), и вызывающий обязан сообщить об обновлении,
// как делал до появления проверки.
func (c *SecurityChecker) Check(current, available string) (hasSecurityFixes bool, ok bool) {
	c.mu.RLock()
	snapshot := c.snapshot
	c.mu.RUnlock()

	if snapshot == nil {
		// Первая выкачка ещё не закончилась или все попытки провалились.
		return false, false
	}

	if time.Since(snapshot.fetchedAt) > advisoriesMaxAge {
		c.log.Warnf("Security check: advisories are older than %s, reporting update.", advisoriesMaxAge)
		return false, false
	}

	currentVersion, currentOK := parseVersion(current)
	availableVersion, availableOK := parseVersion(available)
	if !currentOK || !availableOK {
		c.log.Warnf("Security check: can not parse versions %q and %q, reporting update.", current, available)
		return false, false
	}

	for _, fix := range snapshot.fixes {
		if isSecurityRelevant(fix, currentVersion, availableVersion) {
			c.log.Infof("Security check: update %s -> %s contains fix for %s.", current, available, fix.ghsaID)
			return true, true
		}
	}

	c.log.Infof("Security check: update %s -> %s contains no known security fixes.", current, available)

	return false, true
}

// isSecurityRelevant решает, будет ли уязвимость закрыта переходом с current на available.
//
// Запись про уязвимость перечисляет версии, в которых она исправлена, — по одной на
// каждую поддерживаемую ветку, например 32.0.12, 33.0.6 и 34.0.1. Правил два.
//
// Первое: если на нашей ветке исправление уже установлено, запись нас не касается.
// Проверять надо самую большую из версий нашей ветки, а не любую: если запись
// перечисляет 31.0.1 и 31.0.7, а установлена 31.0.3, то по 31.0.1 мы выглядим
// исправленными, хотя нужного 31.0.7 у нас нет.
//
// Второе: иначе уязвимость будет закрыта, если какая-то из исправленных версий больше
// текущей и не больше доступной.
//
// Вместе правила дают верный ответ и при переходе между ветками. Пример: установлена
// 32.0.14, доступна 33.0.6, исправления перечислены как 32.0.12 и 33.0.6. Второе правило
// само по себе сочло бы обновление нужным, потому что 33.0.6 больше 32.0.14. Но первое
// правило видит, что 32.0.12 у нас уже есть, и отбрасывает запись. Без этого любой, кто
// в порядке на старой ветке, получал бы ложное срабатывание на каждом скрейпе.
func isSecurityRelevant(fix advisoryFix, current, available version) bool {
	var highestOnCurrentBranch version
	for _, patched := range fix.patchedVersions {
		if patched.major() != current.major() {
			continue
		}

		if highestOnCurrentBranch == nil || compareVersions(patched, highestOnCurrentBranch) > 0 {
			highestOnCurrentBranch = patched
		}
	}

	// Правило первое.
	if highestOnCurrentBranch != nil && compareVersions(highestOnCurrentBranch, current) <= 0 {
		return false
	}

	// Правило второе.
	for _, patched := range fix.patchedVersions {
		if compareVersions(current, patched) < 0 && compareVersions(patched, available) <= 0 {
			return true
		}
	}

	return false
}

// parseFixes превращает ответ GitHub в список исправлений, пригодный для сравнения версий.
// Отозванные записи и записи без разобранных версий отбрасываются.
func parseFixes(advisories []advisory) []advisoryFix {
	var result []advisoryFix

	for _, item := range advisories {
		if item.WithdrawnAt != nil {
			continue
		}

		var patched []version
		for _, vulnerability := range item.Vulnerabilities {
			if !isNextcloudServer(vulnerability.Package) {
				continue
			}

			patched = append(patched, parsePatchedVersions(vulnerability)...)
		}

		if len(patched) == 0 {
			continue
		}

		result = append(result, advisoryFix{ghsaID: item.GHSAID, patchedVersions: patched})
	}

	return result
}

// nextcloudServerNames — как в этих данных называют сам Nextcloud Server.
//
// Сверяться со списком, а не искать подстроку "Server", важно: в этом же репозитории
// публикуются advisories по остальным продуктам Nextcloud — Talk, Desktop, Deck, Tables, —
// и их номера версий пересекаются с серверными. Свободный поиск затащил бы в сравнение
// чужие версии и дал ложные срабатывания.
var nextcloudServerNames = map[string]bool{
	"server":           true,
	"nextcloud server": true,
	"nextcloud/server": true,
	"nextcloud":        true,
}

// isNextcloudServer отбирает записи про бесплатную редакцию Nextcloud Server.
//
// Платная редакция публикуется отдельно, под ecosystem "nextcloud enterprise" (а в части
// записей — с опечаткой "nextcloud entreprise"), и её версии состоят из четырёх чисел,
// которым не соответствует ни один публичный релиз. Здесь она не нужна, а точное
// сравнение ecosystem с "nextcloud" её и отсекает.
func isNextcloudServer(pkg advisoryPackage) bool {
	if !strings.EqualFold(strings.TrimSpace(pkg.Ecosystem), "nextcloud") {
		return false
	}

	return nextcloudServerNames[strings.ToLower(strings.TrimSpace(pkg.Name))]
}

// parsePatchedVersions достаёт исправленные версии из одной записи о уязвимости.
//
// Поле patched_versions заполняется людьми и приходит в разном виде: "32.0.12, 33.0.6",
// ">=32.0.0, >=33.0.0" (без пробела), " >= 29.0.0, ..." (с пробелом в начале), просто
// "31.0.1". У части записей поля перепутаны местами: в patched_versions лежат
// неравенства, а сами версии — в vulnerable_version_range. Такие записи распознаются по
// тому, что из patched_versions не удалось получить ни одной версии, а из
// vulnerable_version_range получился чистый список чисел. Нормальный
// vulnerable_version_range почти всегда содержит знак неравенства, поэтому на исправных
// записях этот запасной путь не срабатывает.
func parsePatchedVersions(vulnerability advisoryVulnerability) []version {
	if versions := parseVersionList(vulnerability.PatchedVersions); len(versions) > 0 {
		return versions
	}

	return parseVersionList(vulnerability.VulnerableVersionRange)
}

// parseVersionList разбирает список версий через запятую, отбрасывая всё, что версией не
// является: неравенства вида ">= 32.0.0", пустые элементы, произвольный текст.
func parseVersionList(raw string) []version {
	var result []version

	for _, token := range strings.Split(raw, ",") {
		token = strings.TrimSpace(token)

		// ContainsAny проверяет наличие любого из перечисленных символов. Так отсеиваются
		// границы диапазонов, которые версиями не являются.
		if token == "" || strings.ContainsAny(token, "<>=~^") {
			continue
		}

		if parsed, ok := parseVersion(token); ok {
			result = append(result, parsed)
		}
	}

	return result
}

// parseVersion разбирает "34.0.2.1" в [34, 0, 2, 1].
// Второе возвращаемое значение равно false, если строка версией не является.
func parseVersion(raw string) (version, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false
	}

	parts := strings.Split(raw, ".")
	result := make(version, 0, len(parts))
	for _, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return nil, false
		}

		result = append(result, number)
	}

	return result, true
}

// major возвращает номер ветки — первый компонент версии.
func (v version) major() int {
	if len(v) == 0 {
		return -1
	}

	return v[0]
}

// compareVersions сравнивает версии покомпонентно и возвращает -1, 0 или 1.
//
// Отсутствующие компоненты считаются нулём, поэтому 34.0.1 и 34.0.1.0 равны, а 34.0.2.1
// больше 34.0.2. Это и позволяет сравнивать четырёхкомпонентную версию Nextcloud с
// трёхкомпонентной из advisories.
func compareVersions(a, b version) int {
	length := len(a)
	if len(b) > length {
		length = len(b)
	}

	for i := 0; i < length; i++ {
		left, right := 0, 0
		if i < len(a) {
			left = a[i]
		}
		if i < len(b) {
			right = b[i]
		}

		if left < right {
			return -1
		}
		if left > right {
			return 1
		}
	}

	return 0
}

// fetchAll выкачивает все страницы списка advisories.
//
// Ошибка на любой странице прекращает обход и не возвращает ничего: неполный список
// выглядел бы как отсутствие уязвимостей и мог бы скрыть обновление безопасности.
func (c *SecurityChecker) fetchAll(ctx context.Context) ([]advisory, error) {
	var result []advisory

	next := c.url
	for page := 0; next != ""; page++ {
		if page >= advisoriesMaxPages {
			return nil, fmt.Errorf("too many pages, stopped after %d", advisoriesMaxPages)
		}

		advisories, link, err := c.fetchPage(ctx, next)
		if err != nil {
			return nil, err
		}

		result = append(result, advisories...)
		next = link
	}

	return result, nil
}

// fetchPage запрашивает одну страницу и возвращает её содержимое вместе со ссылкой на
// следующую страницу (пустая строка означает, что страница последняя).
func (c *SecurityChecker) fetchPage(ctx context.Context, pageURL string) ([]advisory, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", c.userAgent)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	// defer выполнится при выходе из функции по любой причине. Тело ответа обязательно
	// закрывать, иначе соединение не вернётся в пул и утечёт.
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		// Исчерпание лимита GitHub выглядит как 403 или 429. Отдельной обработки не
		// требует: следующая попытка будет через advisoriesRetryInterval, а остаток
		// лимита попадает в сообщение, чтобы причину было видно в логе.
		return nil, "", fmt.Errorf("unexpected status code %d (rate limit remaining: %q)",
			res.StatusCode, res.Header.Get("X-RateLimit-Remaining"))
	}

	var advisories []advisory
	if err := json.NewDecoder(res.Body).Decode(&advisories); err != nil {
		return nil, "", fmt.Errorf("can not parse advisories: %w", err)
	}

	return advisories, nextPageURL(pageURL, res.Header.Get("Link")), nil
}

// nextPageURL достаёт адрес следующей страницы из заголовка Link, который выглядит так:
//
//	<https://api.github.com/repositories/63675426/security-advisories?after=...>; rel="next"
//
// Сверяются только схема и хост. Путь сверять нельзя: GitHub отдаёт следующую страницу по
// другому пути, с числовым идентификатором репозитория вместо его имени. Проверка хоста
// нужна, чтобы промежуточный прокси не смог увести выкачку на посторонний адрес.
func nextPageURL(currentURL, linkHeader string) string {
	current, err := url.Parse(currentURL)
	if err != nil {
		return ""
	}

	for _, link := range strings.Split(linkHeader, ",") {
		parts := strings.Split(strings.TrimSpace(link), ";")
		if len(parts) < 2 {
			continue
		}

		isNext := false
		for _, part := range parts[1:] {
			if strings.EqualFold(strings.TrimSpace(part), `rel="next"`) {
				isNext = true
				break
			}
		}
		if !isNext {
			continue
		}

		address := strings.TrimSpace(parts[0])
		if !strings.HasPrefix(address, "<") || !strings.HasSuffix(address, ">") {
			continue
		}

		next, err := url.Parse(strings.Trim(address, "<>"))
		if err != nil || next.Scheme != current.Scheme || next.Host != current.Host {
			return ""
		}

		return next.String()
	}

	return ""
}
