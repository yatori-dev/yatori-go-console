"use client"

import type React from "react"
import { useMemo, useState } from "react"
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Eye, EyeOff } from "lucide-react"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { addAccount, defaultAccountConfig, type AccountConfig } from "@/api/accountApi"
import { toast } from "@/components/ui/use-toast"
import {
    PLATFORMS,
    getAnswerModeDescription,
    getAnswerModeLabel,
    getVideoModeDescription,
    getVideoModeLabel,
    isConcurrentVideoMode,
    platformRequiresUrl,
} from "@/utils/platformUtils"

type AddAccountDialogProps = {
    open: boolean
    onOpenChange: (open: boolean) => void
    onAdded: () => void
}

export function AddAccountDialog({ open, onOpenChange, onAdded }: AddAccountDialogProps) {
    const [config, setConfig] = useState<AccountConfig>(defaultAccountConfig())
    const [showPassword, setShowPassword] = useState(false)
    const [loading, setLoading] = useState(false)

    const platform = useMemo(
        () => PLATFORMS.find((item) => item.code === config.accountType),
        [config.accountType],
    )
    const requiresUrl = platformRequiresUrl(config.accountType)

    const updateConfig = (patch: Partial<AccountConfig>) => {
        setConfig((prev) => ({ ...prev, ...patch }))
    }

    const updateCoursesCustom = (patch: Partial<AccountConfig["coursesCustom"]>) => {
        setConfig((prev) => ({
            ...prev,
            coursesCustom: { ...prev.coursesCustom, ...patch },
        }))
    }

    const handlePlatformChange = (value: string) => {
        const next = PLATFORMS.find((item) => item.code === value)
        const videoModel = next?.videoModes.includes(config.coursesCustom.videoModel)
            ? config.coursesCustom.videoModel
            : 1
        const autoExam = next?.answerModes.includes(config.coursesCustom.autoExam)
            ? config.coursesCustom.autoExam
            : 0
        setConfig((prev) => ({
            ...prev,
            accountType: value,
            url: next?.requiresUrl ? prev.url : "",
            coursesCustom: { ...prev.coursesCustom, videoModel, autoExam },
        }))
    }

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()

        if (!config.accountType || !config.account.trim() || !config.password.trim()) {
            toast({ title: "表单验证失败", description: "请填写平台、账号和密码", variant: "destructive" })
            return
        }
        if (requiresUrl && !config.url.trim()) {
            toast({ title: "表单验证失败", description: "该平台必须填写 URL", variant: "destructive" })
            return
        }

        setLoading(true)
        try {
            const response = await addAccount(config)
            if (response.code === 200) {
                toast({ title: "添加成功", description: response.message || "账号已成功添加" })
                setConfig(defaultAccountConfig())
                setShowPassword(false)
                onOpenChange(false)
                onAdded()
            } else {
                toast({
                    title: "添加失败",
                    description: response.message || "添加账号失败",
                    variant: "destructive",
                })
            }
        } finally {
            setLoading(false)
        }
    }

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>添加新账号</DialogTitle>
                    <DialogDescription>
                        账号会保存到本地数据库，后续不再通过 config.yaml 维护账号
                    </DialogDescription>
                </DialogHeader>
                <form onSubmit={handleSubmit}>
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label>平台</Label>
                            <Select value={config.accountType} onValueChange={handlePlatformChange}>
                                <SelectTrigger className="w-full">
                                    <SelectValue placeholder="请选择账号平台" />
                                </SelectTrigger>
                                <SelectContent>
                                    {PLATFORMS.map((item) => (
                                        <SelectItem key={item.code} value={item.code}>
                                            {item.name}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>

                        {requiresUrl && (
                            <div className="space-y-2">
                                <Label htmlFor="url">平台 URL</Label>
                                <Input
                                    id="url"
                                    placeholder="例如 https://xxx.haiqikeji.com"
                                    value={config.url}
                                    onChange={(e) => updateConfig({ url: e.target.value })}
                                    disabled={loading}
                                />
                            </div>
                        )}

                        <div className="space-y-2">
                            <Label htmlFor="account">账号</Label>
                            <Input
                                id="account"
                                placeholder="请输入登录账号"
                                value={config.account}
                                onChange={(e) => updateConfig({ account: e.target.value })}
                                disabled={loading}
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="password">密码 / Cookie / Token</Label>
                            <div className="relative">
                                <Input
                                    id="password"
                                    type={showPassword ? "text" : "password"}
                                    placeholder="请输入密码、Cookie 或 Token"
                                    value={config.password}
                                    onChange={(e) => updateConfig({ password: e.target.value })}
                                    className="pr-10"
                                    disabled={loading}
                                />
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="absolute right-0 top-0 h-full w-10 hover:bg-transparent"
                                    onClick={() => setShowPassword(!showPassword)}
                                    disabled={loading}
                                >
                                    {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                                </Button>
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="remarkName">备注名（可选）</Label>
                            <Input
                                id="remarkName"
                                placeholder="用于区分多个账号"
                                value={config.remarkName || ""}
                                onChange={(e) => updateConfig({ remarkName: e.target.value })}
                                disabled={loading}
                            />
                        </div>

                        <div className="flex items-center justify-between rounded-lg border p-3">
                            <div>
                                <Label className="text-sm">开启 IP 代理</Label>
                                <p className="text-xs text-muted-foreground mt-1">
                                    开启后从 ip.txt 中随机选择代理
                                </p>
                            </div>
                            <Switch
                                checked={config.isProxy === 1}
                                onCheckedChange={(checked) => updateConfig({ isProxy: checked ? 1 : 0 })}
                                disabled={loading}
                            />
                        </div>

                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                            <div className="space-y-2">
                                <Label>视频模式</Label>
                                <Select
                                    value={String(config.coursesCustom.videoModel)}
                                    onValueChange={(value) => updateCoursesCustom({ videoModel: Number(value) })}
                                    disabled={loading}
                                >
                                    <SelectTrigger>
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {(platform?.videoModes || [0, 1, 2]).map((mode) => (
                                            <SelectItem key={mode} value={String(mode)}>
                                                {getVideoModeLabel(config.accountType, mode)}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                                {getVideoModeDescription(config.accountType, config.coursesCustom.videoModel) && (
                                    <p className="text-xs text-muted-foreground">
                                        {getVideoModeDescription(config.accountType, config.coursesCustom.videoModel)}
                                    </p>
                                )}
                                {isConcurrentVideoMode(config.accountType, config.coursesCustom.videoModel) && (
                                    <div className="space-y-1 pt-1">
                                        <Label className="text-xs">同时任务点数（1-5）</Label>
                                        <Input
                                            type="number"
                                            min={1}
                                            max={5}
                                            value={config.coursesCustom.cxNode ?? 3}
                                            onChange={(e) =>
                                                updateCoursesCustom({
                                                    cxNode: Math.max(1, Math.min(5, Number(e.target.value) || 1)),
                                                })
                                            }
                                            disabled={loading}
                                        />
                                    </div>
                                )}
                            </div>
                            <div className="space-y-2">
                                <Label>答题模式</Label>
                                <Select
                                    value={String(config.coursesCustom.autoExam)}
                                    onValueChange={(value) => updateCoursesCustom({ autoExam: Number(value) })}
                                    disabled={loading}
                                >
                                    <SelectTrigger>
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {(platform?.answerModes || [0]).map((mode) => (
                                            <SelectItem key={mode} value={String(mode)}>
                                                {getAnswerModeLabel(config.accountType, mode)}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                                {config.coursesCustom.autoExam !== 0 && (
                                    <p className="text-xs text-muted-foreground">
                                        {getAnswerModeDescription(config.accountType, config.coursesCustom.autoExam)}
                                    </p>
                                )}
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label>交卷方式</Label>
                            <Select
                                value={String(config.coursesCustom.examAutoSubmit)}
                                onValueChange={(value) => updateCoursesCustom({ examAutoSubmit: Number(value) })}
                                disabled={loading}
                            >
                                <SelectTrigger>
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="0">只保存不交卷</SelectItem>
                                    <SelectItem value="1">答完自动交卷</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={loading}>
                            取消
                        </Button>
                        <Button type="submit" disabled={loading}>
                            {loading ? "添加中..." : "添加账号"}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    )
}
