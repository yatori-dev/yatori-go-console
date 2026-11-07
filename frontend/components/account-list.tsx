"use client"

import { useCallback, useEffect, useState } from "react"
import { AccountCard } from "@/components/account-card"
import { AccountDetail } from "@/components/account-detail"
import { Button } from "@/components/ui/button"
import { Plus } from "lucide-react"
import { AddAccountDialog } from "@/components/add-account-dialog"
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
    const [loading, setLoading] = useState(true)

    const fetchAccounts = useCallback(async (silent = false) => {
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
            } else if (!silent) {
                toast({
                    title: "加载失败",
                    description: response.message,
                    variant: "destructive",
                })
            }
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

    return (
        <div>
            <div className="mb-4 sm:mb-6 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div>
                    <h2 className="text-xl sm:text-2xl font-semibold text-foreground">账号管理</h2>
                    <p className="text-xs sm:text-sm text-muted-foreground mt-1">
                        账号配置保存在本地数据库，不再依赖 config.yaml
                    </p>
                </div>
                <Button onClick={() => setIsAddDialogOpen(true)} className="gap-2 w-full sm:w-auto">
                    <Plus className="h-4 w-4" />
                    添加账号
                </Button>
            </div>

            {loading ? (
                <div className="flex justify-center items-center py-12">
                    <div className="text-muted-foreground">加载中...</div>
                </div>
            ) : (
                <div className="grid gap-3 sm:gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                    {accounts.map((account) => (
                        <AccountCard
                            key={account.uid}
                            account={account}
                            onClick={() => setSelectedAccount(account)}
                            onDelete={() => handleDeleteAccount(account.uid)}
                        />
                    ))}
                </div>
            )}

            {!loading && accounts.length === 0 && (
                <div className="flex flex-col items-center justify-center py-12 sm:py-16 text-center px-4">
                    <div className="rounded-full bg-muted p-4 mb-4">
                        <Plus className="h-8 w-8 text-muted-foreground" />
                    </div>
                    <h3 className="text-base sm:text-lg font-medium text-foreground mb-2">暂无账号</h3>
                    <p className="text-xs sm:text-sm text-muted-foreground mb-4">点击上方按钮添加第一个账号</p>
                </div>
            )}

            <AddAccountDialog
                open={isAddDialogOpen}
                onOpenChange={setIsAddDialogOpen}
                onAdded={() => fetchAccounts(true)}
            />
        </div>
    )
}
