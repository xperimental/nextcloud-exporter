package serverinfo

import (
	"fmt"
)

const (
	infoPathFormat = "%s/ocs/v2.php/apps/serverinfo/api/v1/info?format=json&skipApps=%v"
)

// InfoURL constructs the URL of the info endpoint from the server base URL and optional parameters.
func InfoURL(serverURL string, skipApps bool) string {
	return fmt.Sprintf(infoPathFormat, serverURL, skipApps)
}
