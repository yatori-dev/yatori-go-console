"use client"

import type React from "react"
import { useState } from "react"
import { Card, CardContent, CardFooter, CardHeader } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { User, BookOpen, Trash2, Eye, EyeOff, KeyRound, Leaf, Play } from "lucide-react"
import type { Account } from "@/api/accountApi"
import { getPlatformName, getVideoModeLabel, getAnswerModeLabel } from "@/utils/platformUtils"
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

    const maskText = (text: string) => {
        return "*".repeat(Math.min(text.length, 12))
    }

    const config = account.config
    const displayName = config?.remarkName?.trim() || account.account

    return (
        <>
            <Card className="cursor-pointer transition-all hover:shadow-lg hover:border-primary/50" onClick={onClick}>
                <CardHeader className="pb-3">
                    <div className="flex items-start justify-between">
                        <div className="flex items-center gap-3 min-w-0">
                            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary flex-shrink-0">
                                <User className="h-6 w-6" />
                            </div>
                            <div className="min-w-0">
                                <h3 className="font-semibold text-foreground truncate flex items-center gap-1">
                                    <span className="truncate">{showAccount ? displayName : maskText(displayName)}</span>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-6 w-6 flex-shrink-0"
                                        onClick={(e) => {
                                            e.stopPropagation()
                                            setShowAccount(!showAccount)
                                        }}
                                    >
                                        {showAccount ? <Eye className="h-3.5 w-3.5" /> : <EyeOff className="h-3.5 w-3.5" />}
                                    </Button>
                                </h3>
                                <Badge variant={account.isRunning ? "default" : "secondary"} className="mt-1">
                                    {account.isRunning ? "进行中" : "未进行"}
                                </Badge>
                            </div>
                        </div>
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 flex-shrink-0 text-muted-foreground hover:text-destructive"
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
                            <span className="truncate font-mono">
                                {showPassword ? account.password : maskText(account.password)}
                            </span>
                        </div>
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 flex-shrink-0"
                            onClick={(e) => {
                                e.stopPropagation()
                                setShowPassword(!showPassword)
                            }}
                        >
                            {showPassword ? <Eye className="h-3.5 w-3.5" /> : <EyeOff className="h-3.5 w-3.5" />}
                        </Button>
                    </div>

                    {config && (
                        <div className="flex flex-col gap-1 text-xs text-muted-foreground">
                            <div className="flex items-center gap-2">
                                <BookOpen className="h-3.5 w-3.5" />
                                <span>视频：{getVideoModeLabel(account.accountType, config.coursesCustom.videoModel)}</span>
                            </div>
                            <div className="flex items-center gap-2">
                                <Play className="h-3.5 w-3.5" />
                                <span>答题：{getAnswerModeLabel(account.accountType, config.coursesCustom.autoExam)}</span>
                            </div>
                        </div>
                    )}
                </CardContent>

                <CardFooter className="pt-3 border-t">
                    <Button variant="secondary" className="w-full" size="sm">
                        查看详情
                    </Button>
                </CardFooter>
            </Card>

            <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle className="text-destructive">确认删除账号</DialogTitle>
                        <DialogDescription>
                            你确定要删除账号 <strong>{displayName}</strong> 吗？这个操作将不可撤销。
                        </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                        <Button type="button" variant="outline" onClick={() => setDeleteDialogOpen(false)}>
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
