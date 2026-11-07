package runner

import (
	"errors"
	"testing"
)

// core 里 YatoriCoreInit 与 driverutil.EnsureOCR 各有一个 sync.Once，
// 而 onnxruntime 只能初始化一次，第二次会 panic。这里验证 panic 会被转成
// errOCRConflict，供 login 重试。
func TestCallWithOCRRecoverMapsPanic(t *testing.T) {
	called := 0
	err := callWithOCRRecover(func() error {
		called++
		panic(errors.New("The onnxruntime has already been initialized"))
	})
	if !errors.Is(err, errOCRConflict) {
		t.Fatalf("want errOCRConflict, got %v", err)
	}
	if called != 1 {
		t.Fatalf("want 1 call, got %d", called)
	}
}

func TestCallWithOCRRecoverPassesThrough(t *testing.T) {
	sentinel := errors.New("login failed")
	err := callWithOCRRecover(func() error { return sentinel })
	if !errors.Is(err, sentinel) {
		t.Fatalf("want sentinel error, got %v", err)
	}

	if err := callWithOCRRecover(func() error { return nil }); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
}

func TestCallWithOCRRecoverRethrowsOtherPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("want panic re-thrown")
		}
	}()
	_ = callWithOCRRecover(func() error { panic("boom") })
}
