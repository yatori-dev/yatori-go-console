package runner

import (
	"testing"

	"yatori-go-console/config"
)

func TestNormalizeLoginMode(t *testing.T) {
	cases := map[string]string{
		"app":     "app",
		"Android": "app",
		" MOBILE": "app",
		"phone":   "app",
		"web":     "web",
		"":        "web",
		"pc":      "web",
		"desktop": "web",
	}
	for input, want := range cases {
		if got := normalizeLoginMode(input); got != want {
			t.Fatalf("normalizeLoginMode(%q)=%q want %q", input, got, want)
		}
	}

	if !isAppLogin(config.User{LoginMode: "Android"}) {
		t.Fatal("Android 应被识别为 App 登录")
	}
	if isAppLogin(config.User{LoginMode: "web"}) {
		t.Fatal("web 不应被识别为 App 登录")
	}
}
