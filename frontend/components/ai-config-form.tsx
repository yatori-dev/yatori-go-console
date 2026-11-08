"use client"

import { useEffect, useMemo, useState } from "react"
import { motion } from "framer-motion"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Bot, Database, Eye, EyeOff, Loader2, Save } from "lucide-react"
import { getSetting, normalizeSetting, saveSetting, type Setting } from "@/api/settingApi"
import { toast } from "@/components/ui/use-toast"

const AI_TYPES = [
    { value: "SILICON", label: "SiliconFlow（硅基流动）" },
    { value: "DEEPSEEK", label: "DeepSeek（深度求索）" },
    { value: "CHATGLM", label: "智谱 GLM（ChatGLM）" },
    { value: "TONGYI", label: "通义千问" },
    { value: "XINGHUO", label: "讯飞星火" },
    { value: "DOUBAO", label: "豆包" },
    { value: "METAAI", label: "Meta AI" },
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
                toast({ title: "加载失败", description: response.message, variant: "destructive" })
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
                toast({ title: "保存成功", description: "配置已写入本地数据库" })
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
            <div className="mx-auto w-full max-w-3xl space-y-5">
                {Array.from({ length: 2 }).map((_, index) => (
                    <div key={index} className="surface h-56 animate-pulse p-6">
                        <div className="h-4 w-1/3 rounded bg-muted" />
                        <div className="mt-6 space-y-3">
                            <div className="h-9 w-full rounded bg-muted" />
                            <div className="h-9 w-full rounded bg-muted" />
                            <div className="h-9 w-2/3 rounded bg-muted" />
                        </div>
                    </div>
                ))}
            </div>
        )
    }

    return (
        <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3, ease: "easeOut" }}
            className="mx-auto w-full max-w-3xl space-y-5"
        >
            <Card className="overflow-hidden">
                <CardHeader className="pb-4">
                    <div className="flex items-start gap-3">
                        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
                            <Bot className="h-4 w-4" />
                        </div>
                        <div>
                            <CardTitle className="text-base">AI 答题模型</CardTitle>
                            <CardDescription className="text-xs">
                                账号选择「AI 大模型答题」时使用这套配置
                            </CardDescription>
                        </div>
                    </div>
                </CardHeader>
                <CardContent className="space-y-5">
                    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                        <div className="space-y-2">
                            <Label>服务商</Label>
                            <Select
                                value={setting.aiSetting.aiType}
                                onValueChange={(value) => updateAI({ aiType: value })}
                            >
                                <SelectTrigger className="w-full">
                                    <SelectValue placeholder="选择 AI 服务商" />
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
                    </div>

                    {setting.aiSetting.aiType === "OTHER" && (
                        <div className="space-y-2">
                            <Label htmlFor="ai-url">接口地址</Label>
                            <Input
                                id="ai-url"
                                placeholder="https://api.example.com/v1/chat/completions"
                                value={setting.aiSetting.aiUrl}
                                onChange={(e) => updateAI({ aiUrl: e.target.value })}
                            />
                            <p className="text-xs text-muted-foreground">需要兼容 OpenAI Chat Completions 协议</p>
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
                                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
                                aria-label={showApiKey ? "隐藏密钥" : "显示密钥"}
                            >
                                {showApiKey ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                            </button>
                        </div>
                        <p className="text-xs text-muted-foreground">{currentTypeLabel} 的 API Key，仅保存在本地数据库</p>
                    </div>
                </CardContent>
            </Card>

            <Card className="overflow-hidden">
                <CardHeader className="pb-4">
                    <div className="flex items-start gap-3">
                        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
                            <Database className="h-4 w-4" />
                        </div>
                        <div>
                            <CardTitle className="text-base">外部题库</CardTitle>
                            <CardDescription className="text-xs">账号选择「外置题库答题」时使用</CardDescription>
                        </div>
                    </div>
                </CardHeader>
                <CardContent className="space-y-3">
                    <div className="space-y-2">
                        <Label htmlFor="external-bank-url">题库服务地址</Label>
                        <Input
                            id="external-bank-url"
                            type="url"
                            placeholder="http://localhost:8083"
                            value={setting.apiQueSetting.url}
                            onChange={(e) => updateBank(e.target.value)}
                        />
                    </div>
                    <p className="text-xs text-muted-foreground">
                        留空表示不使用外置题库；接口规范见 yatori 文档 bank-interface-api。
                    </p>
                </CardContent>
            </Card>

            <div className="flex flex-col gap-3 rounded-xl border border-border/80 bg-muted/40 px-4 py-3 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
                <div className="space-y-1">
                    <p>海旗科技选择「内置标答答题」时不需要 AI 配置。</p>
                    <p>配置保存后对之后启动的任务生效，运行中的任务不受影响。</p>
                </div>
                <Button onClick={handleSave} disabled={saving} className="shrink-0 gap-1.5">
                    {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
                    {saving ? "保存中..." : "保存配置"}
                </Button>
            </div>
        </motion.div>
    )
}
