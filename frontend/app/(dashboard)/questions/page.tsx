import { AIConfigForm } from "@/components/ai-config-form"

export default function QuestionsPage() {
  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border/70 bg-background/85 backdrop-blur lg:sticky lg:top-0 lg:z-30">
        <div className="page-container py-5 sm:py-6">
          <div className="flex flex-col gap-1">
            <h1 className="text-xl font-semibold tracking-tight text-foreground sm:text-2xl">AI 配置</h1>
            <p className="text-sm text-muted-foreground">配置答题 AI 模型、接口地址与外置题库</p>
          </div>
        </div>
      </header>

      <div className="page-container py-5 sm:py-6">
        <AIConfigForm />
      </div>
    </div>
  )
}
