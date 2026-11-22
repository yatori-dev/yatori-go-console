package runner

import (
	"errors"
	"testing"
)

func TestIsLoginPageError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "课程接口返回登录页",
			err:  errors.New("yatori: XUEXITONG: courses: invalid character '<' looking for beginning of value"),
			want: true,
		},
		{name: "HTML 关键字", err: errors.New("unexpected <html> body"), want: true},
		{name: "DOCTYPE", err: errors.New("got <!DOCTYPE html>"), want: true},
		{name: "普通网络错误", err: errors.New("dial tcp: i/o timeout"), want: false},
		{name: "空错误", err: nil, want: false},
	}
	for _, item := range cases {
		if got := isLoginPageError(item.err); got != item.want {
			t.Fatalf("%s: isLoginPageError=%v want %v", item.name, got, item.want)
		}
	}
}
