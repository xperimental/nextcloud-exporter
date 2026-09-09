# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make                      # test + build-binary (the default target)
make test                 # go test -cover ./...
make build-binary         # produces ./nextcloud-exporter with version/commit ldflags
make lint                 # golangci-lint run + promtool check rules
make lint-go-fix          # golangci-lint run --fix (includes gofumpt formatting)
make tools                # build the pinned bingo tools into $GOBIN
make image                # local docker buildx image (DOCKER_REPO/DOCKER_TAG overridable)
make deb                  # debian package via fakeroot dpkg-deb

go test ./internal/config/ -run TestConfig                 # single test
go test ./serverinfo/ -run 'TestParseJSON/nc22.json' -v     # single table-test case
```

Linters are pinned via [bingo](https://github.com/bwplotka/bingo) in [.bingo/](.bingo/) and installed into
`$GOBIN` as version-suffixed binaries (e.g. `golangci-lint-v2.11.4`). Always invoke them through `make`, not
a system-installed `golangci-lint` — CI pins the same versions. `gofumpt` (not plain `gofmt`) is the
formatter, enforced by golangci-lint.

`make build-binary` runs `git describe --tags --dirty`, so builds require a git checkout with tags.

## Architecture

A single-binary Prometheus exporter that scrapes the Nextcloud `serverinfo` OCS API on demand and translates
one JSON document into metrics. There is no background polling or persistent state: each Prometheus scrape
triggers exactly one HTTP request to Nextcloud.

The one exception is the opt-in security-update check in
[internal/metrics/security.go](internal/metrics/security.go), gated behind `--enable-info-update-security`
and off by default. A `SecurityChecker` refreshes the advisory list from
`api.github.com/repos/nextcloud/security-advisories/security-advisories` in a background goroutine every 12h
and publishes an immutable snapshot under an `RWMutex`, so a scrape never waits on GitHub. It is passed to
`RegisterCollector` as a nilable argument; nil means the check is off and `collectUpdate` behaves exactly as
before. Every uncertainty (no snapshot yet, snapshot older than 48h, unparseable versions, fetch failure)
reports the update, so the metric only ever drops to 0 when the check positively knows there are no security
fixes. Advisory failures deliberately stay out of `nextcloud_scrape_errors_total` and never touch
`nextcloud_up` — they are logged only. A line-by-line walkthrough of this code, in Russian, is in
[docs/security-check.ru.md](docs/security-check.ru.md).

Request flow: [main.go](main.go) builds the info URL → `internal/client` fetches and classifies HTTP status →
`serverinfo.ParseJSON` decodes → `internal/metrics` emits metrics on the collector channel.

- **[serverinfo/](serverinfo/)** — the only non-`internal` package, importable by third parties. It owns the
  API response schema and URL construction. The Nextcloud API is loosely typed, so several structs carry
  custom unmarshallers: `System` decodes `"yes"`/`"no"` strings into bools (and has *both* JSON and XML
  variants that must be kept in sync), and `Database.Size` accepts either a number or a string. Test fixtures
  in [serverinfo/testdata/](serverinfo/testdata/) capture real-world response variants (`nc22.json`,
  `na-values.json`, `negative-space.json`, `large-freespace.json`) — add a fixture when handling a new
  server quirk rather than hand-building structs.
- **[internal/client/](internal/client/)** — `client.InfoClient` is a *function type* (`func() (*ServerInfo,
  error)`), not an interface; tests inject a closure. It maps HTTP responses to sentinel errors
  (`ErrNotAuthorized`, `ErrRatelimit`, `ErrUnavailable`, `ErrMaintenanceMode`) which the collector turns into
  `cause` label values on `nextcloud_scrape_errors_total`. Adding a new failure mode means touching the
  sentinel, the `cause` constant in [internal/metrics/collector.go](internal/metrics/collector.go), the README
  metric table, and usually [contrib/prometheus-alerts.yaml](contrib/prometheus-alerts.yaml).
- **[internal/metrics/](internal/metrics/)** — a `prometheus.Collector`. All metric names go through the
  `metricPrefix` const (`nextcloud_`) and are declared as package-level `prometheus.NewDesc` vars emitted as
  const metrics. `Collect` never fails a scrape: on error it increments the error counter, sets
  `nextcloud_up` to 0, and still returns. The apps and update metric groups are gated behind config flags
  because they make Nextcloud do extra work.
- **[internal/config/](internal/config/)** — layered config, merged by `mergeConfig`.
- **[internal/login/](internal/login/)** — the `--login` flow (Login Flow v2, needs Nextcloud ≥ 16); a
  separate `RunMode`, unrelated to the exporter path.

### Config precedence (non-obvious)

`parseConfig` merges in the order flags → config file → environment, and `mergeConfig` overrides only with
*non-zero* values. So **environment variables win over the config file, which wins over command-line flags**
— note the README's "Configuration methods" list is ordered confusingly relative to this. Because zero
values never override, a `true` boolean from a lower layer cannot be turned off by a higher one. A `password`
or `authToken` value starting with `@` is treated as a path to a file holding the secret, resolved after all
merging.

## Conventions

Tests are table-driven with a `tt := []struct{ desc string; ...}` slice, `t.Run(tc.desc, ...)`, `go-cmp` for
value comparison, and `testutil.EqualErrorMessage` for error comparison.

User-visible changes go in the `## [Unreleased]` section of [CHANGELOG.md](CHANGELOG.md) (Keep a Changelog
format, semver). New or changed metrics must also be reflected in the README metric table, and where relevant
in [contrib/grafana-dashboard.json](contrib/grafana-dashboard.json) and
[contrib/prometheus-alerts.yaml](contrib/prometheus-alerts.yaml) (the latter is validated by `make
lint-promtool`).
