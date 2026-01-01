import { AccountList } from "@/components/account-list"

export default function AccountsPage() {
  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border bg-card sticky top-0 z-10">
        <div className="px-4 sm:px-6 lg:px-8 py-4 sm:py-6">
          <h1 className="text-xl sm:text-2xl font-bold text-foreground">账号管理</h1>
          <p className="text-xs sm:text-sm text-muted-foreground mt-1">管理和查看所有学习账号</p>
        </div>
      </header>

      <div className="px-4 sm:px-6 lg:px-8 py-4 sm:py-6">
        <AccountList />
      </div>
    </div>
  )
}
