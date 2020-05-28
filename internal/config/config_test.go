package config

import (
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func testEnv(env map[string]string) func(string) string {
	return func(key string) string {
		return env[key]
	}
}

func mustURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}

	return u
}

var compareErrors = cmp.Comparer(func(a, b error) bool {
	aE := a.Error()
	bE := b.Error()

	return aE == bE
})

func TestConfig(t *testing.T) {
	defaults := defaultConfig()
	tt := []struct {
		desc       string
		args       []string
		env        map[string]string
		wantErr    error
		wantConfig Config
	}{
		{
			desc: "flags",
			args: []string{
				"test",
				"--addr",
				"127.0.0.1:9205",
				"--timeout",
				"30s",
				"--url",
				"http://localhost",
				"--username",
				"testuser",
				"--password",
				"testpass",
			},
			env:     map[string]string{},
			wantErr: nil,
			wantConfig: Config{
				ListenAddr: "127.0.0.1:9205",
				Timeout:    30 * time.Second,
				InfoURL:    mustURL("http://localhost/ocs/v2.php/apps/serverinfo/api/v1/info"),
				Username:   "testuser",
				Password:   "testpass",
			},
		},
		{
			desc: "password from file",
			args: []string{
				"test",
				"--url",
				"http://localhost",
				"--username",
				"testuser",
				"--password",
				"@testdata/password",
			},
			env:     map[string]string{},
			wantErr: nil,
			wantConfig: Config{
				ListenAddr: defaults.ListenAddr,
				Timeout:    defaults.Timeout,
				InfoURL:    mustURL("http://localhost/ocs/v2.php/apps/serverinfo/api/v1/info"),
				Username:   "testuser",
				Password:   "testpass",
			},
		},
		{
			desc: "config from file",
			args: []string{
				"test",
				"--config-file",
				"testdata/all.yml",
			},
			env:     map[string]string{},
			wantErr: nil,
			wantConfig: Config{
				ListenAddr: "127.0.0.10:9205",
				Timeout:    10 * time.Second,
				InfoURL:    mustURL("http://localhost/ocs/v2.php/apps/serverinfo/api/v1/info"),
				Username:   "testuser",
				Password:   "testpass",
			},
		},
		{
			desc: "env config",
			args: []string{
				"test",
			},
			env: map[string]string{
				envListenAddress: "127.0.0.11:9205",
				envTimeout:       "15s",
				envInfoURL:       "http://localhost",
				envUsername:      "testuser",
				envPassword:      "testpass",
			},
			wantErr: nil,
			wantConfig: Config{
				ListenAddr: "127.0.0.11:9205",
				Timeout:    15 * time.Second,
				InfoURL:    mustURL("http://localhost/ocs/v2.php/apps/serverinfo/api/v1/info"),
				Username:   "testuser",
				Password:   "testpass",
			},
		},
		{
			desc: "minimal env",
			args: []string{
				"test",
			},
			env: map[string]string{
				envInfoURL:  "http://localhost",
				envUsername: "testuser",
				envPassword: "testpass",
			},
			wantErr: nil,
			wantConfig: Config{
				ListenAddr: defaults.ListenAddr,
				Timeout:    defaults.Timeout,
				InfoURL:    mustURL("http://localhost/ocs/v2.php/apps/serverinfo/api/v1/info"),
				Username:   "testuser",
				Password:   "testpass",
			},
		},
		{
			desc: "wrongflag",
			args: []string{
				"test",
				"--unknown",
			},
			env:     map[string]string{},
			wantErr: errors.New("error parsing flags: unknown flag: --unknown"),
		},
		{
			desc: "no url",
			args: []string{
				"test",
			},
			env:     map[string]string{},
			wantErr: errors.New("need to set an info URL"),
		},
		{
			desc: "no username",
			args: []string{
				"test",
				"--url",
				"http://localhost",
				"--password",
				"testpass",
			},
			env:     map[string]string{},
			wantErr: errors.New("need to provide a username"),
		},
		{
			desc: "no password",
			args: []string{
				"test",
				"--url",
				"http://localhost",
				"--username",
				"testuser",
			},
			env:     map[string]string{},
			wantErr: errors.New("need to provide a password"),
		},
		{
			desc: "env wrong duration",
			args: []string{
				"test",
			},
			env: map[string]string{
				"NEXTCLOUD_TIMEOUT": "unknown",
			},
			wantErr: errors.New("error reading environment variables: time: invalid duration unknown"),
		},
		{
			desc: "password from file error",
			args: []string{
				"test",
				"--url",
				"http://localhost",
				"--password",
				"@testdata/notfound",
			},
			env:     map[string]string{},
			wantErr: errors.New("can not read password file: open testdata/notfound: no such file or directory"),
		},
		{
			desc: "config from file error",
			args: []string{
				"test",
				"--config-file",
				"testdata/notfound.yml",
			},
			env:     map[string]string{},
			wantErr: errors.New("error reading configuration file: open testdata/notfound.yml: no such file or directory"),
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			config, err := parseConfig(tc.args, testEnv(tc.env))

			if diff := cmp.Diff(err, tc.wantErr, compareErrors); diff != "" {
				t.Errorf("error differs: -got +want\n%s", diff)
			}

			if err != nil {
				return
			}

			if diff := cmp.Diff(config, tc.wantConfig); diff != "" {
				t.Errorf("config differs: -got +want\n%s", diff)
			}
		})
	}
}
