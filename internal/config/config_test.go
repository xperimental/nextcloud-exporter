package config

import (
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/xperimental/nextcloud-exporter/internal/testutil"
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
				"--server",
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
				ServerURL:  "http://localhost",
				Username:   "testuser",
				Password:   "testpass",
			},
		},
		{
			desc: "password from file",
			args: []string{
				"test",
				"--server",
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
				ServerURL:  "http://localhost",
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
				ServerURL:  "http://localhost",
				Username:   "testuser",
				Password:   "testpass",
			},
		},
		{
			desc: "config and password from file",
			args: []string{
				"test",
				"--config-file",
				"testdata/passwordfile.yml",
			},
			env:     map[string]string{},
			wantErr: nil,
			wantConfig: Config{
				ListenAddr: "127.0.0.10:9205",
				Timeout:    10 * time.Second,
				ServerURL:  "http://localhost",
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
				envServerURL:     "http://localhost",
				envUsername:      "testuser",
				envPassword:      "testpass",
			},
			wantErr: nil,
			wantConfig: Config{
				ListenAddr: "127.0.0.11:9205",
				Timeout:    15 * time.Second,
				ServerURL:  "http://localhost",
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
				envServerURL: "http://localhost",
				envUsername:  "testuser",
				envPassword:  "testpass",
			},
			wantErr: nil,
			wantConfig: Config{
				ListenAddr: defaults.ListenAddr,
				Timeout:    defaults.Timeout,
				ServerURL:  "http://localhost",
				Username:   "testuser",
				Password:   "testpass",
			},
		},
		{
			desc: "show help",
			args: []string{
				"test",
				"--help",
			},
			env:     map[string]string{},
			wantErr: nil,
			wantConfig: Config{
				RunMode: RunModeHelp,
			},
		},
		{
			desc: "login mode",
			args: []string{
				"test",
				"--login",
				"--server",
				"http://localhost",
			},
			env:     map[string]string{},
			wantErr: nil,
			wantConfig: Config{
				ListenAddr: defaults.ListenAddr,
				Timeout:    defaults.Timeout,
				ServerURL:  "http://localhost",
				RunMode:    RunModeLogin,
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
			desc: "env wrong duration",
			args: []string{
				"test",
			},
			env: map[string]string{
				"NEXTCLOUD_TIMEOUT": "unknown",
			},
			wantErr: errors.New("error reading environment variables: time: invalid duration \"unknown\""),
		},
		{
			desc: "password from file error",
			args: []string{
				"test",
				"--server",
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

			if diff := cmp.Diff(err, tc.wantErr, testutil.ErrorComparer); diff != "" {
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

func TestConfigValidate(t *testing.T) {
	tt := []struct {
		desc    string
		config  Config
		wantErr error
	}{
		{
			desc: "minimal",
			config: Config{
				ServerURL: "https://example.com",
				Username:  "exporter",
				Password:  "testpass",
			},
			wantErr: nil,
		},
		{
			desc: "no url",
			config: Config{
				Username: "exporter",
				Password: "testpass",
			},
			wantErr: errors.New("need to set a server URL"),
		},
		{
			desc: "no username",
			config: Config{
				ServerURL: "https://example.com",
				Password:  "testpass",
			},
			wantErr: errors.New("need to provide a username"),
		},
		{
			desc: "no password",
			config: Config{
				ServerURL: "https://example.com",
				Username:  "exporter",
			},
			wantErr: errors.New("need to provide a password"),
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()

			if diff := cmp.Diff(err, tc.wantErr, testutil.ErrorComparer); diff != "" {
				t.Errorf("error differs: -got +want\n%s", diff)
			}
		})
	}
}
