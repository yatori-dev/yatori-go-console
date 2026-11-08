"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { cn } from "@/lib/utils"
import { Users, Bot, X } from "lucide-react"

type SidebarProps = {
    mobileOpen: boolean
    onMobileOpenChange: (open: boolean) => void
    collapsed: boolean
}

const NAV_ITEMS = [
    { href: "/accounts", label: "账号管理", icon: Users },
    { href: "/questions", label: "AI 配置", icon: Bot },
]

export function Sidebar({ mobileOpen, onMobileOpenChange, collapsed }: SidebarProps) {
    const pathname = usePathname()

    return (
        <>
            {/* 移动端遮罩 */}
            <div
                className={cn(
                    "fixed inset-0 z-40 bg-foreground/30 backdrop-blur-sm transition-opacity lg:hidden",
                    mobileOpen ? "opacity-100" : "pointer-events-none opacity-0",
                )}
                onClick={() => onMobileOpenChange(false)}
                aria-hidden
            />

            <aside
                className={cn(
                    "fixed inset-y-0 left-0 z-50 flex w-72 flex-col border-r border-sidebar-border bg-sidebar",
                    "transition-[transform,width] duration-300 ease-out",
                    "lg:sticky lg:top-0 lg:z-30 lg:h-screen lg:shrink-0 lg:translate-x-0",
                    collapsed ? "lg:w-[76px]" : "lg:w-64",
                    mobileOpen ? "translate-x-0" : "-translate-x-full",
                )}
            >
                {/* 品牌 / 折叠 */}
                <div
                    className={cn(
                        "flex h-16 items-center gap-3 border-b border-sidebar-border px-4",
                        collapsed && "lg:justify-center lg:px-0",
                    )}
                >
                    <Link href="/accounts" className="flex min-w-0 items-center gap-3">
                        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-sidebar-primary text-sm font-bold text-sidebar-primary-foreground shadow-sm">
                            Y
                        </div>
                        {!collapsed && (
                            <div className="min-w-0">
                                <p className="truncate text-sm font-semibold leading-5 text-sidebar-foreground">Yatori</p>
                                <p className="truncate text-[11px] leading-4 text-sidebar-foreground/55">刷课控制台</p>
                            </div>
                        )}
                    </Link>

                    <button
                        onClick={() => onMobileOpenChange(false)}
                        className="ml-auto rounded-md p-1.5 text-sidebar-foreground/60 transition-colors hover:bg-sidebar-accent lg:hidden"
                        aria-label="关闭菜单"
                    >
                        <X className="h-5 w-5" />
                    </button>
                </div>

                {/* 折叠/展开按钮由 DashboardShell 贴在侧边栏右侧，避免放在侧栏内部 */}

                {/* 导航 */}
                <nav className="flex-1 space-y-1 overflow-y-auto p-3">
                    {!collapsed && (
                        <p className="px-2 pb-2 pt-1 text-[11px] font-medium uppercase tracking-wider text-sidebar-foreground/40">
                            导航
                        </p>
                    )}
                    {NAV_ITEMS.map((item) => {
                        const Icon = item.icon
                        const isActive = pathname === item.href || pathname.startsWith(`${item.href}/`)

                        return (
                            <Link
                                key={item.href}
                                href={item.href}
                                onClick={() => onMobileOpenChange(false)}
                                title={collapsed ? item.label : undefined}
                                className={cn(
                                    "group relative flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm transition-colors",
                                    collapsed && "lg:justify-center lg:px-0",
                                    isActive
                                        ? "bg-sidebar-accent font-semibold text-sidebar-accent-foreground"
                                        : "text-sidebar-foreground/70 hover:bg-sidebar-accent/60 hover:text-sidebar-foreground",
                                )}
                            >
                                {isActive && (
                                    <span className="absolute inset-y-2 left-0 w-0.5 rounded-full bg-sidebar-primary" />
                                )}
                                <Icon className={cn("h-[18px] w-[18px] shrink-0", isActive && "text-sidebar-primary")} />
                                {!collapsed && <span className="truncate">{item.label}</span>}
                            </Link>
                        )
                    })}
                </nav>

                {/* 底部 */}
                <div className={cn("border-t border-sidebar-border p-3", collapsed && "lg:px-0")}>
                    {collapsed ? (
                        <div className="hidden justify-center lg:flex">
                            <span className="h-2 w-2 rounded-full bg-emerald-500" />
                        </div>
                    ) : (
                        <div className="flex items-center gap-2 rounded-lg bg-sidebar-accent/60 px-3 py-2">
                            <span className="h-2 w-2 rounded-full bg-emerald-500" />
                            <span className="text-[11px] text-sidebar-foreground/70">服务运行中</span>
                            <span className="ml-auto text-[11px] text-sidebar-foreground/40">v2.6</span>
                        </div>
                    )}
                </div>
            </aside>
        </>
    )
}
