package login

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/xperimental/nextcloud-exporter/internal/testutil"
)

func testClient(url string) *Client {
	return &Client{
		log:       logrus.New(),
		client:    &http.Client{},
		serverURL: url,
		sleepFunc: func() {},
	}
}

func testHandler(status int, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprintln(w, body)
	})
}

func TestGetMajorVersion(t *testing.T) {
	tt := []struct {
		desc        string
		testHandler http.Handler
		wantErr     error
		wantVersion int
	}{
		{
			desc:        "success",
			testHandler: testHandler(http.StatusOK, `{"version": "18.0.2"}`),
			wantVersion: 18,
		},
		{
			desc:        "parse error",
			testHandler: testHandler(http.StatusOK, ``),
			wantErr:     errors.New("error decoding status: EOF"),
		},
		{
			desc:        "error version",
			testHandler: testHandler(http.StatusOK, `{"version": "unparseable"}`),
			wantErr:     errors.New(`can not parse "unparseable" as version: strconv.Atoi: parsing "unparseable": invalid syntax`),
		},
		{
			desc:        "error http",
			testHandler: testHandler(http.StatusInternalServerError, "test error"),
			wantErr:     errors.New("non-ok status: 500"),
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			s := httptest.NewServer(tc.testHandler)
			defer s.Close()
			c := testClient(s.URL)

			version, err := c.getMajorVersion()

			if !testutil.EqualErrorMessage(err, tc.wantErr) {
				t.Errorf("got error %q, want %q", err, tc.wantErr)
			}

			if err != nil {
				return
			}

			if version != tc.wantVersion {
				t.Errorf("got version %d, want %d", version, tc.wantVersion)
			}
		})
	}
}

func TestGetLoginInfo(t *testing.T) {
	tt := []struct {
		desc        string
		testHandler http.Handler
		wantErr     error
		wantInfo    loginInfo
	}{
		{
			desc:        "success",
			testHandler: testHandler(http.StatusOK, `{"login": "http://localhost/login", "poll": {"token": "token", "endpoint": "http://localhost/poll"}}`),
			wantInfo: loginInfo{
				LoginURL: "http://localhost/login",
				PollInfo: pollInfo{
					Token:    "token",
					Endpoint: "http://localhost/poll",
				},
			},
		},
		{
			desc:        "parse error",
			testHandler: testHandler(http.StatusOK, ``),
			wantErr:     errors.New("error decoding login info: EOF"),
		},
		{
			desc:        "error http",
			testHandler: testHandler(http.StatusInternalServerError, "test error"),
			wantErr:     errors.New("non-ok status: 500"),
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			s := httptest.NewServer(tc.testHandler)
			defer s.Close()
			c := testClient(s.URL)

			info, err := c.getLoginInfo()

			if !testutil.EqualErrorMessage(err, tc.wantErr) {
				t.Errorf("got error %q, want %q", err, tc.wantErr)
			}

			if err != nil {
				return
			}

			if diff := cmp.Diff(info, tc.wantInfo); diff != "" {
				t.Errorf("info differs: -got +want\n%s", diff)
			}
		})
	}
}

func TestPollPassword(t *testing.T) {
	tt := []struct {
		desc        string
		testHandler http.Handler
		pollInfo    pollInfo
		wantErr     error
		wantLogin   Login
	}{
		{
			desc:        "success",
			testHandler: testHandler(http.StatusOK, `{"loginName": "username", "appPassword": "password"}`),
			wantLogin: Login{
				Username: "username",
				Password: "password",
			},
		},
		{
			desc:        "parse error",
			testHandler: testHandler(http.StatusOK, ``),
			wantErr:     errors.New("error decoding password info: EOF"),
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			s := httptest.NewServer(tc.testHandler)
			defer s.Close()
			tc.pollInfo.Endpoint = s.URL

			c := testClient("")
			login, err := c.pollLogin(tc.pollInfo)

			if !testutil.EqualErrorMessage(err, tc.wantErr) {
				t.Errorf("got error %q, want %q", err, tc.wantErr)
			}

			if err != nil {
				return
			}

			if diff := cmp.Diff(login, tc.wantLogin); diff != "" {
				t.Errorf("login differs: -got +want\n%s", diff)
			}
		})
	}
}
