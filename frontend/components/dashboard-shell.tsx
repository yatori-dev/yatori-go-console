"use client"

import { useState } from "react"
import { ChevronLeft, ChevronRight, Menu } from "lucide-react"
import { Sidebar } from "@/components/sidebar"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

export function DashboardShell({ children }: { children: React.ReactNode }) {
    const [mobileOpen, setMobileOpen] = useState(false)
    const [collapsed, setCollapsed] = useState(false)

    return (
        <div className="min-h-screen bg-background">
            {/* 移动端顶栏：替代原来的悬浮菜单按钮，避免遮挡页面标题 */}
            <div className="sticky top-0 z-40 flex h-14 items-center gap-3 border-b border-border bg-card/90 px-4 backdrop-blur lg:hidden">
                <Button
                    variant="ghost"
                    size="icon"
                    className="-ml-1 h-9 w-9"
                    onClick={() => setMobileOpen(true)}
                    aria-label="打开菜单"
                >
                    <Menu className="h-5 w-5" />
                </Button>
                <div className="flex items-center gap-2">
                    <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-sm font-bold text-primary-foreground">
                        Y
                    </div>
                    <span className="text-sm font-semibold text-foreground">Yatori</span>
                </div>
                <span className="ml-auto text-xs text-muted-foreground">刷课控制台</span>
            </div>

            <div className="flex min-h-[calc(100vh-3.5rem)] lg:min-h-screen">
                <Sidebar mobileOpen={mobileOpen} onMobileOpenChange={setMobileOpen} collapsed={collapsed} />

                {/* 折叠按钮：固定在侧栏右侧的分界线上，不属于侧栏本身 */}
                <button
                    type="button"
                    onClick={() => setCollapsed((value) => !value)}
                    aria-label={collapsed ? "展开侧边栏" : "收起侧边栏"}
                    className={cn(
                        "fixed top-6 z-40 hidden h-6 w-6 -translate-x-1/2 items-center justify-center rounded-full",
                        "border border-border bg-card text-muted-foreground shadow-md",
                        "transition-[left,color,border-color] duration-300 ease-out",
                        "hover:border-primary/50 hover:text-primary lg:flex",
                        collapsed ? "left-[76px]" : "left-64",
                    )}
                >
                    {collapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
                </button>

                <main className="min-w-0 flex-1">{children}</main>
            </div>
        </div>
    )
}
