"use client"

import { useEffect, useMemo, useState } from "react"
import { motion } from "framer-motion"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Eye, EyeOff, Save, Bot, Loader2 } from "lucide-react"
import { getSetting, normalizeSetting, saveSetting, type Setting } from "@/api/settingApi"
import { toast } from "@/components/ui/use-toast"

const AI_TYPES = [
    { value: "SILICON", label: "SiliconFlow（硅基流动）" },
    { value: "DEEPSEEK", label: "DeepSeek（深度求索）" },
    { value: "CHATGLM", label: "智谱清言（ChatGLM）" },
    { value: "TONGYI", label: "通义千问" },
    { value: "XINGHUO", label: "星火大模型" },
    { value: "DOUBAO", label: "豆包大模型" },
    { value: "METAAI", label: "秘塔 AI" },
    { value: "OPENAI", label: "OpenAI / ChatGPT" },
    { value: "OTHER", label: "其他 OpenAI 兼容接口" },
]

export function AIConfigForm() {
    const [setting, setSetting] = useState<Setting>(normalizeSetting(null))
    const [showApiKey, setShowApiKey] = useState(false)
    const [loading, setLoading] = useState(true)
    const [saving, setSaving] = useState(false)

    useEffect(() => {
        const load = async () => {
            const response = await getSetting()
            if (response.code === 200 && response.data) {
                setSetting(normalizeSetting(response.data.setting))
            } else {
                toast({ title: "读取失败", description: response.message, variant: "destructive" })
            }
            setLoading(false)
        }
        load()
    }, [])

    const updateAI = (patch: Partial<Setting["aiSetting"]>) => {
        setSetting((prev) => ({ ...prev, aiSetting: { ...prev.aiSetting, ...patch } }))
    }

    const updateBank = (url: string) => {
        setSetting((prev) => ({ ...prev, apiQueSetting: { ...prev.apiQueSetting, url } }))
    }

    const handleSave = async () => {
        setSaving(true)
        try {
            const response = await saveSetting(setting)
            if (response.code === 200) {
                toast({ title: "保存成功", description: "全局设置已写入本地数据库" })
            } else {
                toast({ title: "保存失败", description: response.message, variant: "destructive" })
            }
        } finally {
            setSaving(false)
        }
    }

    const currentTypeLabel = useMemo(() => {
        return AI_TYPES.find((item) => item.value === setting.aiSetting.aiType)?.label || setting.aiSetting.aiType
    }, [setting.aiSetting.aiType])

    if (loading) {
        return (
            <div className="flex items-center justify-center py-16">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
                <span className="ml-3 text-muted-foreground">加载设置中...</span>
            </div>
        )
    }

    return (
        <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4, ease: "easeOut" }}
        >
            <div className="max-w-2xl mx-auto w-full">
                <Card>
                    <CardHeader className="px-4 sm:px-6">
                        <div className="flex items-center gap-3">
                            <div className="p-2 rounded-lg bg-primary/10">
                                <Bot className="h-5 w-5 sm:h-6 sm:w-6 text-primary" />
                            </div>
                            <div>
                                <CardTitle className="text-lg sm:text-xl">AI 模型配置</CardTitle>
                                <CardDescription className="text-xs sm:text-sm">
                                    全局 AI 配置，账号选择“AI 答题”时生效
                                </CardDescription>
                            </div>
                        </div>
                    </CardHeader>
                    <CardContent className="space-y-4 sm:space-y-6 px-4 sm:px-6">
                        <div className="space-y-2">
                            <Label>AI 提供商</Label>
                            <Select
                                value={setting.aiSetting.aiType}
                                onValueChange={(value) => updateAI({ aiType: value })}
                            >
                                <SelectTrigger className="w-full">
                                    <SelectValue placeholder="请选择 AI 提供商" />
                                </SelectTrigger>
                                <SelectContent>
                                    {AI_TYPES.map((item) => (
                                        <SelectItem key={item.value} value={item.value}>
                                            {item.label}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="ai-model">模型名称</Label>
                            <Input
                                id="ai-model"
                                placeholder="例如 qwen-plus、deepseek-chat"
                                value={setting.aiSetting.model}
                                onChange={(e) => updateAI({ model: e.target.value })}
                            />
                        </div>

                        {setting.aiSetting.aiType === "OTHER" && (
                            <div className="space-y-2">
                                <Label htmlFor="ai-url">AI 接口地址</Label>
                                <Input
                                    id="ai-url"
                                    placeholder="https://api.example.com/v1/chat/completions"
                                    value={setting.aiSetting.aiUrl}
                                    onChange={(e) => updateAI({ aiUrl: e.target.value })}
                                />
                            </div>
                        )}

                        <div className="space-y-2">
                            <Label htmlFor="ai-key">API Key</Label>
                            <div className="relative">
                                <Input
                                    id="ai-key"
                                    type={showApiKey ? "text" : "password"}
                                    placeholder="请输入 API Key"
                                    value={setting.aiSetting.API_KEY}
                                    onChange={(e) => updateAI({ API_KEY: e.target.value })}
                                    className="pr-10"
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowApiKey(!showApiKey)}
                                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                                    aria-label={showApiKey ? "隐藏密钥" : "显示密钥"}
                                >
                                    {showApiKey ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                                </button>
                            </div>
                            <p className="text-xs text-muted-foreground">{currentTypeLabel} 的 API Key</p>
                        </div>

                        <div className="p-3 rounded-lg bg-muted/50 border text-xs text-muted-foreground">
                            AI 配置会保存到本地 SQLite，与账号配置一起脱离 config.yaml。
                        </div>

                        <Button onClick={handleSave} disabled={saving} className="w-full gap-2">
                            {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
                            {saving ? "保存中..." : "保存 AI 配置"}
                        </Button>
                    </CardContent>
                </Card>

                <Card className="mt-4 sm:mt-6">
                    <CardHeader className="px-4 sm:px-6">
                        <div className="flex items-center gap-3">
                            <div className="p-2 rounded-lg bg-primary/10">
                                <Bot className="h-5 w-5 sm:h-6 sm:w-6 text-primary" />
                            </div>
                            <div>
                                <CardTitle className="text-lg sm:text-xl">外部题库配置</CardTitle>
                                <CardDescription className="text-xs sm:text-sm">
                                    账号选择“外置题库答题”时使用
                                </CardDescription>
                            </div>
                        </div>
                    </CardHeader>
                    <CardContent className="space-y-4 px-4 sm:px-6">
                        <div className="space-y-2">
                            <Label htmlFor="external-bank-url">外部题库 URL</Label>
                            <Input
                                id="external-bank-url"
                                type="url"
                                placeholder="http://localhost:8083"
                                value={setting.apiQueSetting.url}
                                onChange={(e) => updateBank(e.target.value)}
                            />
                            <p className="text-xs text-muted-foreground">
                                题库接口规范见文档：bank-interface-api
                            </p>
                        </div>
                        <Button onClick={handleSave} disabled={saving} className="w-full gap-2">
                            {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
                            {saving ? "保存中..." : "保存题库配置"}
                        </Button>
                    </CardContent>
                </Card>

                <Card className="mt-4 sm:mt-6">
                    <CardHeader className="px-4 sm:px-6">
                        <CardTitle className="text-sm sm:text-base">说明</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-3 text-xs sm:text-sm text-muted-foreground px-4 sm:px-6">
                        <p>海旗科技选择“内置标答答题”时不需要 AI 配置。</p>
                        <p>学习通内置 AI 目前按普通 AI 答题处理。</p>
                        <p>设置保存后对之后启动的任务生效，运行中的任务不受影响。</p>
                    </CardContent>
                </Card>
            </div>
        </motion.div>
    )
}
