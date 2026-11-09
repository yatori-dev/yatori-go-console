package runner

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
)

func TestReadLoginCode(t *testing.T) {
	sendErr := errors.New("SMS provider unavailable")
	tests := []struct {
		name      string
		input     string
		sms       bool
		sendErr   error
		wantCause error
		wantCode  string
		wantError string
		wantSends int
	}{
		{name: "confirm sends once and reads code", input: "Y\r\n123456\r\n", sms: true, wantCode: "123456", wantSends: 1},
		{name: "cancel does not send", input: "\n", sms: true, wantError: "已取消发送短信验证码"},
		{name: "EOF before confirmation does not send", input: "y", sms: true, wantError: "EOF"},
		{name: "send failure is returned", input: "y\n", sms: true, sendErr: sendErr, wantCause: sendErr, wantError: "发送短信验证码失败: SMS provider unavailable", wantSends: 1},
		{
			name: "transport error omits request URL", input: "y\n", sms: true,
			sendErr:   &url.Error{Op: "Get", URL: "https://example.test/sms?phone=fixture&code=fixture", Err: sendErr},
			wantCause: sendErr, wantError: "发送短信验证码失败: SMS provider unavailable", wantSends: 1,
		},
		{name: "no SMS callback reads code directly", input: "123456\n", wantCode: "123456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			sends := 0
			var sendSMS func(context.Context) error
			if tt.sms {
				sendSMS = func(gotCtx context.Context) error {
					if gotCtx != ctx {
						t.Error("SendSMS did not receive the login context")
					}
					sends++
					return tt.sendErr
				}
			}

			code, err := readLoginCode(ctx, strings.NewReader(tt.input), sendSMS)
			if sends != tt.wantSends || code != tt.wantCode {
				t.Fatalf("got code %q and %d sends, want code %q and %d sends", code, sends, tt.wantCode, tt.wantSends)
			}
			if tt.wantError == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if err == nil || err.Error() != tt.wantError {
				t.Fatalf("got error %v, want %q", err, tt.wantError)
			}
			if tt.wantCause != nil && !errors.Is(err, tt.wantCause) {
				t.Fatal("send error was not wrapped")
			}
		})
	}
}
