"use client"

import type React from "react"
import { useState } from "react"
import { Card } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { User, Trash2, Eye, EyeOff, KeyRound, ArrowRight, AtSign, Pencil } from "lucide-react"
import type { Account } from "@/api/accountApi"
import { getAnswerModeLabel, getVideoModeLabel, getPlatformName } from "@/utils/platformUtils"
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog"
import { cn } from "@/lib/utils"

type AccountCardProps = {
    account: Account
    onClick: () => void
    onEdit: () => void
    onDelete: () => void
}

export function AccountCard({ account, onClick, onEdit, onDelete }: AccountCardProps) {
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

    const maskText = (text: string) => "•".repeat(Math.min(text.length, 10))

    const config = account.config
    const displayName = config?.remarkName?.trim() || account.account
    const state = account.status?.state
    const failed = state === "failed" || Boolean(account.lastError)
    const running = account.isRunning

    return (
        <>
            <Card
                role="button"
                tabIndex={0}
                onClick={onClick}
                onKeyDown={(event) => {
                    if (event.key === "Enter" || event.key === " ") {
                        event.preventDefault()
                        onClick()
                    }
                }}
                className={cn(
                    "group cursor-pointer gap-0 overflow-hidden py-0 transition-all duration-200",
                    "hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-lg",
                    "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
                )}
            >
                <div className="flex items-start gap-3 p-5 pb-3">
                    <div className="relative shrink-0">
                        <div className="flex h-11 w-11 items-center justify-center rounded-full bg-primary/10 text-primary">
                            <User className="h-5 w-5" />
                        </div>
                        <span
                            className={cn(
                                "absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-card",
                                running ? "bg-emerald-500" : failed ? "bg-destructive" : "bg-muted-foreground/40",
                            )}
                        />
                    </div>

                    <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-2">
                            <h3 className="truncate text-sm font-semibold text-foreground">{displayName}</h3>
                            {running && (
                                <Badge variant="secondary" className="shrink-0 gap-1 px-1.5 py-0 text-[11px]">
                                    <span className="relative flex h-1.5 w-1.5">
                                        <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-500 opacity-75" />
                                        <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-emerald-500" />
                                    </span>
                                    运行中
                                </Badge>
                            )}
                            {!running && failed && (
                                <Badge variant="destructive" className="shrink-0 px-1.5 py-0 text-[11px]">
                                    失败
                                </Badge>
                            )}
                        </div>
                        <p className="mt-0.5 truncate text-xs text-muted-foreground">
                            {getPlatformName(account.accountType)}
                        </p>
                    </div>

                    <div className="flex items-center gap-0.5">
                        <Button
                            variant="ghost"
                            size="icon"
                            className="-mt-1 h-8 w-8 text-muted-foreground transition-colors hover:text-primary"
                            onClick={(e) => {
                                e.stopPropagation()
                                onEdit()
                            }}
                            disabled={account.isRunning}
                            title={account.isRunning ? "任务运行中，请先停止任务再编辑" : "编辑账号"}
                            aria-label="编辑账号"
                        >
                            <Pencil className="h-4 w-4" />
                        </Button>
                        <Button
                            variant="ghost"
                            size="icon"
                            className="-mr-1 -mt-1 h-8 w-8 text-muted-foreground transition-colors hover:text-destructive"
                            onClick={handleDeleteClick}
                            aria-label="删除账号"
                        >
                            <Trash2 className="h-4 w-4" />
                        </Button>
                    </div>
                </div>

                <div className="space-y-2.5 px-5 pb-4">
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <AtSign className="h-3.5 w-3.5 shrink-0" />
                        <span className="truncate">{account.account}</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <KeyRound className="h-3.5 w-3.5 shrink-0" />
                        <span className="truncate font-mono">
                            {showPassword ? account.password : maskText(account.password)}
                        </span>
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 shrink-0 text-muted-foreground"
                            onClick={(e) => {
                                e.stopPropagation()
                                setShowPassword(!showPassword)
                            }}
                            aria-label={showPassword ? "隐藏密码" : "显示密码"}
                        >
                            {showPassword ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
                        </Button>
                    </div>

                    {config && (
                        <div className="flex flex-wrap gap-1.5">
                            <Badge variant="secondary" className="font-normal">
                                视频：{getVideoModeLabel(account.accountType, config.coursesCustom.videoModel)}
                            </Badge>
                            <Badge variant="outline" className="font-normal">
                                答题：{getAnswerModeLabel(account.accountType, config.coursesCustom.autoExam)}
                            </Badge>
                        </div>
                    )}
                </div>

                <div className="flex items-center justify-between border-t border-border/70 px-5 py-2.5 text-xs">
                    <span className={cn("text-muted-foreground", failed && !running && "text-destructive")}>
                        {running ? "任务执行中" : failed ? "上次任务失败" : "未运行"}
                    </span>
                    <span className="inline-flex items-center gap-1 font-medium text-primary">
                        查看详情
                        <ArrowRight className="h-3.5 w-3.5 transition-transform group-hover:translate-x-0.5" />
                    </span>
                </div>
            </Card>

            <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
                <DialogContent className="sm:max-w-md" onClick={(e) => e.stopPropagation()}>
                    <DialogHeader>
                        <DialogTitle className="text-destructive">确认删除账号</DialogTitle>
                        <DialogDescription>
                            删除后账号配置和任务记录都会被移除，确定要删除 <strong>{displayName}</strong> 吗？
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
