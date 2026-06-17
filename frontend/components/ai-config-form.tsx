"use client"

import { useEffect, useMemo, useState } from "react"
import { motion } from "framer-motion"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Eye, EyeOff, Save, Bot, TestTube, Loader2, ExternalLink, RotateCcw, CheckCircle2, XCircle } from "lucide-react"
import { getAiConfig, saveAiConfig, testAiConfig, AIConfig, AIApiResponse } from "@/api/systemApi"
import { PROVIDER_PRESETS, getProviderPreset, defaultConfigForProvider } from "@/api/providerPresets"
import { useToast } from "@/hooks/use-toast"

export type AIConfigFormProps = Record<string, never>

const endpointPath = (endpoint: string, customEndpoint?: string) => {
  if (endpoint === "responses") return "/v1/responses"
  if (endpoint === "custom") {
    const ep = (customEndpoint || "").trim().replace(/^\/+/, "")
    return ep ? `/${ep}` : ""
  }
  return "/v1/chat/completions"
}

const fullAiUrl = (config: AIConfig) => {
  const baseUrl = (config.baseUrl || "").trim().replace(/\/+$/, "")
  const path = endpointPath(config.endpoint, config.customEndpoint)
  if (!baseUrl) return path
  return `${baseUrl}${path}`
}

export function AIConfigForm(_: AIConfigFormProps = {}) {
  const { toast } = useToast()
  const [config, setConfig] = useState<AIConfig>({
    provider: "", runtimeProvider: "", model: "", apiKey: "", baseUrl: "", endpoint: "chat", customEndpoint: "",
  })
  const [externalBankUrl, setExternalBankUrl] = useState("")
  const [showApiKey, setShowApiKey] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [isTesting, setIsTesting] = useState(false)
  const [originalProvider, setOriginalProvider] = useState("")
  const [testResult, setTestResult] = useState<(AIApiResponse & { testedAt: string }) | null>(null)

  // 初次加载
  useEffect(() => {
    (async () => {
      try {
        const r = await getAiConfig()
        if (r.success && r.aiSetting) {
          const a = r.aiSetting
          setConfig({
            provider: a.provider || "",
            runtimeProvider: a.runtimeProvider || a.provider || "",
            model: a.model || "",
            apiKey: a.apiKey || "",
            baseUrl: a.baseUrl || "",
            endpoint: a.endpoint || "chat",
            customEndpoint: a.customEndpoint || "",
            aiUrl: a.aiUrl,
          })
          setOriginalProvider(a.provider || "")
          if (a.apiKey) setShowApiKey(false)
        }
        if (r.externalBankUrl !== undefined) setExternalBankUrl(r.externalBankUrl)
      } catch (e) {
        console.error("加载AI配置失败:", e)
        toast({ variant: "destructive", title: "加载配置失败", description: String(e) })
      } finally {
        setIsLoading(false)
      }
    })()
  }, [])

  // 切换供应商时自动填默认值（保留用户已填的 apiKey）
  const onProviderChange = (newProvider: string) => {
    const def = defaultConfigForProvider(newProvider)
    setConfig(prev => ({
      ...def,
      apiKey: prev.apiKey,
      model: def.model || prev.model,
    }))
    setTestResult(null)
  }

  const providerPreset = useMemo(() => getProviderPreset(config.provider), [config.provider])
  const showCustomEndpoint = config.endpoint === "custom"
  const isFormValid = Boolean(config.provider && config.model && config.apiKey && config.baseUrl && (!showCustomEndpoint || config.customEndpoint?.trim()))
  const currentUrl = fullAiUrl(config)
  const suggestedModels = providerPreset?.models || []

  const handleSave = async () => {
    if (!isFormValid) {
      toast({ variant: "destructive", title: "保存失败", description: "请填写完整的供应商、模型、API 密钥、基础 URL 和端点" })
      return
    }
    setIsSaving(true)
    try {
      const r = await saveAiConfig(config)
      if (r.success) {
        toast({ title: "配置已保存", description: "AI 模型配置已写入 config.yaml，重启 yatori 后生效" })
      } else {
        toast({ variant: "destructive", title: "保存失败", description: r.message })
      }
    } catch (e: any) {
      toast({ variant: "destructive", title: "保存出错", description: String(e) })
    } finally {
      setIsSaving(false)
    }
  }

  const handleTest = async () => {
    if (!isFormValid) {
      toast({ variant: "destructive", title: "无法测试", description: "请先填写供应商、模型、API 密钥、基础 URL 和端点" })
      return
    }
    setIsTesting(true)
    setTestResult(null)
    try {
      const r = await testAiConfig(config)
      setTestResult({ ...r, testedAt: new Date().toLocaleString() })
      if (r.success) {
        toast({ title: "连接成功", description: r.message, duration: 8000 })
      } else {
        toast({ variant: "destructive", title: "连接失败", description: r.message, duration: 10000 })
      }
    } catch (e: any) {
      const message = String(e)
      setTestResult({ success: false, message, testedAt: new Date().toLocaleString() })
      toast({ variant: "destructive", title: "测试出错", description: message })
    } finally {
      setIsTesting(false)
    }
  }

  const handleReset = () => {
    setTestResult(null)
    if (!originalProvider) {
      setConfig({ provider: "", runtimeProvider: "", model: "", apiKey: "", baseUrl: "", endpoint: "chat", customEndpoint: "" })
      return
    }
    setConfig(c => ({ ...c, model: "", apiKey: "" }))
  }

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64 text-muted-foreground">
        <Loader2 className="mr-2 h-4 w-4 animate-spin" /> 加载配置中...
      </div>
    )
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: "easeOut" }}
    >
      <div className="max-w-2xl mx-auto w-full">
        {/* AI 模型配置 Card */}
        <Card>
          <CardHeader className="px-4 sm:px-6">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-primary/10">
                <Bot className="h-5 w-5 sm:h-6 sm:w-6 text-primary" />
              </div>
              <div>
                <CardTitle className="text-lg sm:text-xl">AI 模型配置</CardTitle>
                <CardDescription className="text-xs sm:text-sm">配置用于自动答题的 AI 模型和凭证</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-4 sm:space-y-6 px-4 sm:px-6">
            {/* AI Provider Selection */}
            <div className="space-y-2">
              <Label htmlFor="ai-provider" className="text-sm sm:text-base font-medium">
                AI 提供商 <span className="text-red-500">*</span>
              </Label>
              <Select value={config.provider} onValueChange={onProviderChange}>
                <SelectTrigger id="ai-provider" className="w-full">
                  <SelectValue placeholder="请选择 AI 提供商" />
                </SelectTrigger>
                <SelectContent>
                  {PROVIDER_PRESETS.map(p => (
                    <SelectItem key={p.value} value={p.value}>{p.label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {providerPreset?.hint && (
                <p className="text-xs sm:text-sm text-muted-foreground">{providerPreset.hint}</p>
              )}
            </div>

            {/* Model Input */}
            <div className="space-y-2">
              <Label htmlFor="model" className="text-sm sm:text-base font-medium">
                模型名称 <span className="text-red-500">*</span>
              </Label>
              <Input
                id="model"
                placeholder={providerPreset?.defaultModel || "例如: gpt-4o-mini, qwen-plus, deepseek-chat"}
                value={config.model}
                onChange={(e) => setConfig({ ...config, model: e.target.value })}
                className="text-sm sm:text-base"
              />
              <p className="text-xs sm:text-sm text-muted-foreground">
                {providerPreset?.defaultModel ? `默认建议: ${providerPreset.defaultModel}` : "请输入目标模型 ID"}
              </p>
              {suggestedModels.length > 0 && (
                <div className="flex flex-wrap gap-2">
                  {suggestedModels.map(model => (
                    <Button key={model} type="button" variant="outline" size="sm" className="h-7 text-xs" onClick={() => setConfig({ ...config, model })}>
                      {model}
                    </Button>
                  ))}
                </div>
              )}
            </div>

            {/* API Key */}
            <div className="space-y-2">
              <Label htmlFor="api-key" className="text-sm sm:text-base font-medium">
                API 密钥 <span className="text-red-500">*</span>
              </Label>
              <div className="relative">
                <Input
                  id="api-key"
                  type={showApiKey ? "text" : "password"}
                  placeholder="请输入 API 密钥 (sk-...)"
                  value={config.apiKey}
                  onChange={(e) => setConfig({ ...config, apiKey: e.target.value })}
                  className="pr-10 text-sm sm:text-base"
                />
                <button
                  type="button"
                  onClick={() => setShowApiKey(!showApiKey)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                  aria-label={showApiKey ? "隐藏密钥" : "显示密钥"}
                >
                  {showApiKey ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
              <p className="text-xs sm:text-sm text-muted-foreground">密钥将存储在 config.yaml 中，页面不会在测试结果中展示密钥</p>
            </div>

            {/* Base URL */}
            <div className="space-y-2">
              <Label htmlFor="base-url" className="text-sm sm:text-base font-medium">
                API 基础 URL <span className="text-red-500">*</span>
              </Label>
              <Input
                id="base-url"
                placeholder="https://api.openai.com"
                value={config.baseUrl}
                onChange={(e) => setConfig({ ...config, baseUrl: e.target.value })}
                className="text-sm sm:text-base font-mono"
              />
              <p className="text-xs sm:text-sm text-muted-foreground">可包含网关前缀，但不要包含最终接口路径；系统会自动拼接端点</p>
            </div>

            {/* Endpoint Mode */}
            <div className="space-y-2">
              <Label htmlFor="endpoint" className="text-sm sm:text-base font-medium">
                端点模式
              </Label>
              <Select value={config.endpoint} onValueChange={(v) => setConfig({ ...config, endpoint: v })}>
                <SelectTrigger id="endpoint" className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="chat">Chat Completions (/v1/chat/completions) - OpenAI 兼容</SelectItem>
                  <SelectItem value="responses">Responses (/v1/responses) - OpenAI 新 API</SelectItem>
                  <SelectItem value="custom">自定义路径 (Custom)</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs sm:text-sm text-muted-foreground">
                实际请求 URL = <code className="text-xs bg-muted px-1 rounded break-all">{currentUrl || "<请填写>"}</code>
              </p>
            </div>

            {/* Custom Endpoint Path */}
            {showCustomEndpoint && (
              <div className="space-y-2">
                <Label htmlFor="custom-endpoint" className="text-sm sm:text-base font-medium">
                  自定义端点路径 <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="custom-endpoint"
                  placeholder="例如: api/v1/chat/completions"
                  value={config.customEndpoint || ""}
                  onChange={(e) => setConfig({ ...config, customEndpoint: e.target.value })}
                  className="text-sm sm:text-base font-mono"
                />
                <p className="text-xs sm:text-sm text-muted-foreground">不含域名，可填写多级路径，开头的 / 会自动处理</p>
              </div>
            )}

            {/* Configuration Preview */}
            {isFormValid && (
              <div className="p-3 sm:p-4 rounded-lg bg-muted/50 border border-border">
                <h4 className="text-xs sm:text-sm font-medium mb-2 text-foreground flex items-center gap-2">
                  <ExternalLink className="h-3.5 w-3.5" /> 当前配置
                </h4>
                <div className="space-y-1 text-xs sm:text-sm text-muted-foreground">
                  <div className="flex justify-between gap-4">
                    <span>提供商:</span>
                    <span className="font-medium text-foreground">{providerPreset?.label || config.provider}</span>
                  </div>
                  <div className="flex justify-between gap-4">
                    <span>运行时类型:</span>
                    <span className="font-medium text-foreground">{config.runtimeProvider || providerPreset?.runtimeProvider || config.provider}</span>
                  </div>
                  <div className="flex justify-between gap-4">
                    <span>模型:</span>
                    <span className="font-medium text-foreground break-all">{config.model}</span>
                  </div>
                  <div className="flex justify-between gap-4">
                    <span>完整 URL:</span>
                    <span className="font-mono text-foreground break-all text-right">{currentUrl}</span>
                  </div>
                  <div className="flex justify-between gap-4">
                    <span>API 密钥:</span>
                    <span className="font-mono text-foreground">
                      {config.apiKey ? "••••••••" + config.apiKey.slice(-4) : "未设置"}
                    </span>
                  </div>
                </div>
              </div>
            )}

            {/* Save / Test / Reset Buttons */}
            <div className="flex flex-col sm:flex-row gap-3 pt-4">
              <Button onClick={handleSave} disabled={!isFormValid || isSaving} className="flex-1 gap-2">
                {isSaving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
                {isSaving ? "保存中..." : "保存配置"}
              </Button>
              <Button onClick={handleTest} disabled={!isFormValid || isTesting} variant="secondary" className="flex-1 gap-2">
                {isTesting ? <Loader2 className="h-4 w-4 animate-spin" /> : <TestTube className="h-4 w-4" />}
                {isTesting ? "测试中..." : "测试连通性"}
              </Button>
              <Button variant="outline" onClick={handleReset} className="sm:w-auto gap-2">
                <RotateCcw className="h-4 w-4" /> 重置
              </Button>
            </div>

            {testResult && (
              <div className={`p-3 sm:p-4 rounded-lg border ${testResult.success ? "bg-green-50 border-green-200 text-green-900" : "bg-red-50 border-red-200 text-red-900"}`}>
                <h4 className="text-sm font-medium mb-2 flex items-center gap-2">
                  {testResult.success ? <CheckCircle2 className="h-4 w-4" /> : <XCircle className="h-4 w-4" />}
                  {testResult.success ? "AI 连通性正常" : "AI 连通性异常"}
                </h4>
                <div className="space-y-1 text-xs sm:text-sm">
                  <div>结果：{testResult.message}</div>
                  {testResult.statusCode !== undefined && <div>HTTP 状态：{testResult.statusCode}</div>}
                  {testResult.durationMs !== undefined && <div>耗时：{testResult.durationMs} ms</div>}
                  {testResult.runtimeProvider && <div>运行时类型：{testResult.runtimeProvider}</div>}
                  {testResult.url && <div className="break-all">测试 URL：{testResult.url}</div>}
                  <div>测试时间：{testResult.testedAt}</div>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </motion.div>
  )
}