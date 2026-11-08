"use client"

import type React from "react"
import { useEffect, useState } from "react"
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
import { Eye, EyeOff } from "lucide-react"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import {
    addAccount,
    defaultAccountConfig,
    normalizeConfig,
    updateAccount,
    type Account,
    type AccountConfig,
} from "@/api/accountApi"
import { toast } from "@/components/ui/use-toast"
import { PLATFORMS, platformRequiresUrl } from "@/utils/platformUtils"

type AccountFormDialogProps = {
    open: boolean
    onOpenChange: (open: boolean) => void
    account?: Account | null
    onSaved: () => void
}

export function AccountFormDialog({ open, onOpenChange, account, onSaved }: AccountFormDialogProps) {
    const isEdit = Boolean(account)
    const [config, setConfig] = useState<AccountConfig>(() =>
        account ? normalizeConfig(account.config) : defaultAccountConfig(),
    )
    const [showPassword, setShowPassword] = useState(false)
    const [loading, setLoading] = useState(false)

    // 打开弹窗时按当前账号重置表单；编辑过程中轮询刷新不会覆盖输入
    useEffect(() => {
        if (!open) {
            return
        }
        setConfig(account ? normalizeConfig(account.config) : defaultAccountConfig())
        setShowPassword(false)
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [open, account?.uid])

    const requiresUrl = platformRequiresUrl(config.accountType)

    const updateConfig = (patch: Partial<AccountConfig>) => {
        setConfig((prev) => ({ ...prev, ...patch }))
    }

    const handlePlatformChange = (value: string) => {
        setConfig((prev) => ({
            ...prev,
            accountType: value,
            loginMode: value === "XUEXITONG" ? prev.loginMode || "web" : "web",
            url: platformRequiresUrl(value) ? prev.url : "",
        }))
    }

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()

        if (!config.accountType || !config.account.trim() || !config.password.trim()) {
            toast({ title: "无法保存", description: "请选择平台并填写账号和密码", variant: "destructive" })
            return
        }
        if (requiresUrl && !config.url.trim()) {
            toast({ title: "无法保存", description: "该平台需要填写平台 URL", variant: "destructive" })
            return
        }

        setLoading(true)
        try {
            const response = account ? await updateAccount(account.uid, config) : await addAccount(config)
            if (response.code === 200) {
                toast({
                    title: isEdit ? "保存成功" : "添加成功",
                    description: isEdit ? "账号配置已更新" : response.message || "账号已保存",
                })
                onOpenChange(false)
                onSaved()
            } else {
                toast({
                    title: isEdit ? "保存失败" : "添加失败",
                    description: response.message || "操作失败",
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
                    <DialogTitle>{isEdit ? "编辑账号" : "添加新账号"}</DialogTitle>
                    <DialogDescription>
                        {isEdit
                            ? "修改账号、密码或登录方式；其他刷课配置请在账号详情里调整"
                            : "先填写平台和登录信息，其他配置（IP 代理、视频模式、答题模式等）添加后在账号详情里调整"}
                    </DialogDescription>
                </DialogHeader>
                <form onSubmit={handleSubmit}>
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label>平台</Label>
                            <Select value={config.accountType} onValueChange={handlePlatformChange} disabled={loading}>
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

                        {config.accountType === "XUEXITONG" && (
                            <div className="space-y-2">
                                <Label>登录方式</Label>
                                <Select
                                    value={config.loginMode === "app" ? "app" : "web"}
                                    onValueChange={(value) => updateConfig({ loginMode: value })}
                                    disabled={loading}
                                >
                                    <SelectTrigger className="w-full">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="web">Web 网页登录</SelectItem>
                                        <SelectItem value="app">Android App 登录（自动获取人脸签名）</SelectItem>
                                    </SelectContent>
                                </Select>
                                <p className="text-xs text-muted-foreground">
                                    App 登录会自动获取人脸签名 cid/sc；触发短信验证时会在账号详情页弹窗提示。
                                </p>
                            </div>
                        )}

                        {requiresUrl && (
                            <div className="space-y-2">
                                <Label htmlFor="account-url">平台 URL</Label>
                                <Input
                                    id="account-url"
                                    placeholder="例如 https://xxx.haiqikeji.com"
                                    value={config.url}
                                    onChange={(e) => updateConfig({ url: e.target.value })}
                                    disabled={loading}
                                />
                            </div>
                        )}

                        <div className="space-y-2">
                            <Label htmlFor="account-name">账号</Label>
                            <Input
                                id="account-name"
                                placeholder="请输入登录账号"
                                value={config.account}
                                onChange={(e) => updateConfig({ account: e.target.value })}
                                disabled={loading}
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="account-password">密码 / Cookie / Token</Label>
                            <div className="relative">
                                <Input
                                    id="account-password"
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
                            <Label htmlFor="account-remark">备注名（可选）</Label>
                            <Input
                                id="account-remark"
                                placeholder="用于区分多个账号"
                                value={config.remarkName || ""}
                                onChange={(e) => updateConfig({ remarkName: e.target.value })}
                                disabled={loading}
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={loading}>
                            取消
                        </Button>
                        <Button type="submit" disabled={loading}>
                            {loading ? "保存中..." : isEdit ? "保存修改" : "添加账号"}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    )
}
