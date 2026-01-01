import apiClient from "@/api/base";
import {DeleteAccountsResponse} from "@/api/accountApi";

// 删除账号的API返回结构
export interface CourseListResponse {
    code: number
    message: string
    data?:any
}
// 获取课程列表的API
export const getCourseList = async (uid: string): Promise<CourseListResponse> => {
    try {
        // 统一使用类型断言保持一致性
        return await apiClient.get(`/v1/getAccountCourseList/${uid}`) as CourseListResponse;
    } catch (error: any) {
        console.error('Error in deleteAccount:', error)
        return {
            code: 500,
            message: error.response?.data?.message || '删除账号失败'
        }
    }
}
