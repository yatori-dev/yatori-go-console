package examples

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func A(ctx context.Context) {
	if err := B(ctx); err != nil {
		fmt.Println("A 提前结束：", err)
		return
	}
	fmt.Println("A函数")
}

func B(ctx context.Context) error {
	if err := C(ctx); err != nil {
		fmt.Println("B 提前结束：", err)
		return err
	}
	fmt.Println("B函数")
	return nil
}

func C(ctx context.Context) error {
	for {
		time.Sleep(1 * time.Second)
		select {
		case <-ctx.Done():
			return ctx.Err() // 这里把“被取消”的信息抛出去
		default:
			fmt.Println("调用循环")
		}
	}
}
func TestChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		A(ctx)
	}()

	time.Sleep(8 * time.Second)
	cancel() // ⬅ 整个 A → B → C 立刻全部退出
	for {
	}
}
