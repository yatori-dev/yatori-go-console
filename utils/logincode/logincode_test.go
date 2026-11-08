package logincode

import (
	"context"
	"testing"
	"time"
)

func TestWaitAndSubmit(t *testing.T) {
	key := Key("XUEXITONG", "logincode-test")
	defer Clear(key)

	done := make(chan struct{})
	go func() {
		defer close(done)
		// 等挂起后再提交
		for i := 0; i < 100; i++ {
			if Pending(key) != nil {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if pending := Pending(key); pending == nil || pending.Kind != "sms" || pending.Phone != "13800000000" {
			t.Errorf("unexpected pending: %+v", pending)
		}
		if err := Submit(key, "1234"); err != nil {
			t.Errorf("submit failed: %v", err)
		}
	}()

	code, err := Wait(context.Background(), key, Challenge{Kind: "sms", Prompt: "验证码", Phone: "13800000000"}, time.Second)
	if err != nil {
		t.Fatalf("wait failed: %v", err)
	}
	if code != "1234" {
		t.Fatalf("want code 1234, got %s", code)
	}
	<-done
	if Pending(key) != nil {
		t.Fatal("pending should be cleared after wait returns")
	}
}

func TestSubmitWithoutPending(t *testing.T) {
	if err := Submit("XUEXITONG:none", "1"); err == nil {
		t.Fatal("expected error when no pending challenge")
	}
}
