package service

import (
	"testing"

	"yatori-go-console/config"
)

func TestNormalizeAccount(t *testing.T) {
	user := config.User{
		AccountType: "hqkj",
		URL:         "https://demo.haiqikeji.com/student/home",
		Account:     " tester ",
		Password:    "secret",
	}
	if err := normalizeAccount(&user, true); err != nil {
		t.Fatal(err)
	}
	if user.AccountType != "HQKJ" {
		t.Fatalf("account type = %s, want HQKJ", user.AccountType)
	}
	if user.URL != "https://demo.haiqikeji.com" {
		t.Fatalf("url = %s", user.URL)
	}
	if user.Account != "tester" {
		t.Fatalf("account = %s", user.Account)
	}
	if user.CoursesCustom.CxWorkSw == nil || *user.CoursesCustom.CxWorkSw != 1 {
		t.Fatal("default cxWorkSw should be 1")
	}
}

func TestNormalizeAccountRequiresURL(t *testing.T) {
	user := config.User{AccountType: "YINGHUA", Account: "tester", Password: "secret"}
	if err := normalizeAccount(&user, true); err == nil {
		t.Fatal("yinghua without url should fail")
	}
}

func TestMergeUserPreservesPassword(t *testing.T) {
	existing := config.User{AccountType: "XUEXITONG", Account: "tester", Password: "old"}
	incoming := config.User{AccountType: "XUEXITONG", Account: "tester", RemarkName: "备注"}
	merged := mergeUser(existing, incoming)
	if merged.Password != "old" {
		t.Fatalf("password = %s, want old", merged.Password)
	}
	if merged.RemarkName != "备注" {
		t.Fatalf("remark = %s", merged.RemarkName)
	}
}
