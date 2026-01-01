"use client"

import { useState } from "react"
import { motion } from "framer-motion"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Eye, EyeOff, Save, Bot } from "lucide-react"

export type AIConfig = {
  provider: string
  model: string
  apiKey: string
}

export function AIConfigForm() {
  const [config, setConfig] = useState<AIConfig>({
    provider: "",
    model: "",
    apiKey: "",
  })
  const [externalBankUrl, setExternalBankUrl] = useState("")
  const [showApiKey, setShowApiKey] = useState(false)
  const [isSaved, setIsSaved] = useState(false)
  const [isBankSaved, setIsBankSaved] = useState(false)

  const handleSave = () => {
    console.log("Saving AI config:", config)
    setIsSaved(true)
    setTimeout(() => setIsSaved(false), 2000)
  }

  const handleSaveBank = () => {
    if (!externalBankUrl) return
    console.log("Saving external bank config:", { externalBankUrl })
    setIsBankSaved(true)
    setTimeout(() => setIsBankSaved(false), 2000)
  }

  const isFormValid = config.provider && config.model && config.apiKey
  const isBankFormValid = externalBankUrl

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: "easeOut" }}
    >
    <div className="max-w-2xl mx-auto w-full">
      <Card>
        <CardHeader className="px-4 sm:px-6">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-primary/10">
              <Bot className="h-5 w-5 sm:h-6 sm:w-6 text-primary" />
            </div>
            <div>
              <CardTitle className="text-lg sm:text-xl">AI模型配置</CardTitle>
              <CardDescription className="text-xs sm:text-sm">配置用于自动答题的AI模型和凭证</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4 sm:space-y-6 px-4 sm:px-6">
          {/* AI Provider Selection */}
          <div className="space-y-2">
            <Label htmlFor="ai-provider" className="text-sm sm:text-base font-medium">
              AI提供商 <span className="text-red-500">*</span>
            </Label>
            <Select value={config.provider} onValueChange={(value) => setConfig({ ...config, provider: value })}>
              <SelectTrigger id="ai-provider" className="w-full">
                <SelectValue placeholder="请选择AI提供商" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="tongyi">通义千问</SelectItem>
                <SelectItem value="deepseek">Deepseek</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-xs sm:text-sm text-muted-foreground">选择要使用的AI服务提供商</p>
          </div>

          {/* Model Input */}
          <div className="space-y-2">
            <Label htmlFor="model" className="text-sm sm:text-base font-medium">
              模型名称 <span className="text-red-500">*</span>
            </Label>
            <Input
              id="model"
              placeholder="例如: qwen-max, deepseek-chat"
              value={config.model}
              onChange={(e) => setConfig({ ...config, model: e.target.value })}
              className="text-sm sm:text-base"
            />
            <p className="text-xs sm:text-sm text-muted-foreground">
              {config.provider === "tongyi"
                ? "例如: qwen-max, qwen-plus, qwen-turbo"
                : config.provider === "deepseek"
                  ? "例如: deepseek-chat, deepseek-coder"
                  : "请输入要使用的模型名称"}
            </p>
          </div>

          {/* API Key Input with Toggle Visibility */}
          <div className="space-y-2">
            <Label htmlFor="api-key" className="text-sm sm:text-base font-medium">
              API密钥 <span className="text-red-500">*</span>
            </Label>
            <div className="relative">
              <Input
                id="api-key"
                type={showApiKey ? "text" : "password"}
                placeholder="请输入API密钥"
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
            <p className="text-xs sm:text-sm text-muted-foreground">API密钥将被安全存储，用于调用AI服务</p>
          </div>

          {/* Configuration Preview */}
          {isFormValid && (
            <div className="p-3 sm:p-4 rounded-lg bg-muted/50 border border-border">
              <h4 className="text-xs sm:text-sm font-medium mb-2 text-foreground">当前配置</h4>
              <div className="space-y-1 text-xs sm:text-sm text-muted-foreground">
                <div className="flex justify-between gap-4">
                  <span>AI提供商:</span>
                  <span className="font-medium text-foreground">
                    {config.provider === "tongyi" ? "通义千问" : "Deepseek"}
                  </span>
                </div>
                <div className="flex justify-between gap-4">
                  <span>模型:</span>
                  <span className="font-medium text-foreground break-all">{config.model}</span>
                </div>
                <div className="flex justify-between gap-4">
                  <span>API密钥:</span>
                  <span className="font-mono text-foreground">
                    {config.apiKey ? "••••••••" + config.apiKey.slice(-4) : "未设置"}
                  </span>
                </div>
              </div>
            </div>
          )}

          {/* Save Button */}
          <div className="flex flex-col sm:flex-row gap-3 pt-4">
            <Button onClick={handleSave} disabled={!isFormValid} className="flex-1 gap-2">
              <Save className="h-4 w-4" />
              {isSaved ? "保存成功" : "保存配置"}
            </Button>
            <Button
              variant="outline"
              onClick={() => setConfig({ provider: "", model: "", apiKey: "" })}
              disabled={!isFormValid}
              className="sm:w-auto"
            >
              重置
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* 外部题库配置 */}
      <Card className="mt-4 sm:mt-6">
        <CardHeader className="px-4 sm:px-6">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-primary/10">
              <Bot className="h-5 w-5 sm:h-6 sm:w-6 text-primary" />
            </div>
            <div>
              <CardTitle className="text-lg sm:text-xl">外部题库配置</CardTitle>
              <CardDescription className="text-xs sm:text-sm">配置用于自动考试的外部题库URL</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4 sm:space-y-6 px-4 sm:px-6">
          {/* External Question Bank URL */}
          <div className="space-y-2">
            <Label htmlFor="external-bank-url" className="text-sm sm:text-base font-medium">
              外部题库URL <span className="text-red-500">*</span>
            </Label>
            <Input
              id="external-bank-url"
              type="url"
              placeholder="https://example.com/api/questions"
              value={externalBankUrl}
              onChange={(e) => setExternalBankUrl(e.target.value)}
              className="text-sm sm:text-base"
            />
            <p className="text-xs sm:text-sm text-muted-foreground">
              外部题库API地址，用于自动考试时获取题目和答案
            </p>
          </div>

          {/* Save Button */}
          <div className="flex flex-col sm:flex-row gap-3 pt-4">
            <Button onClick={handleSaveBank} disabled={!isBankFormValid} className="flex-1 gap-2">
              <Save className="h-4 w-4" />
              {isBankSaved ? "保存成功" : "保存题库配置"}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Help Information */}
      <Card className="mt-4 sm:mt-6">
        <CardHeader className="px-4 sm:px-6">
          <CardTitle className="text-sm sm:text-base">如何获取API密钥？</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 text-xs sm:text-sm text-muted-foreground px-4 sm:px-6">
          <div>
            <h5 className="font-medium text-foreground mb-1">通义千问</h5>
            <p>访问阿里云控制台，在通义千问服务页面创建API Key</p>
          </div>
          <div>
            <h5 className="font-medium text-foreground mb-1">Deepseek</h5>
            <p>访问 Deepseek 官网，注册账号后在API管理页面获取密钥</p>
          </div>
        </CardContent>
      </Card>
    </div>
    </motion.div>
  )
}
