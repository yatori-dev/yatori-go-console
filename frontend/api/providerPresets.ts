import type { AIConfig } from './systemApi'

export interface ProviderPreset {
  value: string
  label: string
  defaultBaseUrl: string
  defaultEndpoint: 'chat' | 'responses' | 'custom'
  defaultModel: string
  customEndpoint?: string
  hint?: string
}

export const PROVIDER_PRESETS: ProviderPreset[] = [
  { value: 'OPENAI',   label: 'OpenAI (官方)',         defaultBaseUrl: 'https://api.openai.com',                                       defaultEndpoint: 'responses', defaultModel: 'gpt-4o-mini', hint: '支持 /v1/responses（新 API）和 /v1/chat/completions' },
  { value: 'TONGYI',   label: '通义千问 (DashScope)',   defaultBaseUrl: 'https://dashscope.aliyuncs.com/compatible-mode',                defaultEndpoint: 'chat',      defaultModel: 'qwen-plus',   hint: '阿里云百炼 - OpenAI 兼容模式' },
  { value: 'DEEPSEEK', label: 'DeepSeek',              defaultBaseUrl: 'https://api.deepseek.com',                                       defaultEndpoint: 'chat',      defaultModel: 'deepseek-chat' },
  { value: 'SILICON',  label: '硅基流动 (SiliconFlow)', defaultBaseUrl: 'https://api.siliconflow.cn',                                    defaultEndpoint: 'chat',      defaultModel: 'Qwen/Qwen2.5-7B-Instruct' },
  { value: 'METAAI',   label: '秘塔 AI',               defaultBaseUrl: 'https://metaso.cn',                                             defaultEndpoint: 'custom',    customEndpoint: 'api/v1/chat/completions', defaultModel: 'metaso' },
  { value: 'OLLAMA',   label: 'Ollama (本地)',          defaultBaseUrl: 'http://localhost:11434',                                        defaultEndpoint: 'chat',      defaultModel: 'llama3.2' },
  { value: 'DOUBAO',   label: '豆包 (火山引擎)',        defaultBaseUrl: 'https://ark.cn-beijing.volces.com/api/v3',                       defaultEndpoint: 'chat',      defaultModel: 'doubao-pro-32k' },
  { value: 'CHATGLM',  label: '智谱 AI',                defaultBaseUrl: 'https://open.bigmodel.cn/api/paas/v4',                          defaultEndpoint: 'chat',      defaultModel: 'glm-4-flash' },
  { value: 'XINGHUO',  label: '星火 (讯飞)',            defaultBaseUrl: 'https://spark-api-open.xf-yun.com/v1',                          defaultEndpoint: 'chat',      defaultModel: 'general' },
  { value: 'OTHER',    label: '其他 (OpenAI 兼容)',     defaultBaseUrl: '',                                                              defaultEndpoint: 'chat',      defaultModel: '' },
  { value: 'CUSTOM',   label: '完全自定义',             defaultBaseUrl: '',                                                              defaultEndpoint: 'custom',    customEndpoint: '', defaultModel: '' },
]

export const getProviderPreset = (value: string): ProviderPreset | undefined => {
  return PROVIDER_PRESETS.find(p => p.value === value)
}

export const defaultConfigForProvider = (value: string): AIConfig => {
  const preset = getProviderPreset(value)
  if (!preset) {
    return { provider: value, model: '', apiKey: '', baseUrl: '', endpoint: 'chat' }
  }
  return {
    provider: preset.value,
    model: preset.defaultModel,
    apiKey: '',
    baseUrl: preset.defaultBaseUrl,
    endpoint: preset.defaultEndpoint,
    customEndpoint: preset.customEndpoint || '',
  }
}
