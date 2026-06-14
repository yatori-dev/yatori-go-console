"use client"

import { useEffect, useState, useTransition } from "react"
import { usePathname, useRouter } from "next/navigation"
import { Sidebar } from "@/components/sidebar"
import { AccountList } from "@/components/account-list"
import { AIConfigForm } from "@/components/ai-config-form"
import { startTransition } from "react"

type Tab = "accounts" | "questions"

function pathnameToTab(pathname: string | null): Tab {
  if (!pathname) return "accounts"
  if (pathname.includes("questions")) return "questions"
  return "accounts"
}

function tabToPathname(tab: Tab): string {
  return tab === "questions" ? "/questions" : "/accounts"
}

export default function DashboardLayout({
  children: _children,
}: {
  children: React.ReactNode
}) {
  const pathname = usePathname()
  const router = useRouter()
  const [activeTab, setActiveTab] = useState<Tab>(pathnameToTab(pathname))

  // 当外部修改 URL 时（比如浏览器前进后退），同步 activeTab
  useEffect(() => {
    setActiveTab(pathnameToTab(pathname))
  }, [pathname])

  const handleTabChange = (tab: Tab) => {
    if (tab === activeTab) return
    setActiveTab(tab)
    // 用 history.replaceState 同步 URL，不触发 React Server Components 重新拉取，
    // 也不会让 Next.js router 发起整页加载
    const newPath = tabToPathname(tab)
    window.history.replaceState(window.history.state, "", newPath)
  }

  return (
    <div className="flex min-h-screen bg-background">
      <Sidebar activeTab={activeTab} onTabChange={handleTabChange} />
      <main className="flex-1 transition-all duration-300 lg:pt-0 pt-16">
        {activeTab === "accounts" ? <AccountList /> : <AIConfigForm />}
      </main>
    </div>
  )
}
