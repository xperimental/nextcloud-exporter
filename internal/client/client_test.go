package client

import (
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/xperimental/nextcloud-exporter/internal/testutil"
	"github.com/xperimental/nextcloud-exporter/serverinfo"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	wantUserAgent := "test-ua"
	wantUsername := "test-user"
	wantPassword := "test-password"
	wantToken := "test-token"

	tt := []struct {
		desc     string
		password string
		token    string
		handler  func(t *testing.T) http.Handler
		wantInfo *serverinfo.ServerInfo
		wantErr  error
	}{
		{
			desc:     "password",
			password: wantPassword,
			token:    "",
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					user, password, ok := req.BasicAuth()
					if !ok {
						t.Error("failed to get authentication header")
					}

					if user != wantUsername {
						t.Errorf("got username %q, want %q", user, wantUsername)
					}

					if password != wantPassword {
						t.Errorf("got password %q, want %q", password, wantPassword)
					}

					fmt.Fprintln(w, "{}")
				})
			},
			wantInfo: &serverinfo.ServerInfo{},
			wantErr:  nil,
		},
		{
			desc:     "token",
			password: "",
			token:    wantToken,
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					token := req.Header.Get(nextcloudTokenHeader)
					if token != wantToken {
						t.Errorf("got token %q, want %q", token, wantToken)
					}

					fmt.Fprintln(w, "{}")
				})
			},
			wantInfo: &serverinfo.ServerInfo{},
			wantErr:  nil,
		},
		{
			desc:  "user-agent",
			token: wantToken,
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					ua := req.UserAgent()
					if ua != wantUserAgent {
						t.Errorf("got user-agent %q, want %q", ua, wantUserAgent)
					}

					fmt.Fprintln(w, "{}")
				})
			},
			wantInfo: &serverinfo.ServerInfo{},
			wantErr:  nil,
		},
		{
			desc:     "auth error",
			password: "",
			token:    "",
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
				})
			},
			wantInfo: nil,
			wantErr:  ErrNotAuthorized,
		},
		{
			desc:     "parse error",
			password: "",
			token:    "",
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
			},
			wantInfo: nil,
			wantErr:  errors.New("can not parse server info: EOF"),
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			s := httptest.NewServer(tc.handler(t))
			defer s.Close()

			client := New(s.URL, wantUsername, tc.password, tc.token, time.Second, wantUserAgent, false)

			info, err := client()

			if !testutil.EqualErrorMessage(err, tc.wantErr) {
				t.Errorf("got error %q, want %q", err, tc.wantErr)
			}

			if err != nil {
				return
			}

			if diff := cmp.Diff(info, tc.wantInfo); diff != "" {
				t.Errorf("info differs: -got+want\n%s", diff)
			}
		})
	}
}
