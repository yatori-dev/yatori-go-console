import { AccountList } from "@/components/account-list"

export default function AccountsPage() {
  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border/70 bg-background/85 backdrop-blur lg:sticky lg:top-0 lg:z-30">
        <div className="page-container py-5 sm:py-6">
          <div className="flex flex-col gap-1">
            <h1 className="text-xl font-semibold tracking-tight text-foreground sm:text-2xl">账号管理</h1>
            <p className="text-sm text-muted-foreground">管理学习账号、查看任务状态与运行日志</p>
          </div>
        </div>
      </header>

      <div className="page-container py-5 sm:py-6">
        <AccountList />
      </div>
    </div>
  )
}
