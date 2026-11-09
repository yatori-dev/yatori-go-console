package runner

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"yatori-go-console/config"

	yatori "github.com/yatori-dev/yatori-go-core/v3"
)

func TestPromptLoginCode(t *testing.T) {
	sendFailure := errors.New("synthetic send failure")
	privateURL := "https://example.invalid/send?phone=synthetic&captcha=synthetic"
	for _, tc := range []struct {
		name      string
		input     string
		hasSend   bool
		sendErr   error
		wantCalls int
		wantCode  string
		wantErr   bool
		wantCause error
	}{
		{name: "confirm_crlf", input: "1\r\n123456\r\n", hasSend: true, wantCalls: 1, wantCode: "123456"},
		{name: "confirm_lf", input: " 1 \n 123456 \n", hasSend: true, wantCalls: 1, wantCode: "123456"},
		{name: "cancel", input: "\r\n", hasSend: true, wantErr: true},
		{name: "decline", input: "n\n", hasSend: true, wantErr: true},
		{name: "eof_before_confirm", hasSend: true, wantErr: true, wantCause: io.EOF},
		{name: "incomplete_confirm", input: "1", hasSend: true, wantErr: true, wantCause: io.EOF},
		{name: "send_failure", input: "1\n", hasSend: true, sendErr: sendFailure, wantCalls: 1, wantErr: true, wantCause: sendFailure},
		{name: "url_error", input: "1\n", hasSend: true, sendErr: &url.Error{Op: "Get", URL: privateURL, Err: sendFailure}, wantCalls: 1, wantErr: true, wantCause: sendFailure},
		{name: "no_send_callback", input: "123456\r\n", wantCode: "123456"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// 标准输入属于整个进程，这些用例按顺序执行。
			path := filepath.Join(t.TempDir(), "stdin.txt")
			if err := os.WriteFile(path, []byte(tc.input), 0600); err != nil {
				t.Fatal(err)
			}
			input, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			original := os.Stdin
			os.Stdin = input
			t.Cleanup(func() {
				os.Stdin = original
				input.Close()
			})

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			req := yatori.LoginCodeRequest{Kind: "sms", Prompt: "Synthetic SMS challenge"}
			calls := 0
			if tc.hasSend {
				req.SendSMS = func(sendCtx context.Context) error {
					calls++
					if sendCtx != ctx {
						t.Error("SendSMS received a different context")
					}
					return tc.sendErr
				}
			}

			// 通过实际配置入口取得 CLI 回调，全部数据均为合成数据。
			options := buildOptions(config.Setting{}, config.User{
				AccountType: "XUEXITONG",
				LoginMode:   "app",
			})
			code, err := options.LoginCodeProvider(ctx, req)
			if calls != tc.wantCalls || code != tc.wantCode || (err != nil) != tc.wantErr {
				t.Errorf("got calls=%d, code=%q, err=%v; want calls=%d, code=%q, error=%t",
					calls, code, err, tc.wantCalls, tc.wantCode, tc.wantErr)
			}
			if tc.wantCause != nil && !errors.Is(err, tc.wantCause) {
				t.Errorf("got error %v, want cause %v", err, tc.wantCause)
			}
			if err != nil && strings.Contains(err.Error(), privateURL) {
				t.Errorf("error contains request URL: %v", err)
			}
		})
	}
}
