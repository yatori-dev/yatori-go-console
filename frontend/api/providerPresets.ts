import type { AIConfig } from './systemApi'

export interface ProviderPreset {
  value: string
  label: string
  defaultBaseUrl: string
  defaultEndpoint: 'chat' | 'responses' | 'custom'
  defaultModel: string
  runtimeProvider?: string
  customEndpoint?: string
  hint?: string
  models?: string[]
  supportsCustomEndpoint?: boolean
}

const compatibleHint = 'OpenAI 兼容接口，运行时将使用 OTHER，以确保自定义端点生效'

export const PROVIDER_PRESETS: ProviderPreset[] = [
  { value: 'OPENAI', label: 'OpenAI (官方)', defaultBaseUrl: 'https://api.openai.com', defaultEndpoint: 'responses', defaultModel: 'gpt-4o-mini', runtimeProvider: 'OPENAI', hint: '支持 /v1/responses（新 API）和 /v1/chat/completions', models: ['gpt-4o-mini', 'gpt-4o', 'gpt-4.1-mini', 'gpt-4.1'] },
  { value: 'TONGYI', label: '通义千问 (DashScope)', defaultBaseUrl: 'https://dashscope.aliyuncs.com/compatible-mode', defaultEndpoint: 'chat', defaultModel: 'qwen-plus', runtimeProvider: 'TONGYI', hint: '阿里云百炼 - OpenAI 兼容模式', models: ['qwen-plus', 'qwen-turbo', 'qwen-max', 'qwen-long'] },
  { value: 'DEEPSEEK', label: 'DeepSeek', defaultBaseUrl: 'https://api.deepseek.com', defaultEndpoint: 'chat', defaultModel: 'deepseek-chat', runtimeProvider: 'DEEPSEEK', models: ['deepseek-chat', 'deepseek-reasoner'] },
  { value: 'SILICON', label: '硅基流动 (SiliconFlow)', defaultBaseUrl: 'https://api.siliconflow.cn', defaultEndpoint: 'chat', defaultModel: 'Qwen/Qwen2.5-7B-Instruct', runtimeProvider: 'SILICON', models: ['Qwen/Qwen2.5-7B-Instruct', 'deepseek-ai/DeepSeek-V3', 'deepseek-ai/DeepSeek-R1'] },
  { value: 'METAAI', label: '秘塔 AI', defaultBaseUrl: 'https://metaso.cn', defaultEndpoint: 'custom', customEndpoint: 'api/v1/chat/completions', defaultModel: 'metaso', runtimeProvider: 'METAAI' },
  { value: 'OLLAMA', label: 'Ollama (本地)', defaultBaseUrl: 'http://localhost:11434', defaultEndpoint: 'chat', defaultModel: 'llama3.2', runtimeProvider: 'OTHER', hint: compatibleHint, models: ['llama3.2', 'qwen2.5', 'deepseek-r1'] },
  { value: 'DOUBAO', label: '豆包 (火山引擎)', defaultBaseUrl: 'https://ark.cn-beijing.volces.com/api/v3', defaultEndpoint: 'chat', defaultModel: 'doubao-pro-32k', runtimeProvider: 'DOUBAO', models: ['doubao-pro-32k', 'doubao-lite-32k'] },
  { value: 'CHATGLM', label: '智谱 AI', defaultBaseUrl: 'https://open.bigmodel.cn/api/paas/v4', defaultEndpoint: 'chat', defaultModel: 'glm-4-flash', runtimeProvider: 'CHATGLM', models: ['glm-4-flash', 'glm-4-plus', 'glm-4-air'] },
  { value: 'XINGHUO', label: '星火 (讯飞)', defaultBaseUrl: 'https://spark-api-open.xf-yun.com/v1', defaultEndpoint: 'chat', defaultModel: 'general', runtimeProvider: 'XINGHUO', models: ['general', 'generalv3.5', '4.0Ultra'] },
  { value: 'OPENROUTER', label: 'OpenRouter', defaultBaseUrl: 'https://openrouter.ai/api', defaultEndpoint: 'chat', defaultModel: 'openai/gpt-4o-mini', runtimeProvider: 'OTHER', hint: compatibleHint, models: ['openai/gpt-4o-mini', 'deepseek/deepseek-chat', 'qwen/qwen-plus'] },
  { value: 'MOONSHOT', label: 'Moonshot / Kimi', defaultBaseUrl: 'https://api.moonshot.cn', defaultEndpoint: 'chat', defaultModel: 'moonshot-v1-8k', runtimeProvider: 'OTHER', hint: compatibleHint, models: ['moonshot-v1-8k', 'moonshot-v1-32k', 'kimi-k2-0711-preview'] },
  { value: 'MINIMAX', label: 'MiniMax', defaultBaseUrl: 'https://api.minimax.chat', defaultEndpoint: 'chat', defaultModel: 'abab6.5s-chat', runtimeProvider: 'OTHER', hint: compatibleHint, models: ['abab6.5s-chat', 'MiniMax-Text-01'] },
  { value: 'BAICHUAN', label: '百川智能', defaultBaseUrl: 'https://api.baichuan-ai.com', defaultEndpoint: 'chat', defaultModel: 'Baichuan4', runtimeProvider: 'OTHER', hint: compatibleHint, models: ['Baichuan4', 'Baichuan3-Turbo'] },
  { value: 'STEPFUN', label: '阶跃星辰 (StepFun)', defaultBaseUrl: 'https://api.stepfun.com', defaultEndpoint: 'chat', defaultModel: 'step-1-8k', runtimeProvider: 'OTHER', hint: compatibleHint, models: ['step-1-8k', 'step-1-32k', 'step-2-16k'] },
  { value: 'AI302', label: '302.AI', defaultBaseUrl: 'https://api.302.ai', defaultEndpoint: 'chat', defaultModel: 'gpt-4o-mini', runtimeProvider: 'OTHER', hint: compatibleHint, models: ['gpt-4o-mini', 'deepseek-chat', 'qwen-plus'] },
  { value: 'GROQ', label: 'Groq', defaultBaseUrl: 'https://api.groq.com/openai', defaultEndpoint: 'chat', defaultModel: 'llama-3.1-8b-instant', runtimeProvider: 'OTHER', hint: compatibleHint, models: ['llama-3.1-8b-instant', 'llama-3.3-70b-versatile', 'mixtral-8x7b-32768'] },
  { value: 'TOGETHER', label: 'Together AI', defaultBaseUrl: 'https://api.together.xyz', defaultEndpoint: 'chat', defaultModel: 'meta-llama/Llama-3.3-70B-Instruct-Turbo', runtimeProvider: 'OTHER', hint: compatibleHint, models: ['meta-llama/Llama-3.3-70B-Instruct-Turbo', 'Qwen/Qwen2.5-72B-Instruct-Turbo'] },
  { value: 'PERPLEXITY', label: 'Perplexity', defaultBaseUrl: 'https://api.perplexity.ai', defaultEndpoint: 'chat', defaultModel: 'sonar', runtimeProvider: 'OTHER', hint: compatibleHint, models: ['sonar', 'sonar-pro', 'sonar-reasoning'] },
  { value: 'MODELSCOPE', label: 'ModelScope 魔搭', defaultBaseUrl: 'https://api-inference.modelscope.cn', defaultEndpoint: 'chat', defaultModel: 'Qwen/Qwen2.5-7B-Instruct', runtimeProvider: 'OTHER', hint: compatibleHint, models: ['Qwen/Qwen2.5-7B-Instruct', 'Qwen/Qwen2.5-72B-Instruct'] },
  { value: 'LMSTUDIO', label: 'LM Studio (本地)', defaultBaseUrl: 'http://localhost:1234', defaultEndpoint: 'chat', defaultModel: 'local-model', runtimeProvider: 'OTHER', hint: compatibleHint },
  { value: 'OTHER', label: '其他 (OpenAI 兼容)', defaultBaseUrl: '', defaultEndpoint: 'chat', defaultModel: '', runtimeProvider: 'OTHER', hint: compatibleHint },
  { value: 'CUSTOM', label: '完全自定义', defaultBaseUrl: '', defaultEndpoint: 'custom', customEndpoint: '', defaultModel: '', runtimeProvider: 'OTHER', hint: '完全自定义 URL 和路径，运行时使用 OTHER' },
]

export const getProviderPreset = (value: string): ProviderPreset | undefined => {
  return PROVIDER_PRESETS.find(p => p.value === value)
}

export const defaultConfigForProvider = (value: string): AIConfig => {
  const preset = getProviderPreset(value)
  if (!preset) {
    return { provider: value, runtimeProvider: 'OTHER', model: '', apiKey: '', baseUrl: '', endpoint: 'chat' }
  }
  return {
    provider: preset.value,
    runtimeProvider: preset.runtimeProvider || preset.value,
    model: preset.defaultModel,
    apiKey: '',
    baseUrl: preset.defaultBaseUrl,
    endpoint: preset.defaultEndpoint,
    customEndpoint: preset.customEndpoint || '',
  }
}
