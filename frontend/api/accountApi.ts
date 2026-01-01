import apiClient from './base'
import type { Account } from '@/components/account-list'

// 定义请求和响应类型
export interface AddAccountParams {
  accountType: string
  url:string
  account: string
  password: string
}

export interface AddAccountResponse {
  code: number
  message: string
  data?: Account
}

// 添加账号
export const addAccount = async (params: AddAccountParams): Promise<AddAccountResponse> => {
  try {
    // 统一使用类型断言保持一致性
    return await apiClient.post('/v1/addAccount', params) as AddAccountResponse;
  } catch (error: any) {
    return {
      code: 500,
      message: error.response?.data?.message || '添加账号失败'
    }
  }
}

// 获取账号列表的API返回结构
export interface GetAccountsResponse {
  code: number
  message: string
  data: {
    users: Array<{ uid:string;accountType: string;url: string; account: string; password: string;isRunning: boolean; userConfigJson: string }>
    total: number
  }
}
//获取账号列表的API
export const getAccounts = async (): Promise<GetAccountsResponse> => {
  try {
    // 使用as断言告诉TypeScript，apiClient.get返回的就是GetAccountsResponse类型
    return await apiClient.get('/v1/accountList') as GetAccountsResponse;
  } catch (error: any) {
    console.error('Error in getAccounts:', error)
    return {
      code: 500,
      message: error.response?.data?.message || '获取账号列表失败',
      data: {
        users: [],
        total: 0
      }
    }
  }
}

// 删除账号的API返回结构
export interface DeleteAccountsResponse {
  code: number
  message: string
}
//删除账号
export const deleteAccountForUid = async (uid: string): Promise<DeleteAccountsResponse> => {
  try {
    // 统一使用类型断言保持一致性
    return await apiClient.post(`/v1/deleteAccount`, { uid: uid }) as DeleteAccountsResponse;
  } catch (error: any) {
    console.error('Error in deleteAccount:', error)
    return {
      code: 500,
      message: error.response?.data?.message || '删除账号失败'
    }
  }
}

// 更新账号配置


export interface StartBrushForUidResponse {
    code: number
    message: string
}
// 开始刷课API
export const startBrushForUid = async (uid: string): Promise<StartBrushForUidResponse> => {
    try {
        // 统一使用类型断言保持一致性
        return await apiClient.post(`/v1/startBrush/${uid}`) as StartBrushForUidResponse;
    } catch (error: any) {
        console.error('Error in deleteAccount:', error)
        return {
            code: 500,
            message: error.response?.data?.message || '删除账号失败'
        }
    }
}
export interface StopBrushForUidResponse {
    code: number
    message: string
}
// 取消刷课API
export const stopBrushForUid = async (uid: string): Promise<StopBrushForUidResponse> => {
    try {
        // 统一使用类型断言保持一致性
        return await apiClient.post(`/v1/stopBrush/${uid}`) as StopBrushForUidResponse;
    } catch (error: any) {
        console.error('Error in deleteAccount:', error)
        return {
            code: 500,
            message: error.response?.data?.message || '删除账号失败'
        }
    }
}
