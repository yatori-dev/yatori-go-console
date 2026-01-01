"use client"

import type React from "react"
import { useState } from "react"
import { Card, CardContent, CardFooter, CardHeader } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { User, BookOpen, Clock, Trash2, Eye, EyeOff, KeyRound, Leaf } from "lucide-react"
import type { Account } from "@/components/account-list"
import { getPlatformName } from "@/utils/platformUtils"
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog"

type AccountCardProps = {
    account: Account
    onClick: () => void
    onDelete: () => void
}

export function AccountCard({ account, onClick, onDelete }: AccountCardProps) {
    const [showAccount, setShowAccount] = useState(true)
    const [showPassword, setShowPassword] = useState(false)
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)

    const handleDeleteClick = (e: React.MouseEvent) => {
        e.stopPropagation()
        setDeleteDialogOpen(true)
    }

    const confirmDelete = () => {
        setDeleteDialogOpen(false)
        onDelete()
    }

    const cancelDelete = () => {
        setDeleteDialogOpen(false)
    }

    const toggleAccountVisibility = (e: React.MouseEvent) => {
        e.stopPropagation()
        setShowAccount(!showAccount)
    }

    const togglePasswordVisibility = (e: React.MouseEvent) => {
        e.stopPropagation()
        setShowPassword(!showPassword)
    }

    const maskText = (text: string) => {
        return "*".repeat(text.length)
    }


    return (
        <>
            <Card className="cursor-pointer transition-all hover:shadow-lg hover:border-primary/50" onClick={onClick}>
                <CardHeader className="pb-3">
                    <div className="flex items-start justify-between">
                        <div className="flex items-center gap-3">
                            <div
                                className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary">
                                <User className="h-6 w-6" />
                            </div>
                            <div>
                                <h3 className="font-semibold text-foreground">{showAccount ? account.account : maskText(account.account)}<Button
                                    variant="ghost" size="icon" className="h-6 w-6 flex-shrink-0"
                                    onClick={toggleAccountVisibility}>
                                    {showAccount ? <Eye className="h-3.5 w-3.5" /> : <EyeOff className="h-3.5 w-3.5" />}
                                </Button></h3>
                                <Badge variant={account.status === "active" ? "default" : "secondary"} className="mt-1">
                                    {account.isRunning ? "进行中" : "未进行"}
                                </Badge>
                            </div>
                        </div>
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 text-muted-foreground hover:text-destructive"
                            onClick={handleDeleteClick}
                        >
                            <Trash2 className="h-4 w-4" />
                        </Button>
                    </div>
                </CardHeader>

                <CardContent className="space-y-3 pb-3">
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <Leaf className="h-4 w-4" />
                        <span>{getPlatformName(account.accountType)}</span>
                    </div>
                    <div className="flex items-center justify-between gap-2 text-sm">
                        <div className="flex items-center gap-2 text-muted-foreground flex-1 min-w-0">
                            <KeyRound className="h-4 w-4 flex-shrink-0" />
                            <span
                                className="truncate font-mono">{showPassword ? account.password : maskText(account.password)}</span>
                        </div>
                        <Button variant="ghost" size="icon" className="h-6 w-6 flex-shrink-0"
                                onClick={togglePasswordVisibility}>
                            {showPassword ? <Eye className="h-3.5 w-3.5" /> : <EyeOff className="h-3.5 w-3.5" />}
                        </Button>
                    </div>


                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <BookOpen className="h-4 w-4" />
                        <span>{account.courseCount} 门课程</span>
                    </div>

                    {account.lastLogin && (
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                            <Clock className="h-4 w-4" />
                            <span>{account.lastLogin}</span>
                        </div>
                    )}
                </CardContent>

                <CardFooter className="pt-3 border-t">
                    <Button variant="secondary" className="w-full" size="sm">
                        查看详情
                    </Button>
                </CardFooter>
            </Card>

            {/* 自定义删除确认对话框 */}
            <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle className="text-destructive">确认删除账号</DialogTitle>
                        <DialogDescription>
                            你确定要删除账号 <strong>{account.account}</strong> 吗？这个操作将不可撤销。
                        </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                        <Button type="button" variant="outline" onClick={cancelDelete}>
                            取消
                        </Button>
                        <Button type="button" variant="destructive" onClick={confirmDelete}>
                            删除
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </>
    )
}
