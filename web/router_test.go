package web

import "testing"

// 路由注册冒烟测试：新增接口如果和已有路由冲突，gin 会在注册时 panic。
func TestLoginCodeRoutesRegistered(t *testing.T) {
	router := serverInit()

	want := map[string]bool{
		"GET /api/v1/accountLoginCode/:uid":       false,
		"POST /api/v1/accountLoginCode/:uid":      false,
		"POST /api/v1/accountLoginCode/:uid/send": false,
	}
	for _, route := range router.Routes() {
		key := route.Method + " " + route.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for key, found := range want {
		if !found {
			t.Fatalf("route not registered: %s", key)
		}
	}
}
