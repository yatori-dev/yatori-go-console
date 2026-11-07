package runner

import (
	"context"
	"testing"
	"time"

	"yatori-go-console/config"

	yatori "github.com/yatori-dev/yatori-go-core/v3"
)

type fakeClient struct {
	platform   yatori.Platform
	courses    []yatori.Course
	nodes      []yatori.Node
	works      []yatori.Work
	exams      []yatori.Exam
	studyCalls int
	workCalls  int
	examCalls  int
}

func (f *fakeClient) Platform() yatori.Platform { return f.platform }
func (f *fakeClient) Capabilities() yatori.Capability {
	return yatori.CapLogin | yatori.CapCourses | yatori.CapTree | yatori.CapStudy | yatori.CapWorks | yatori.CapExams
}
func (f *fakeClient) Login(context.Context) error                      { return nil }
func (f *fakeClient) LoginWithCookie(context.Context, string) error    { return nil }
func (f *fakeClient) Session() yatori.Session                          { return yatori.Session{} }
func (f *fakeClient) Courses(context.Context) ([]yatori.Course, error) { return f.courses, nil }
func (f *fakeClient) Tree(context.Context, yatori.Course) ([]yatori.Node, error) {
	return f.nodes, nil
}
func (f *fakeClient) Study(context.Context, yatori.StudyRequest) (*yatori.StudyResult, error) {
	f.studyCalls++
	return &yatori.StudyResult{Progress: 100, Submitted: true, Completed: true}, nil
}
func (f *fakeClient) Works(context.Context, yatori.Course) ([]yatori.Work, error) {
	return f.works, nil
}
func (f *fakeClient) DoWork(context.Context, yatori.WorkRequest) (*yatori.AnswerResult, error) {
	f.workCalls++
	return &yatori.AnswerResult{Title: "work", Submitted: true, Message: "ok"}, nil
}
func (f *fakeClient) Exams(context.Context, yatori.Course) ([]yatori.Exam, error) {
	return f.exams, nil
}
func (f *fakeClient) DoExam(context.Context, yatori.ExamRequest) (*yatori.AnswerResult, error) {
	f.examCalls++
	return &yatori.AnswerResult{Title: "exam", Submitted: true, Message: "ok"}, nil
}
func (f *fakeClient) SignIn(context.Context, yatori.SignInRequest) (*yatori.SignInResult, error) {
	return nil, nil
}
func (f *fakeClient) Call(context.Context, yatori.Action, map[string]any) (any, error) {
	return nil, nil
}
func (f *fakeClient) Close() error { return nil }

func TestAnswerSource(t *testing.T) {
	cases := []struct {
		platform yatori.Platform
		autoExam int
		want     yatori.AnswerSource
	}{
		{yatori.XueXiTong, 0, yatori.AnswerSourceManual},
		{yatori.XueXiTong, 1, yatori.AnswerSourceAI},
		{yatori.XueXiTong, 2, yatori.AnswerSourceBank},
		{yatori.XueXiTong, 3, yatori.AnswerSourceAI},
		{yatori.HaiQiKeJi, 2, yatori.AnswerSourceBuiltin},
	}
	for _, item := range cases {
		user := config.User{CoursesCustom: config.CoursesCustom{AutoExam: item.autoExam}}
		if got := answerSource(item.platform, user); got != item.want {
			t.Fatalf("platform=%s autoExam=%d: got %s, want %s", item.platform, item.autoExam, got, item.want)
		}
	}
}

func TestCourseAllowed(t *testing.T) {
	user := config.User{CoursesCustom: config.CoursesCustom{
		IncludeCourses: []string{"课程A"},
		ExcludeCourses: []string{"课程B"},
	}}
	if !courseAllowed(user, yatori.Course{Name: "课程A"}) {
		t.Fatal("include course should be allowed")
	}
	if courseAllowed(user, yatori.Course{Name: "课程B"}) {
		t.Fatal("exclude course should be skipped")
	}
	if courseAllowed(user, yatori.Course{Name: "课程C"}) {
		t.Fatal("course not in include list should be skipped")
	}
}

func TestMatchCourseENAEAStyle(t *testing.T) {
	course := yatori.Course{Name: "课程A", Extra: map[string]string{"projectName": "项目A", "titleTag": "模块1"}}
	if !matchCourse(course, []string{"项目A-->模块1"}) {
		t.Fatal("ENAAE style pattern should match project and title tag")
	}
}

func TestCourseDateAllowed(t *testing.T) {
	now := time.Now()
	course := yatori.Course{Extra: map[string]string{
		"startDate": now.AddDate(0, 0, 1).Format("2006-01-02"),
	}}
	if courseDateAllowed(course) {
		t.Fatal("future course should be skipped")
	}
	course.Extra["startDate"] = now.AddDate(0, 0, -2).Format("2006-01-02")
	course.Extra["endDate"] = now.AddDate(0, 0, -1).Format("2006-01-02")
	if courseDateAllowed(course) {
		t.Fatal("expired course should be skipped")
	}
}

func TestIsFinishedStatus(t *testing.T) {
	for _, status := range []string{"finished", "已完成", "已交卷", "complete"} {
		if !isFinishedStatus(status) {
			t.Fatalf("%q should be treated as finished", status)
		}
	}
	for _, status := range []string{"", "进行中", "未交", "未完成", "incomplete"} {
		if isFinishedStatus(status) {
			t.Fatalf("%q should not be treated as finished", status)
		}
	}
}

func TestRunClientFake(t *testing.T) {
	client := &fakeClient{
		platform: yatori.XueXiTong,
		courses:  []yatori.Course{{ID: "c1", Name: "课程1"}},
		nodes:    []yatori.Node{{ID: "n1", CourseID: "c1", Name: "视频1", Kind: yatori.NodeVideo, Duration: 30 * time.Second}},
		works:    []yatori.Work{{ID: "w1", Name: "作业1"}},
		exams:    []yatori.Exam{{ID: "e1", Name: "考试1"}},
	}
	user := config.User{
		AccountType: "XUEXITONG",
		Account:     "tester",
		CoursesCustom: config.CoursesCustom{
			VideoModel:     2,
			AutoExam:       1,
			ExamAutoSubmit: 1,
		},
	}
	RunClient(context.Background(), client, config.Setting{}, user, nil)
	if client.studyCalls == 0 {
		t.Fatal("expected Study to be called")
	}
	if client.workCalls != 1 {
		t.Fatalf("expected 1 DoWork call, got %d", client.workCalls)
	}
	if client.examCalls != 1 {
		t.Fatalf("expected 1 DoExam call, got %d", client.examCalls)
	}
}

func TestRunClientFakeHQKJBatch(t *testing.T) {
	client := &fakeClient{
		platform: yatori.HaiQiKeJi,
		courses:  []yatori.Course{{ID: "c1", Name: "课程1"}},
		nodes:    []yatori.Node{{ID: "n1", CourseID: "c1", Name: "视频1", Kind: yatori.NodeVideo, Duration: 30 * time.Second}},
		works:    []yatori.Work{{ID: "w1", Name: "作业1"}, {ID: "w2", Name: "作业2"}},
		exams:    []yatori.Exam{{ID: "e1", Name: "考试1"}, {ID: "e2", Name: "考试2"}},
	}
	user := config.User{
		AccountType: "HQKJ",
		Account:     "tester",
		CoursesCustom: config.CoursesCustom{
			VideoModel:     2,
			AutoExam:       2,
			ExamAutoSubmit: 1,
		},
	}
	RunClient(context.Background(), client, config.Setting{}, user, nil)
	if client.workCalls != 1 {
		t.Fatalf("HQKJ DoWork should be called once per course, got %d", client.workCalls)
	}
	if client.examCalls != 1 {
		t.Fatalf("HQKJ DoExam should be called once per course, got %d", client.examCalls)
	}
}
