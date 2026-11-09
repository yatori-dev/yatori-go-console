package runner

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	yatori "github.com/yatori-dev/yatori-go-core/v3"
)

func TestPromptLoginSMS(t *testing.T) {
	for _, tc := range []struct {
		name, input, wantCode, wantError string
		wantSends                        int
		sendErr                          error
	}{
		{name: "confirm", input: "y\r\n123456\r\n", wantCode: "123456", wantSends: 1},
		{name: "decline", input: "n\n", wantError: "已取消发送短信验证码"},
		{name: "failure", input: "y\n", wantError: "发送短信验证码失败", wantSends: 1, sendErr: errors.New("mock failure")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer r.Close()
			saved := os.Stdin
			os.Stdin = r
			defer func() { os.Stdin = saved }()
			if _, err := w.WriteString(tc.input); err != nil {
				w.Close()
				t.Fatal(err)
			}
			w.Close()

			sends := 0
			code, err := promptLoginCode(t.Context(), yatori.LoginCodeRequest{
				Kind: "sms",
				SendSMS: func(context.Context) error {
					sends++
					return tc.sendErr
				},
			})
			if sends != tc.wantSends || code != tc.wantCode {
				t.Fatalf("got code %q and %d sends", code, sends)
			}
			if tc.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("got error %v, want %q", err, tc.wantError)
			}
		})
	}
}

func TestAndroidLoginMode(t *testing.T) {
	for _, mode := range []string{"Android", " android ", "app"} {
		if !usesAppLogin(mode) {
			t.Errorf("mode %q did not select App login", mode)
		}
	}
	if usesAppLogin("web") || usesAppLogin("") {
		t.Error("web and default modes must retain existing behavior")
	}
}
