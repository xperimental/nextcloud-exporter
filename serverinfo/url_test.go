package serverinfo

import (
	"testing"
)

func TestInfoURL(t *testing.T) {
	tt := []struct {
		desc       string
		serverURL  string
		skipApps   bool
		skipUpdate bool
		wantURL    string
	}{
		{
			desc:      "do not skip apps",
			serverURL: "https://nextcloud.example.com",
			wantURL:   "https://nextcloud.example.com/ocs/v2.php/apps/serverinfo/api/v1/info?format=json&skipApps=false",
		},
		{
			desc:      "skip apps",
			serverURL: "https://nextcloud.example.com",
			skipApps:  true,
			wantURL:   "https://nextcloud.example.com/ocs/v2.php/apps/serverinfo/api/v1/info?format=json&skipApps=true",
		},
		{
			desc:       "do not skip update",
			serverURL:  "https://nextcloud.example.com",
			skipUpdate: false,
			wantURL:    "https://nextcloud.example.com/ocs/v2.php/apps/serverinfo/api/v1/info?format=json&skipUpdate=false",
		},
		{
			desc:       "skip update",
			serverURL:  "https://nextcloud.example.com",
			skipUpdate: true,
			wantURL:    "https://nextcloud.example.com/ocs/v2.php/apps/serverinfo/api/v1/info?format=json&skipUpdate=true",
		},
		{
			desc:       "do not skip update and do not skip apps",
			serverURL:  "https://nextcloud.example.com",
			skipApps:   false,
			skipUpdate: false,
			wantURL:    "https://nextcloud.example.com/ocs/v2.php/apps/serverinfo/api/v1/info?format=json&skipUpdate=false&skipApps=false",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			url := InfoURL(tc.serverURL, tc.skipApps, tc.skipUpdate)
			if url != tc.wantURL {
				t.Errorf("got url %q, want %q", url, tc.wantURL)
			}
		})
	}
}
