"use client"

import {useState, useEffect} from "react"
import {AccountCard} from "@/components/account-card"
import {AccountDetail} from "@/components/account-detail"
import {Button} from "@/components/ui/button"
import {Plus} from "lucide-react"
import {AddAccountDialog} from "@/components/add-account-dialog"
import {getAccounts, deleteAccountForUid as apiDeleteAccountForUid} from "@/api/accountApi"
import {toast} from "@/components/ui/use-toast"

export type Account = {
    uid: string
    accountType: string
    url?: string
    account: string
    password: string
    isRunning: boolean
    userConfigJson: string
    status: "active" | "inactive"
    courseCount: number
    lastLogin?: string
}

export type Course = {
    courseId: string //课程ID
    courseName: string //课程名称
    progress: number //进度
    totalLessons: number
    completedLessons: number
    instructor: string //授课老师
}

const coursesData: Record<string, Course[]> = {
    "1": [
        {
            courseId: "c1",
            courseName: "高等数学",
            progress: 75,
            totalLessons: 48,
            completedLessons: 36,
            instructor: "王教授",
        },
        {
            courseId: "c2",
            courseName: "计算机网络",
            progress: 60,
            totalLessons: 40,
            completedLessons: 24,
            instructor: "李老师",
        },
        {
            courseId: "c3",
            courseName: "数据结构",
            progress: 85,
            totalLessons: 45,
            completedLessons: 38,
            instructor: "赵老师",
        },
        {
            courseId: "c4",
            courseName: "线性代数",
            progress: 45,
            totalLessons: 36,
            completedLessons: 16,
            instructor: "刘教授"
        },
        {
            courseId: "c5",
            courseName: "操作系统",
            progress: 30,
            totalLessons: 42,
            completedLessons: 13,
            instructor: "陈老师"
        },
    ],
}

export function AccountList() {
    const [accounts, setAccounts] = useState<Account[]>([])
    const [selectedAccount, setSelectedAccount] = useState<Account | null>(null)
    const [isAddDialogOpen, setIsAddDialogOpen] = useState(false)
    const [loading, setLoading] = useState(true)

    // 获取账号列表数据
    useEffect(() => {
        const fetchAccounts = async () => {
            setLoading(true)
            try {
                const response = await getAccounts()
                if (response.code === 200) {
                    // 将API返回的数据转换为组件需要的Account类型
                    const formattedAccounts: Account[] = response.data.users.map((user, index) => ({
                        uid: user.uid, // 使用索引作为临时ID
                        accountType: user.accountType,
                        url: user.url,
                        account: user.account,
                        password: user.password,
                        isRunning: user.isRunning,
                        userConfigJson: user.userConfigJson,
                        status: "active", // 默认状态为active
                        courseCount: 0, // 默认课程数量为0
                    }))
                    setAccounts(formattedAccounts)
                } else {
                    console.error(response.message)
                }
            } catch (error) {
                console.error("Failed to fetch accounts:", error)
            } finally {
                setLoading(false)
            }
        }

        fetchAccounts()
    }, [])

    const handleAccountClick = (account: Account) => {
        setSelectedAccount(account)
    }

    const handleBack = () => {
        setSelectedAccount(null)
    }

    const handleDeleteAccount = async (uid: string) => {
        try {
            // 调用API删除账号
            const response = await apiDeleteAccountForUid(uid)
            if (response.code === 200) {
                setAccounts(accounts.filter((acc) => acc.uid !== uid))
            } else {
                // API删除失败
                console.error("删除失败:", response.message)
                toast({
                    title: "删除失败",
                    description: response.message,
                    variant: "destructive",
                })
            }
        } catch (error) {
            console.error("删除失败:", error)
            toast({
                title: "删除失败",
                description: "网络错误，请稍后重试",
                variant: "destructive",
            })
        }
        if (selectedAccount?.uid === uid) {
            setSelectedAccount(null)
        }
    }

    const handleAddAccount = (account: Account) => {
        const newAccount = {
            ...account
        }
        setAccounts([...accounts, newAccount])
    }

    if (selectedAccount) {
        return (
            <AccountDetail account={selectedAccount} onBack={handleBack}/>
        )
    }

    return (
        <div>
            <div className="mb-4 sm:mb-6 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div>
                    <h2 className="text-xl sm:text-2xl font-semibold text-foreground">账号管理</h2>
                    <p className="text-xs sm:text-sm text-muted-foreground mt-1">管理和查看所有学习账号</p>
                </div>
                <Button onClick={() => setIsAddDialogOpen(true)} className="gap-2 w-full sm:w-auto">
                    <Plus className="h-4 w-4"/>
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
                            onClick={() => handleAccountClick(account)}
                            onDelete={() => handleDeleteAccount(account.uid)}
                        />
                    ))}
                </div>
            )}

            {!loading && accounts.length === 0 && (
                <div className="flex flex-col items-center justify-center py-12 sm:py-16 text-center px-4">
                    <div className="rounded-full bg-muted p-4 mb-4">
                        <Plus className="h-8 w-8 text-muted-foreground"/>
                    </div>
                    <h3 className="text-base sm:text-lg font-medium text-foreground mb-2">暂无账号</h3>
                    <p className="text-xs sm:text-sm text-muted-foreground mb-4">点击上方按钮添加第一个账号</p>
                </div>
            )}

            <AddAccountDialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen} onAdd={handleAddAccount}/>
        </div>
    )
}
