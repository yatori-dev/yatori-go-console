import apiClient from './base'

// ============================================================
// AI 配置（与后端 /v1/saveAiConfig / /v1/testAiConfig / /v1/getAiConfig 对齐）
// ============================================================

export interface AIConfig {
  provider: string       // 供应商标识 (OPENAI / TONGYI / DEEPSEEK / SILICON / METAAI / DOUBAO / CHATGLM / XINGHUO / OLLAMA / OTHER / CUSTOM)
  model: string
  apiKey: string
  baseUrl: string
  endpoint: string       // chat | responses | custom:<path>
  customEndpoint?: string
  aiUrl?: string
}

export interface AIConfigResponse {
  success: boolean
  message?: string
  aiSetting?: AIConfig
  externalBankUrl?: string
}

export interface AIApiResponse {
  success: boolean
  message: string
}

export const getAiConfig = async (): Promise<AIConfigResponse> => {
  try {
    return await apiClient.get('/v1/getAiConfig') as AIConfigResponse
  } catch (error: any) {
    return { success: false, message: error.response?.data?.message || '加载AI配置失败' }
  }
}

export const saveAiConfig = async (config: AIConfig): Promise<AIApiResponse> => {
  try {
    return await apiClient.post('/v1/saveAiConfig', {
      provider: config.provider,
      model: config.model,
      apiKey: config.apiKey,
      baseUrl: config.baseUrl,
      endpoint: config.endpoint,
      customEndpoint: config.customEndpoint || '',
    }) as AIApiResponse
  } catch (error: any) {
    return { success: false, message: error.response?.data?.message || error.response?.data?.error || '保存AI配置失败' }
  }
}

export const testAiConfig = async (config: AIConfig): Promise<AIApiResponse> => {
  try {
    return await apiClient.post('/v1/testAiConfig', {
      provider: config.provider,
      model: config.model,
      apiKey: config.apiKey,
      baseUrl: config.baseUrl,
      endpoint: config.endpoint,
      customEndpoint: config.customEndpoint || '',
    }, { timeout: 60000 }) as AIApiResponse
  } catch (error: any) {
    return { success: false, message: error.response?.data?.message || error.response?.data?.error || '测试AI连接失败' }
  }
}
