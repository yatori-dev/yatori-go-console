import apiClient from './base'

export type CoursesSetting = {
    name: string
    includeExams: string[]
    excludeExams: string[]
}

export type CoursesCustom = {
    studyTime: string
    cxNode?: number | null
    cxChapterTestSw?: number | null
    cxWorkSw?: number | null
    cxExamSw?: number | null
    shuffleSw: number
    videoModel: number
    autoExam: number
    examAutoSubmit: number
    excludeCourses: string[]
    includeCourses: string[]
    coursesSettings: CoursesSetting[]
}

export type AccountConfig = {
    accountType: string
    url: string
    remarkName?: string
    account: string
    password: string
    isProxy: number
    informEmails: string[]
    coursesCustom: CoursesCustom
}

export type UserPO = {
    uid: string
    accountType: string
    url: string
    account: string
    password: string
    userConfigJson: string
}

export type AccountStatus = {
    uid: string
    state: 'idle' | 'starting' | 'running' | 'stopping' | 'succeeded' | 'failed' | 'canceled'
    taskId?: string
    courses?: string[]
    startedAt?: string
    finishedAt?: string
    lastError?: string
}

export type Account = UserPO & {
    isRunning: boolean
    lastError?: string
    status?: AccountStatus
    config: AccountConfig
}

export type ApiResponse<T = unknown> = {
    code: number
    message: string
    data?: T
}

export type AccountListData = {
    users: Account[]
    total: number
    page: number
    size: number
}

export type AccountPayload = {
    user: UserPO
    config: AccountConfig
}

export const defaultCoursesCustom = (): CoursesCustom => ({
    studyTime: '10-30',
    cxNode: 3,
    cxChapterTestSw: 1,
    cxWorkSw: 1,
    cxExamSw: 1,
    shuffleSw: 0,
    videoModel: 1,
    autoExam: 0,
    examAutoSubmit: 0,
    excludeCourses: [],
    includeCourses: [],
    coursesSettings: [],
})

export const defaultAccountConfig = (): AccountConfig => ({
    accountType: 'XUEXITONG',
    url: '',
    remarkName: '',
    account: '',
    password: '',
    isProxy: 0,
    informEmails: [],
    coursesCustom: defaultCoursesCustom(),
})

export const normalizeConfig = (config?: Partial<AccountConfig> | null): AccountConfig => {
    const base = defaultAccountConfig()
    if (!config) {
        return base
    }
    return {
        ...base,
        ...config,
        informEmails: Array.isArray(config.informEmails) ? config.informEmails : [],
        coursesCustom: {
            ...base.coursesCustom,
            ...(config.coursesCustom || {}),
            excludeCourses: Array.isArray(config.coursesCustom?.excludeCourses)
                ? config.coursesCustom!.excludeCourses
                : [],
            includeCourses: Array.isArray(config.coursesCustom?.includeCourses)
                ? config.coursesCustom!.includeCourses
                : [],
            coursesSettings: Array.isArray(config.coursesCustom?.coursesSettings)
                ? config.coursesCustom!.coursesSettings
                : [],
        },
    }
}

export const getAccounts = async (page = 1, pageSize = 100): Promise<ApiResponse<AccountListData>> => {
    try {
        return (await apiClient.get('/v1/accountList', {
            params: { page, pageSize },
        })) as ApiResponse<AccountListData>
    } catch (error: any) {
        return {
            code: 500,
            message: error.response?.data?.message || '获取账号列表失败',
            data: { users: [], total: 0, page, size: pageSize },
        }
    }
}

export const getAccount = async (uid: string): Promise<ApiResponse<{ user: UserPO; config: AccountConfig; isRunning: boolean; lastError?: string }>> => {
    try {
        return (await apiClient.get(`/v1/getAccountInformForUid/${uid}`)) as ApiResponse<{
            user: UserPO
            config: AccountConfig
            isRunning: boolean
            lastError?: string
        }>
    } catch (error: any) {
        return { code: 500, message: error.response?.data?.message || '获取账号信息失败' }
    }
}

export const addAccount = async (config: AccountConfig): Promise<ApiResponse<AccountPayload>> => {
    try {
        return (await apiClient.post('/v1/addAccount', config)) as ApiResponse<AccountPayload>
    } catch (error: any) {
        return { code: 500, message: error.response?.data?.message || '添加账号失败' }
    }
}

export const updateAccount = async (uid: string, config: AccountConfig): Promise<ApiResponse<AccountPayload>> => {
    try {
        return (await apiClient.post('/v1/updateAccount', { uid, ...config })) as ApiResponse<AccountPayload>
    } catch (error: any) {
        return { code: 500, message: error.response?.data?.message || '更新账号失败' }
    }
}

export const deleteAccountForUid = async (uid: string): Promise<ApiResponse> => {
    try {
        return (await apiClient.post('/v1/deleteAccount', { uid })) as ApiResponse
    } catch (error: any) {
        return { code: 500, message: error.response?.data?.message || '删除账号失败' }
    }
}

export const checkAccountLogin = async (uid: string): Promise<ApiResponse<{ platform: string }>> => {
    try {
        return (await apiClient.post('/v1/accountLoginCheck', { uid })) as ApiResponse<{ platform: string }>
    } catch (error: any) {
        return { code: 500, message: error.response?.data?.message || '登录检测失败' }
    }
}

export const getAccountLog = async (uid: string, limit = 300): Promise<ApiResponse<{ lines: string[] }>> => {
    try {
        return (await apiClient.get(`/v1/accountLog/${uid}`, { params: { limit } })) as ApiResponse<{
            lines: string[]
        }>
    } catch (error: any) {
        return { code: 500, message: error.response?.data?.message || '获取运行日志失败', data: { lines: [] } }
    }
}

export const getAccountStatus = async (uid: string): Promise<ApiResponse<{ status: AccountStatus }>> => {
    try {
        return (await apiClient.get(`/v1/accountStatus/${uid}`)) as ApiResponse<{ status: AccountStatus }>
    } catch (error: any) {
        return { code: 500, message: error.response?.data?.message || '获取任务状态失败' }
    }
}

export const startBrushForUid = async (uid: string, courses: string[] = []): Promise<ApiResponse> => {
    try {
        return (await apiClient.post(`/v1/startBrush/${uid}`, { courses })) as ApiResponse
    } catch (error: any) {
        // 兼容旧版后端：旧版只注册了 GET 路由
        if (error.response?.status === 404) {
            try {
                return (await apiClient.get(`/v1/startBrush/${uid}`)) as ApiResponse
            } catch (fallbackError: any) {
                return {
                    code: 500,
                    message: fallbackError.response?.data?.message || '启动任务失败',
                }
            }
        }
        return { code: 500, message: error.response?.data?.message || '启动任务失败' }
    }
}

export const stopBrushForUid = async (uid: string): Promise<ApiResponse> => {
    try {
        return (await apiClient.post(`/v1/stopBrush/${uid}`)) as ApiResponse
    } catch (error: any) {
        // 兼容旧版后端：旧版只注册了 GET 路由
        if (error.response?.status === 404) {
            try {
                return (await apiClient.get(`/v1/stopBrush/${uid}`)) as ApiResponse
            } catch (fallbackError: any) {
                return {
                    code: 500,
                    message: fallbackError.response?.data?.message || '停止任务失败',
                }
            }
        }
        return { code: 500, message: error.response?.data?.message || '停止任务失败' }
    }
}
