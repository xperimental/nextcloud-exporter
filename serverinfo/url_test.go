package serverinfo

import (
	"testing"
)

func TestInfoURL(t *testing.T) {
	tt := []struct {
		desc      string
		serverURL string
		skipApps  bool
		wantURL   string
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
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			url := InfoURL(tc.serverURL, tc.skipApps)
			if url != tc.wantURL {
				t.Errorf("got url %q, want %q", url, tc.wantURL)
			}
		})
	}
}
