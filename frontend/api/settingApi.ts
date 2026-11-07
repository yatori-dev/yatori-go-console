import apiClient from './base'
import type { ApiResponse } from './accountApi'

export type BasicSetting = {
    completionTone: number
    colorLog: number
    logOutFileSw: number
    logLevel: string
    logModel: number
    webModel?: number
}

export type EmailInform = {
    sw: number
    SMTPHost: string
    SMTPPort: number
    userName: string
    password: string
}

export type AiSetting = {
    aiType: string
    aiUrl: string
    model: string
    API_KEY: string
}

export type ApiQueSetting = {
    url: string
}

export type Setting = {
    basicSetting: BasicSetting
    emailInform: EmailInform
    aiSetting: AiSetting
    apiQueSetting: ApiQueSetting
}

export const defaultSetting = (): Setting => ({
    basicSetting: {
        completionTone: 1,
        colorLog: 1,
        logOutFileSw: 1,
        logLevel: 'INFO',
        logModel: 0,
        webModel: 1,
    },
    emailInform: {
        sw: 0,
        SMTPHost: '',
        SMTPPort: 465,
        userName: '',
        password: '',
    },
    aiSetting: {
        aiType: 'TONGYI',
        aiUrl: '',
        model: '',
        API_KEY: '',
    },
    apiQueSetting: {
        url: '',
    },
})

export const normalizeSetting = (setting?: Partial<Setting> | null): Setting => {
    const base = defaultSetting()
    if (!setting) {
        return base
    }
    return {
        basicSetting: { ...base.basicSetting, ...(setting.basicSetting || {}) },
        emailInform: { ...base.emailInform, ...(setting.emailInform || {}) },
        aiSetting: { ...base.aiSetting, ...(setting.aiSetting || {}) },
        apiQueSetting: { ...base.apiQueSetting, ...(setting.apiQueSetting || {}) },
    }
}

export const getSetting = async (): Promise<ApiResponse<{ setting: Setting }>> => {
    try {
        return (await apiClient.get('/v1/setting')) as ApiResponse<{ setting: Setting }>
    } catch (error: any) {
        return { code: 500, message: error.response?.data?.message || '读取设置失败' }
    }
}

export const saveSetting = async (setting: Setting): Promise<ApiResponse<{ setting: Setting }>> => {
    try {
        return (await apiClient.post('/v1/setting', setting)) as ApiResponse<{ setting: Setting }>
    } catch (error: any) {
        return { code: 500, message: error.response?.data?.message || '保存设置失败' }
    }
}
