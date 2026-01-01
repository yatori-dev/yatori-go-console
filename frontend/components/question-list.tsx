"use client"

import { useState } from "react"
import { Card, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Plus, Search, FileQuestion, Clock, CheckCircle2, XCircle, AlertCircle, Trash2 } from "lucide-react"
import { AddQuestionDialog } from "@/components/add-question-dialog"

export type Question = {
  id: string
  question: string
  options: string[]
  correctAnswer: number
  category: string
  difficulty: "easy" | "medium" | "hard"
  courseId?: string
  courseName?: string
  lastUsed?: string
  accuracy?: number
}

const initialQuestions: Question[] = [
  {
    id: "q1",
    question: "以下哪个不是JavaScript的数据类型？",
    options: ["String", "Number", "Boolean", "Float"],
    correctAnswer: 3,
    category: "编程",
    difficulty: "easy",
    courseName: "JavaScript基础",
    lastUsed: "2025-01-15",
    accuracy: 85,
  },
  {
    id: "q2",
    question: "TCP协议属于OSI模型的哪一层？",
    options: ["应用层", "传输层", "网络层", "数据链路层"],
    correctAnswer: 1,
    category: "计算机网络",
    difficulty: "medium",
    courseName: "计算机网络",
    lastUsed: "2025-01-14",
    accuracy: 72,
  },
  {
    id: "q3",
    question: "以下哪个排序算法的平均时间复杂度为O(n log n)？",
    options: ["冒泡排序", "选择排序", "快速排序", "插入排序"],
    correctAnswer: 2,
    category: "数据结构",
    difficulty: "medium",
    courseName: "数据结构与算法",
    lastUsed: "2025-01-13",
    accuracy: 68,
  },
  {
    id: "q4",
    question: "下列哪个是微积分的基本定理？",
    options: ["牛顿-莱布尼茨公式", "泰勒公式", "拉格朗日中值定理", "柯西不等式"],
    correctAnswer: 0,
    category: "数学",
    difficulty: "hard",
    courseName: "高等数学",
    lastUsed: "2025-01-12",
    accuracy: 55,
  },
  {
    id: "q5",
    question: "HTTP状态码404表示什么？",
    options: ["服务器错误", "未找到资源", "请求成功", "权限不足"],
    correctAnswer: 1,
    category: "计算机网络",
    difficulty: "easy",
    courseName: "Web开发",
    lastUsed: "2025-01-15",
    accuracy: 92,
  },
]

export function QuestionList() {
  const [questions, setQuestions] = useState<Question[]>(initialQuestions)
  const [selectedQuestions, setSelectedQuestions] = useState<string[]>([])
  const [searchTerm, setSearchTerm] = useState("")
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false)
  const [filterDifficulty, setFilterDifficulty] = useState<string>("all")

  const filteredQuestions = questions.filter((q) => {
    const matchesSearch =
      q.question.toLowerCase().includes(searchTerm.toLowerCase()) ||
      q.category.toLowerCase().includes(searchTerm.toLowerCase()) ||
      q.courseName?.toLowerCase().includes(searchTerm.toLowerCase())

    const matchesDifficulty = filterDifficulty === "all" || q.difficulty === filterDifficulty

    return matchesSearch && matchesDifficulty
  })

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedQuestions(filteredQuestions.map((q) => q.id))
    } else {
      setSelectedQuestions([])
    }
  }

  const handleSelectQuestion = (questionId: string, checked: boolean) => {
    if (checked) {
      setSelectedQuestions([...selectedQuestions, questionId])
    } else {
      setSelectedQuestions(selectedQuestions.filter((id) => id !== questionId))
    }
  }

  const handleDeleteSelected = () => {
    if (selectedQuestions.length === 0) return

    if (confirm(`确定要删除选中的 ${selectedQuestions.length} 道题目吗？`)) {
      setQuestions(questions.filter((q) => !selectedQuestions.includes(q.id)))
      setSelectedQuestions([])
    }
  }

  const handleAddQuestion = (question: Omit<Question, "id">) => {
    const newQuestion = {
      ...question,
      id: `q${Date.now()}`,
    }
    setQuestions([newQuestion, ...questions])
  }

  const getDifficultyColor = (difficulty: string) => {
    switch (difficulty) {
      case "easy":
        return "bg-green-500/10 text-green-700 dark:text-green-400"
      case "medium":
        return "bg-yellow-500/10 text-yellow-700 dark:text-yellow-400"
      case "hard":
        return "bg-red-500/10 text-red-700 dark:text-red-400"
      default:
        return ""
    }
  }

  const getDifficultyLabel = (difficulty: string) => {
    switch (difficulty) {
      case "easy":
        return "简单"
      case "medium":
        return "中等"
      case "hard":
        return "困难"
      default:
        return difficulty
    }
  }

  const getAccuracyIcon = (accuracy?: number) => {
    if (!accuracy) return null

    if (accuracy >= 80) {
      return <CheckCircle2 className="h-4 w-4 text-green-500" />
    } else if (accuracy >= 60) {
      return <AlertCircle className="h-4 w-4 text-yellow-500" />
    } else {
      return <XCircle className="h-4 w-4 text-red-500" />
    }
  }

  return (
    <div>
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="搜索题目、分类或课程..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>
        <div className="flex gap-2">
          <div className="flex gap-1 bg-muted rounded-lg p-1">
            <Button
              variant={filterDifficulty === "all" ? "secondary" : "ghost"}
              size="sm"
              onClick={() => setFilterDifficulty("all")}
            >
              全部
            </Button>
            <Button
              variant={filterDifficulty === "easy" ? "secondary" : "ghost"}
              size="sm"
              onClick={() => setFilterDifficulty("easy")}
            >
              简单
            </Button>
            <Button
              variant={filterDifficulty === "medium" ? "secondary" : "ghost"}
              size="sm"
              onClick={() => setFilterDifficulty("medium")}
            >
              中等
            </Button>
            <Button
              variant={filterDifficulty === "hard" ? "secondary" : "ghost"}
              size="sm"
              onClick={() => setFilterDifficulty("hard")}
            >
              困难
            </Button>
          </div>
          <Button onClick={() => setIsAddDialogOpen(true)} className="gap-2">
            <Plus className="h-4 w-4" />
            添加题目
          </Button>
        </div>
      </div>

      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Checkbox
              id="select-all-questions"
              checked={selectedQuestions.length === filteredQuestions.length && filteredQuestions.length > 0}
              onCheckedChange={handleSelectAll}
            />
            <label htmlFor="select-all-questions" className="text-sm text-muted-foreground cursor-pointer">
              全选 ({selectedQuestions.length}/{filteredQuestions.length})
            </label>
          </div>
        </div>
        {selectedQuestions.length > 0 && (
          <Button variant="destructive" size="sm" onClick={handleDeleteSelected} className="gap-2">
            <Trash2 className="h-4 w-4" />
            删除选中 ({selectedQuestions.length})
          </Button>
        )}
      </div>

      <div className="space-y-3">
        {filteredQuestions.map((question) => (
          <Card
            key={question.id}
            className={`transition-all ${
              selectedQuestions.includes(question.id) ? "border-primary bg-primary/5" : "hover:border-primary/50"
            }`}
          >
            <CardContent className="p-4">
              <div className="flex items-start gap-4">
                <Checkbox
                  id={question.id}
                  checked={selectedQuestions.includes(question.id)}
                  onCheckedChange={(checked) => handleSelectQuestion(question.id, checked as boolean)}
                  className="mt-1"
                />
                <div className="flex-1 min-w-0">
                  <div className="flex items-start justify-between gap-4 mb-3">
                    <div className="flex-1">
                      <h4 className="font-semibold text-foreground mb-2">{question.question}</h4>
                      <div className="flex flex-wrap items-center gap-2 mb-3">
                        <Badge variant="outline">{question.category}</Badge>
                        <Badge className={getDifficultyColor(question.difficulty)}>
                          {getDifficultyLabel(question.difficulty)}
                        </Badge>
                        {question.courseName && <Badge variant="secondary">{question.courseName}</Badge>}
                      </div>
                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mb-3">
                        {question.options.map((option, index) => (
                          <div
                            key={index}
                            className={`text-sm px-3 py-2 rounded-md border ${
                              index === question.correctAnswer
                                ? "bg-green-500/10 border-green-500/30 text-green-700 dark:text-green-400"
                                : "bg-muted/50 border-border"
                            }`}
                          >
                            <span className="font-medium mr-2">{String.fromCharCode(65 + index)}.</span>
                            {option}
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    {question.lastUsed && (
                      <div className="flex items-center gap-1">
                        <Clock className="h-4 w-4" />
                        <span>最后使用: {question.lastUsed}</span>
                      </div>
                    )}
                    {question.accuracy !== undefined && (
                      <div className="flex items-center gap-1">
                        {getAccuracyIcon(question.accuracy)}
                        <span>准确率: {question.accuracy}%</span>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {filteredQuestions.length === 0 && (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <FileQuestion className="h-12 w-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-medium text-foreground mb-2">{searchTerm ? "未找到匹配的题目" : "暂无题目"}</h3>
            <p className="text-sm text-muted-foreground">
              {searchTerm ? "尝试更改搜索条件" : "点击上方按钮添加第一道题目"}
            </p>
          </CardContent>
        </Card>
      )}

      <AddQuestionDialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen} onAdd={handleAddQuestion} />
    </div>
  )
}
