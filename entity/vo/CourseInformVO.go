package vo

type CourseInformResponse struct {
	CourseId   string  `json:"courseId"`   //课程ID
	CourseName string  `json:"courseName"` //课程名称
	Instructor string  `json:"instructor"` //授课老师
	Progress   float32 `json:"progress"`   //课程进度
}
