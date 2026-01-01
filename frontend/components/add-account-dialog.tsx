"use client"

import type React from "react"

import { useState } from "react"
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
import type { Account } from "@/components/account-list"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { addAccount as apiAddAccount } from "@/api/accountApi"
import { toast } from "@/components/ui/use-toast"

type AddAccountDialogProps = {
    open: boolean
    onOpenChange: (open: boolean) => void
    onAdd: (account: Account) => void
}

export function AddAccountDialog({ open, onOpenChange, onAdd }: AddAccountDialogProps) {
    const [accountType, setAccountType] = useState("")
    const [url, setUrl] = useState("")
    const [account, setAccount] = useState("")
    const [password, setPassword] = useState("")
    const [showPassword, setShowPassword] = useState(false)
    const [loading, setLoading] = useState(false)

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()

        // 验证必填字段
        if (!accountType.trim() || !account.trim() || !password.trim()) {
            toast({
                title: "表单验证失败",
                description: "请填写所有必填字段",
                variant: "destructive",
            })
            return
        }

        // 当选择英华学堂时，需要验证url字段
        if (accountType.trim() === "YINGHUA" && !url.trim()) {
            toast({
                title: "表单验证失败",
                description: "请填写英华学堂的平台URL",
                variant: "destructive",
            })
            return
        }

        setLoading(true)

        try {
            // 调用API添加账号
            const response = await apiAddAccount({
                accountType: accountType.trim(),
                url: accountType.trim() === "YINGHUA" ? url.trim() : "",
                account: account.trim(),
                password: password.trim(),
            })

            if (response.code===200) {
                // API添加成功
                toast({
                    title: "添加成功",
                    description: response.message || "账号已成功添加",
                    variant: "default",
                })

                // 调用父组件的onAdd函数将新账号添加到本地列表
                if (response.data) {
                    onAdd(response.data)
                } else {
                    // 如果API没有返回完整的账号数据，创建一个简化版
                    const newAccount = {
                        accountType: accountType.trim(),
                        account: account.trim(),
                        password: password.trim(),
                        status: "active",
                        courseCount: 0,
                        lastLogin: new Date().toLocaleString("zh-CN"),
                        url: accountType.trim() === "YINGHUA" ? url.trim() : undefined,
                        uid: Date.now().toString(), // 生成临时UID
                    } as Account
                    onAdd(newAccount)
                }

                // 清空表单
                setAccountType("")
                setUrl("")
                setAccount("")
                setPassword("")

                // 关闭对话框
                onOpenChange(false)
            } else {
                // API添加失败
                toast({
                    title: "添加失败",
                    description: response.message || "添加账号失败",
                    variant: "destructive",
                })
            }
        } catch (error) {
            // 网络或其他错误
            console.error("添加账号失败:", error)
            toast({
                title: "网络错误",
                description: "无法连接到服务器，请稍后重试",
                variant: "destructive",
            })
        } finally {
            setLoading(false)
        }
    }

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-md">
                <DialogHeader>
                    <DialogTitle>添加新账号</DialogTitle>
                    <DialogDescription>填写账号信息以添加到管理列表</DialogDescription>
                </DialogHeader>
                <form onSubmit={handleSubmit}>
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="account">平台</Label>
                            <Select value={accountType} onValueChange={setAccountType}>
                                <SelectTrigger id="platform" className="w-full">
                                    <SelectValue placeholder="请选择账号平台" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="XUEXITONG">学习通</SelectItem>
                                    <SelectItem value="YINGHUA">英华学堂</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        {/* 仅当选择英华学堂时显示平台URL输入项 */}
                        {accountType === "YINGHUA" && (
                            <div className="space-y-2">
                                <Label htmlFor="url">平台URL</Label>
                                <Input
                                    id="url"
                                    placeholder="请输入英华学堂的平台URL"
                                    value={url}
                                    onChange={(e) => setUrl(e.target.value)}
                                />
                            </div>
                        )}

                        <div className="space-y-2">
                            <Label htmlFor="account">账号</Label>
                            <Input
                                id="account"
                                placeholder="请输入登录账号"
                                value={account}
                                onChange={(e) => setAccount(e.target.value)}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="password">密码/Cookie/Token</Label>
                            <div className="relative">
                                <Input
                                    id="password"
                                    type={showPassword ? "text" : "password"}
                                    placeholder="请输入密码/Cookie/Token"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    className="pr-10"
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
