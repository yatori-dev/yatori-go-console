"use client"

import { useCallback, useEffect, useState } from "react"
import { AccountCard } from "@/components/account-card"
import { AccountDetail } from "@/components/account-detail"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Plus, RefreshCw, Users } from "lucide-react"
import { AddAccountDialog } from "@/components/add-account-dialog"
import { AccountFormDialog } from "@/components/account-form-dialog"
import {
    deleteAccountForUid as apiDeleteAccountForUid,
    getAccounts,
    normalizeConfig,
    type Account,
    type AccountConfig,
} from "@/api/accountApi"
import { toast } from "@/components/ui/use-toast"

export type { Account } from "@/api/accountApi"

export function AccountList() {
    const [accounts, setAccounts] = useState<Account[]>([])
    const [selectedAccount, setSelectedAccount] = useState<Account | null>(null)
    const [isAddDialogOpen, setIsAddDialogOpen] = useState(false)
    const [editingAccount, setEditingAccount] = useState<Account | null>(null)
    const [loading, setLoading] = useState(true)
    const [refreshing, setRefreshing] = useState(false)

    const fetchAccounts = useCallback(async (silent = false): Promise<boolean> => {
        if (!silent) {
            setLoading(true)
        }
        try {
            const response = await getAccounts()
            if (response.code === 200 && response.data) {
                const formatted = response.data.users.map((user) => {
                    let config: Partial<AccountConfig> | undefined = user.config
                    if (!config && user.userConfigJson) {
                        try {
                            config = JSON.parse(user.userConfigJson)
                        } catch {
                            config = undefined
                        }
                    }
                    return {
                        ...user,
                        config: normalizeConfig(config),
                        isRunning: Boolean(user.isRunning),
                    }
                })
                setAccounts(formatted)
                // 用函数式更新读取最新选中状态：
                // 返回列表时 setSelectedAccount(null) 后，这里不会再用旧闭包把详情页“拉回来”。
                setSelectedAccount((current) => {
                    if (!current) {
                        return null
                    }
                    const next = formatted.find((item) => item.uid === current.uid)
                    if (!next) {
                        return current
                    }
                    // 只同步运行态，避免每 5 秒替换整个对象导致详情页重渲染/滚动跳动
                    if (
                        current.isRunning === next.isRunning &&
                        current.lastError === next.lastError &&
                        current.status?.state === next.status?.state &&
                        current.status?.taskId === next.status?.taskId
                    ) {
                        return current
                    }
                    return {
                        ...current,
                        isRunning: next.isRunning,
                        lastError: next.lastError,
                        status: next.status,
                    }
                })
                return true
            }
            if (!silent) {
                toast({
                    title: "加载失败",
                    description: response.message,
                    variant: "destructive",
                })
            }
            return false
        } catch {
            if (!silent) {
                toast({ title: "加载失败", description: "网络异常，请稍后重试", variant: "destructive" })
            }
            return false
        } finally {
            if (!silent) {
                setLoading(false)
            }
        }
    }, [])

    useEffect(() => {
        fetchAccounts()
    }, [fetchAccounts])

    // 轮询运行状态，避免任务在后台跑而界面一直显示未开始
    useEffect(() => {
        const timer = setInterval(() => {
            fetchAccounts(true)
        }, 5000)
        return () => clearInterval(timer)
    }, [fetchAccounts])

    const handleRefresh = async () => {
        if (refreshing) {
            return
        }
        setRefreshing(true)
        const success = await fetchAccounts(true)
        setRefreshing(false)
        if (success) {
            toast({
                title: "已刷新",
                description: `账号列表已更新（${new Date().toLocaleTimeString("zh-CN", { hour12: false })}）`,
            })
        }
    }

    const handleDeleteAccount = async (uid: string) => {
        const response = await apiDeleteAccountForUid(uid)
        if (response.code === 200) {
            setAccounts((prev) => prev.filter((account) => account.uid !== uid))
            if (selectedAccount?.uid === uid) {
                setSelectedAccount(null)
            }
            toast({ title: "删除成功", description: "账号已删除" })
        } else {
            toast({
                title: "删除失败",
                description: response.message,
                variant: "destructive",
            })
        }
    }

    if (selectedAccount) {
        return (
            <AccountDetail
                account={selectedAccount}
                onBack={() => {
                    setSelectedAccount(null)
                    fetchAccounts(true)
                }}
                onChanged={() => fetchAccounts(true)}
            />
        )
    }

    const runningCount = accounts.filter((account) => account.isRunning).length

    return (
        <div className="space-y-5">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                    <span className="inline-flex items-center gap-1.5">
                        <Users className="h-4 w-4" />
                        共 <span className="font-medium text-foreground">{accounts.length}</span> 个账号
                    </span>
                    {runningCount > 0 && (
                        <Badge variant="secondary" className="gap-1.5 font-normal">
                            <span className="relative flex h-1.5 w-1.5">
                                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-500 opacity-75" />
                                <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-emerald-500" />
                            </span>
                            {runningCount} 个运行中
                        </Badge>
                    )}
                </div>
                <div className="flex items-center gap-2">
                    <Button
                        variant="outline"
                        size="sm"
                        className="gap-1.5"
                        onClick={handleRefresh}
                        disabled={loading || refreshing}
                    >
                        <RefreshCw className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`} />
                        {refreshing ? "刷新中..." : "刷新"}
                    </Button>
                    <Button className="gap-1.5" onClick={() => setIsAddDialogOpen(true)}>
                        <Plus className="h-4 w-4" />
                        添加账号
                    </Button>
                </div>
            </div>

            {loading ? (
                <div className="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(320px,1fr))]">
                    {Array.from({ length: 3 }).map((_, index) => (
                        <div key={index} className="surface h-44 animate-pulse p-5">
                            <div className="flex items-center gap-3">
                                <div className="h-11 w-11 rounded-full bg-muted" />
                                <div className="flex-1 space-y-2">
                                    <div className="h-3.5 w-1/2 rounded bg-muted" />
                                    <div className="h-3 w-1/3 rounded bg-muted" />
                                </div>
                            </div>
                            <div className="mt-6 space-y-2">
                                <div className="h-3 w-2/3 rounded bg-muted" />
                                <div className="h-3 w-1/2 rounded bg-muted" />
                            </div>
                        </div>
                    ))}
                </div>
            ) : accounts.length > 0 ? (
                <div className="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(320px,1fr))]">
                    {accounts.map((account) => (
                        <AccountCard
                            key={account.uid}
                            account={account}
                            onClick={() => setSelectedAccount(account)}
                            onEdit={() => setEditingAccount(account)}
                            onDelete={() => handleDeleteAccount(account.uid)}
                        />
                    ))}
                </div>
            ) : (
                <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border bg-card/50 px-6 py-16 text-center">
                    <div className="mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-primary/10 text-primary">
                        <Users className="h-7 w-7" />
                    </div>
                    <h3 className="text-base font-semibold text-foreground">还没有账号</h3>
                    <p className="mt-1 max-w-sm text-sm text-muted-foreground">
                        添加第一个学习账号，之后可以在账号详情里选课、配置刷课模式和查看进度
                    </p>
                    <Button className="mt-5 gap-1.5" onClick={() => setIsAddDialogOpen(true)}>
                        <Plus className="h-4 w-4" />
                        添加账号
                    </Button>
                </div>
            )}

            <AddAccountDialog
                open={isAddDialogOpen}
                onOpenChange={setIsAddDialogOpen}
                onAdded={() => fetchAccounts(true)}
            />

            <AccountFormDialog
                open={!!editingAccount}
                account={editingAccount}
                onOpenChange={(open) => {
                    if (!open) {
                        setEditingAccount(null)
                    }
                }}
                onSaved={() => {
                    setEditingAccount(null)
                    fetchAccounts(true)
                }}
            />
        </div>
    )
}
