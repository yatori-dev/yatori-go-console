import apiClient from './base'
import type { ApiResponse } from './accountApi'

export type Course = {
    courseId: string
    courseName: string
    progress: number
    instructor: string
}

export type CourseListData = {
    courseList: Course[]
}

export const getCourseList = async (uid: string): Promise<ApiResponse<CourseListData>> => {
    try {
        return (await apiClient.get(`/v1/getAccountCourseList/${uid}`)) as ApiResponse<CourseListData>
    } catch (error: any) {
        return {
            code: 500,
            message:
                error.response?.data?.error ||
                error.response?.data?.message ||
                error.message ||
                '获取课程列表失败',
            data: { courseList: [] },
        }
    }
}
