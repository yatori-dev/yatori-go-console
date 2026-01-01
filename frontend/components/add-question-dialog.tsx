"use client"

import type React from "react"

import { useState } from "react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import type { Question } from "@/components/question-list"

type AddQuestionDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  onAdd: (question: Omit<Question, "id">) => void
}

export function AddQuestionDialog({ open, onOpenChange, onAdd }: AddQuestionDialogProps) {
  const [question, setQuestion] = useState("")
  const [options, setOptions] = useState(["", "", "", ""])
  const [correctAnswer, setCorrectAnswer] = useState(0)
  const [category, setCategory] = useState("")
  const [difficulty, setDifficulty] = useState<"easy" | "medium" | "hard">("medium")
  const [courseName, setCourseName] = useState("")

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (!question.trim()) {
      alert("请输入题目")
      return
    }

    if (options.some((opt) => !opt.trim())) {
      alert("请填写所有选项")
      return
    }

    if (!category.trim()) {
      alert("请输入分类")
      return
    }

    onAdd({
      question: question.trim(),
      options: options.map((opt) => opt.trim()),
      correctAnswer,
      category: category.trim(),
      difficulty,
      courseName: courseName.trim() || undefined,
      lastUsed: new Date().toISOString().split("T")[0],
      accuracy: undefined,
    })

    // Reset form
    setQuestion("")
    setOptions(["", "", "", ""])
    setCorrectAnswer(0)
    setCategory("")
    setDifficulty("medium")
    setCourseName("")
    onOpenChange(false)
  }

  const handleOptionChange = (index: number, value: string) => {
    const newOptions = [...options]
    newOptions[index] = value
    setOptions(newOptions)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>添加新题目</DialogTitle>
          <DialogDescription>填写题目信息以添加到题库</DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="question">题目</Label>
              <Textarea
                id="question"
                placeholder="请输入题目内容"
                value={question}
                onChange={(e) => setQuestion(e.target.value)}
                rows={3}
              />
            </div>

            <div className="space-y-3">
              <Label>选项</Label>
              <RadioGroup
                value={correctAnswer.toString()}
                onValueChange={(val) => setCorrectAnswer(Number.parseInt(val))}
              >
                {options.map((option, index) => (
                  <div key={index} className="flex items-center gap-3">
                    <RadioGroupItem value={index.toString()} id={`option-${index}`} />
                    <Label htmlFor={`option-${index}`} className="font-normal flex-1">
                      <Input
                        placeholder={`选项 ${String.fromCharCode(65 + index)}`}
                        value={option}
                        onChange={(e) => handleOptionChange(index, e.target.value)}
                      />
                    </Label>
                  </div>
                ))}
              </RadioGroup>
              <p className="text-xs text-muted-foreground">选择正确答案</p>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="category">分类</Label>
                <Input
                  id="category"
                  placeholder="如：计算机网络"
                  value={category}
                  onChange={(e) => setCategory(e.target.value)}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="difficulty">难度</Label>
                <Select value={difficulty} onValueChange={(val) => setDifficulty(val as "easy" | "medium" | "hard")}>
                  <SelectTrigger id="difficulty">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="easy">简单</SelectItem>
                    <SelectItem value="medium">中等</SelectItem>
                    <SelectItem value="hard">困难</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="courseName">课程名称（可选）</Label>
              <Input
                id="courseName"
                placeholder="如：计算机网络基础"
                value={courseName}
                onChange={(e) => setCourseName(e.target.value)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              取消
            </Button>
            <Button type="submit">添加题目</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
