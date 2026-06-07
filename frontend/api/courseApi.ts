import apiClient from "@/api/base";

export interface CourseListResponse {
    code: number
    message: string
    data?:any
}

// 获取课程列表的API
export const getCourseList = async (uid: string): Promise<CourseListResponse> => {
    try {
        return await apiClient.get(`/v1/getAccountCourseList/${uid}`, {
            timeout: 30000,
        }) as CourseListResponse;
    } catch (error: any) {
        console.error('Error in getCourseList:', error)
        return {
            code: 500,
            message: error.code === 'ECONNABORTED'
                ? '课程加载超时，请稍后重试'
                : error.response?.data?.message || '获取课程列表失败'
        }
    }
}
