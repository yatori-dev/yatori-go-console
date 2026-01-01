"use client"

import {useState, useEffect, useCallback} from "react"
import {motion} from "framer-motion"
import {Button} from "@/components/ui/button"
import {Card, CardContent, CardHeader} from "@/components/ui/card"
import {Badge} from "@/components/ui/badge"
import {Input} from "@/components/ui/input"
import {Checkbox} from "@/components/ui/checkbox"
import {Progress} from "@/components/ui/progress"
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select"
import {ArrowLeft, Play, Pause, User, Check, BookOpen, Loader2, Eye, EyeOff, KeyRound, Leaf, Trash2, Lock, RefreshCcw} from "lucide-react"
import type {Account, Course} from "@/components/account-list"
import {getPlatformName} from "@/utils/platformUtils"
import {getCourseList} from "@/api/courseApi";
import {addAccount as apiAddAccount} from "@/api";
import {toast} from "@/components/ui/use-toast";

type AccountDetailProps = {
    account: Account
    onBack: () => void
}

export function AccountDetail({account, onBack}: AccountDetailProps) {
    const [selectedCourses, setSelectedCourses] = useState<string[]>([])
    const [courseList, setCourseList] = useState<Course[]>([])
    const [isLoadingCourses, setIsLoadingCourses] = useState(false)
    const [isProcessing, setIsProcessing] = useState(false)
    const [showAccount, setShowAccount] = useState(true)
    const [showPassword, setShowPassword] = useState(false)

    // 功能配置状态
    // const [videoMode, setVideoMode] = useState<string>("0")
    // const [examMode, setExamMode] = useState<string>("0")
    // const [submitMode, setSubmitMode] = useState<string>("0")

    //账号配置
    const [accountConfig, setAccountConfig] = useState<any>(JSON.parse(account.userConfigJson))
    // 按钮状态
    const [isRunning, setIsRunning] = useState(false)

    // 通知邮箱列表状态
    const [emails, setEmails] = useState<string[]>([]);

    // 分页状态
    const [currentPage, setCurrentPage] = useState(1);
    const [pageSize, setPageSize] = useState(5);

    const handleLoadCourses = useCallback(async () => {
        try {
            setIsLoadingCourses(true)
            const response = await getCourseList(account.uid)

            if (response.code === 200) {
                //赋值课程列表
                setCourseList(response.data['courseList'])
                toast({
                    title: "课程加载成功",
                    description: response.message || "课程列表已成功加载",
                    variant: "default",
                })
            } else {
                // API加载失败
                toast({
                    title: "课程加载失败",
                    description: response.message || "课程列表加载失败",
                    variant: "destructive",
                })
            }
        } catch (error) {
            // 网络或其他错误
            console.error("课程加载失败:", error)
            toast({
                title: "网络错误",
                description: "无法连接到服务器，请稍后重试",
                variant: "destructive",
            })
        } finally {
            setIsLoadingCourses(false)
        }
    }, [account.uid])

    // 页面加载时自动加载课程列表
    useEffect(() => {
        handleLoadCourses()
    }, [account, handleLoadCourses])

    // 设置视屏模式
    const handleVideoModel=(val:Number)=>{
        setAccountConfig((prev:any)=>({
            ...prev,
            coursesCustom:{
                ...prev.coursesCustom,
                videoModel:val
            }
        }))
    }

    // 设置考试模式
    const handleAutoExam=(val:Number)=>{
        setAccountConfig((prev:any)=>({
            ...prev,
            coursesCustom:{
                ...prev.coursesCustom,
                autoExam:val
            }
        }))
    }

    // 设置视屏模式
    const handleExamAutoSubmit=(val:Number)=>{
        setAccountConfig((prev:any)=>({
            ...prev,
            coursesCustom:{
                ...prev.coursesCustom,
                examAutoSubmit:val
            }
        }))
    }

    // 添加邮箱
    const handleAddEmail = () => {
        setEmails([...emails, ""])
    }

    // 更新邮箱
    const handleEmailChange = (index: number, value: string) => {
        const newEmails = [...emails]
        newEmails[index] = value
        setEmails(newEmails)
    }

    // 删除邮箱
    const handleRemoveEmail = (index: number) => {
        const newEmails = emails.filter((_, i) => i !== index)
        setEmails(newEmails)
    }

    // 分页计算
    const paginatedCourses = () => {
        const startIndex = (currentPage - 1) * pageSize;
        const endIndex = startIndex + pageSize;
        return courseList.slice(startIndex, endIndex);
    };

    // 计算总页数
    const totalPages = Math.ceil(courseList.length / pageSize);

    const handleSelectAll = (checked: boolean) => {
        if (checked) {
            setSelectedCourses(courseList.map((c) => c.courseId))
        } else {
            setSelectedCourses([])
        }
    }

    const handleSelectCourse = (courseId: string, checked: boolean) => {
        if (checked) {
            setSelectedCourses([...selectedCourses, courseId])
        } else {
            setSelectedCourses(selectedCourses.filter((id) => id !== courseId))
        }
    }

    const handleStart = async () => {
        if (selectedCourses.length === 0) {
            alert("请至少选择一门课程")
            return
        }

        setIsProcessing(true)
        await new Promise((resolve) => setTimeout(resolve, 2000))
        alert(`已开始处理 ${selectedCourses.length} 门课程`)
        setIsProcessing(false)
        setSelectedCourses([])
    }

    const maskText = (text: string) => {
        return "*".repeat(text.length)
    }


    return (
        <motion.div
            initial={{opacity: 0, y: 20}}
            animate={{opacity: 1, y: 0}}
            transition={{duration: 0.5, ease: "easeOut"}}
        >
            <Button variant="ghost" onClick={onBack} className="mb-4 sm:mb-6 gap-2 text-sm sm:text-base">
                <ArrowLeft className="h-4 w-4"/>
                返回账号列表
            </Button>

            <Card className="mb-4 sm:mb-6">
                <CardHeader>
                    <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4">
                        <div
                            className="flex h-12 w-12 sm:h-16 sm:w-16 items-center justify-center rounded-full bg-primary/10 text-primary">
                            <User className="h-6 w-6 sm:h-8 sm:w-8"/>
                        </div>
                        <div className="flex-1 w-full">
                            <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-3 mb-2">
                                <h2 className="text-xl sm:text-2xl font-semibold text-foreground">{showAccount ? account.account : maskText(account.account)}
                                    <Button variant="ghost" size="icon" className="h-6 w-6"
                                            onClick={() => setShowAccount(!showAccount)}>
                                        {showAccount ? <Eye className="h-3 w-3"/> : <EyeOff className="h-3 w-3"/>}
                                    </Button>
                                </h2>
                                <Badge variant={account.status === "active" ? "default" : "secondary"}>
                                    {account.status === "active" ? "活跃" : "未激活"}
                                </Badge>
                            </div>
                            <div className="space-y-2">
                                <div className="flex items-center gap-2 text-xs sm:text-sm text-muted-foreground">
                                    <Leaf className="h-3 w-3 sm:h-4 sm:w-4"/>
                                    <span>{getPlatformName(account.accountType)}</span>
                                </div>
                                <div className="flex items-center justify-between gap-2 text-xs sm:text-sm">
                                    <div className="flex items-center gap-2 text-muted-foreground">
                                        <KeyRound className="h-3 w-3 sm:h-4 sm:w-4"/>
                                        <span
                                            className="font-mono">{showPassword ? account.password : maskText(account.password)}</span>
                                    </div>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-6 w-6"
                                        onClick={() => setShowPassword(!showPassword)}
                                    >
                                        {showPassword ? <Eye className="h-3 w-3"/> : <EyeOff className="h-3 w-3"/>}
                                    </Button>
                                </div>
                                <div className="flex items-center gap-2 text-xs sm:text-sm text-muted-foreground">
                                    <BookOpen className="h-3 w-3 sm:h-4 sm:w-4"/>
                                    <span>{courseList.length} 门课程</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </CardHeader>
                <CardContent className="space-y-6">
                    <div className="flex flex-wrap gap-3">
                        <Button
                            variant="default"
                            className="gap-2 transition-all hover:shadow-md"
                            disabled={isRunning}
                        >
                            <Check className="h-4 w-4"/>
                            保存
                        </Button>
                        <Button
                            variant={isRunning ? "destructive" : "default"}
                            className="gap-2 transition-all hover:shadow-md"
                            onClick={() => setIsRunning(prev => !prev)}
                        >
                            {isRunning ? <Pause className="h-4 w-4 animate-pulse"/> : <Play className="h-4 w-4"/>}
                            {isRunning ? "取消" : "开始"}
                        </Button>
                    </div>
                </CardContent>
            </Card>
            {/* 功能配置卡片 */}
            <div className="relative mb-4 sm:mb-6">
                <Card className="">
                    <CardHeader>
                        <h3 className="text-lg sm:text-xl font-semibold text-foreground">功能配置</h3>
                    </CardHeader>
                    <CardContent className="space-y-6">
                        {/* 刷视频模式选择 */}
                        <div className="flex items-center justify-between gap-4">
                            <label className="text-sm sm:text-base font-medium text-foreground">刷视频模式</label>
                            <div className="flex-1 max-w-sm">
                                <Select
                                    // value={videoMode}
                                    value={String(accountConfig.coursesCustom.videoModel)}
                                    onValueChange={(val) => handleVideoModel(Number(val))}
                                    // onValueChange={(val)=>setVideoMode(Number(val))}
                                    disabled={isRunning}
                                >
                                    <SelectTrigger>
                                        <SelectValue placeholder="选择刷视频模式"/>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="0">不刷</SelectItem>
                                        <SelectItem value="1">普通模式</SelectItem>
                                        <SelectItem value="2">暴力模式</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </div>

                        {/* 自动考试模式选择 */}
                        <div className="flex items-center justify-between gap-4">
                            <label className="text-sm sm:text-base font-medium text-foreground">自动考试模式</label>
                            <div className="flex-1 max-w-sm">
                                <Select
                                    value={String(accountConfig.coursesCustom.autoExam)}
                                    onValueChange={(val) => handleAutoExam(Number(val))}
                                    disabled={isRunning}
                                >
                                    <SelectTrigger>
                                        <SelectValue placeholder="选择自动考试模式"/>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="0">不考</SelectItem>
                                        <SelectItem value="1">AI考试</SelectItem>
                                        <SelectItem value="2">外置题库考试</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </div>

                        {/* 自动提交试卷模式 */}
                        <div className="flex items-center justify-between gap-4">
                            <label className="text-sm sm:text-base font-medium text-foreground">提交试卷模式</label>
                            <div className="flex-1 max-w-sm">
                                <Select
                                    value={String(accountConfig.coursesCustom.examAutoSubmit)}
                                    onValueChange={(val) => handleExamAutoSubmit(Number(val))}
                                    disabled={isRunning}
                                >
                                    <SelectTrigger>
                                        <SelectValue placeholder="选择提交试卷模式"/>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="0">只保存</SelectItem>
                                        <SelectItem value="1">自动提交</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </div>

                        {/* 通知邮箱列表 */}
                        <div className="space-y-3">
                            <div className="flex items-center justify-between">
                                <label className="text-sm sm:text-base font-medium text-foreground">通知邮箱列表</label>
                                <Button
                                    variant="outline"
                                    size="sm"
                                    className="gap-2"
                                    onClick={handleAddEmail}
                                    disabled={isRunning}
                                >
                                    新增邮箱
                                </Button>
                            </div>
                            <div className="space-y-2">
                                {emails.map((email, index) => (
                                    <div key={index} className="flex items-center gap-2">
                                        <div className="flex-1 max-w-sm">
                                            <Input
                                                type="email"
                                                placeholder="输入邮箱地址"
                                                value={email}
                                                onChange={(e) => handleEmailChange(index, e.target.value)}
                                                className="w-full"
                                                disabled={isRunning}
                                            />
                                        </div>
                                        <Button
                                            variant="outline"
                                            size="icon"
                                            className="h-9 w-9"
                                            onClick={() => handleRemoveEmail(index)}
                                            disabled={isRunning}
                                        >
                                            <Trash2 className="h-4 w-4 text-muted-foreground"/>
                                        </Button>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </CardContent>
                </Card>
                {isRunning && (
                    <div className="absolute inset-0 bg-gray-500/20 backdrop-blur-[1px] pointer-events-none z-10 rounded-lg flex flex-col items-center justify-center gap-2">
                        <Lock className="h-8 w-8 text-gray-700" />
                        <span className="text-base font-medium text-gray-800">请先取消任务</span>
                    </div>
                )}
            </div>

            <div className="relative">
                <div
                    className="mb-4 sm:mb-6 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 sm:gap-4">
                    <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4">
                        <h3 className="text-lg sm:text-xl font-semibold text-foreground">课程列表</h3>
                        <div className="flex items-center gap-2">
                            <Checkbox
                                id="select-all"
                                checked={selectedCourses.length === courseList.length && courseList.length > 0}
                                onCheckedChange={handleSelectAll}
                                disabled={isRunning}
                            />
                            <label htmlFor="select-all"
                                   className="text-xs sm:text-sm text-muted-foreground cursor-pointer">
                                全选 ({selectedCourses.length}/{courseList.length})
                            </label>
                        </div>
                    </div>
                    {/* 刷新按钮 */}
                    <div>
                        <Button
                            variant="outline"
                            size="sm"
                            className="gap-2"
                            onClick={handleLoadCourses}
                            disabled={isLoadingCourses}
                        >
                            <RefreshCcw className="h-4 w-4 animate-spin-slow" />
                            刷新课程
                        </Button>
                    </div>
                </div>

                {isLoadingCourses ? (
                    <div className="flex justify-center items-center py-12">
                        <Loader2 className="h-10 w-10 animate-spin text-primary"/>
                        <span className="ml-3 text-lg text-muted-foreground">加载课程中...</span>
                    </div>
                ) : (
                    <>
                        <div className="space-y-2 sm:space-y-3">
                            {paginatedCourses().map((course) => (
                                <div key={course.courseId} className="relative">
                                    <Card
                                        className={`transition-all ${
                                            selectedCourses.includes(course.courseId) ? "border-primary bg-primary/5" : "hover:border-primary/50"
                                        }`}
                                    >
                                        <CardContent className="p-3 sm:p-4">
                                            <div className="flex items-start gap-3 sm:gap-4">
                                                <Checkbox
                                                    id={course.courseId}
                                                    checked={selectedCourses.includes(course.courseId)}
                                                    onCheckedChange={(checked) => handleSelectCourse(course.courseId, checked as boolean)}
                                                    className="mt-1"
                                                    disabled={isRunning}
                                                />
                                                <div className="flex-1 min-w-0">
                                                    <div
                                                        className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-2 sm:gap-4 mb-3">
                                                        <div className="flex-1 min-w-0">
                                                            <h4 className="font-semibold text-sm sm:text-base text-foreground mb-1 break-words">
                                                                {course.courseName}
                                                            </h4>
                                                            <div
                                                                className="flex flex-wrap items-center gap-2 sm:gap-3 text-xs sm:text-sm text-muted-foreground">
                                                                <span className="truncate">{course.instructor}</span>
                                                            </div>
                                                        </div>
                                                        <div className="text-left sm:text-right">
                                                            <div
                                                                className="text-base sm:text-lg font-semibold text-primary">{course.progress}%
                                                            </div>
                                                            <div className="text-xs text-muted-foreground">完成度</div>
                                                        </div>
                                                    </div>
                                                    <Progress value={course.progress} className="h-1.5 sm:h-2"/>
                                                </div>
                                            </div>
                                        </CardContent>
                                    </Card>
                                    {isRunning && (
                                        <div className="absolute inset-0 bg-gray-500/20 backdrop-blur-[1px] pointer-events-none z-10 rounded-lg flex flex-col items-center justify-center gap-1">
                                            <Lock className="h-6 w-6 text-gray-700" />
                                            <span className="text-sm font-medium text-gray-800">请先取消任务</span>
                                        </div>
                                    )}
                                </div>
                            ))}
                        </div>

                        {courseList.length === 0 && (
                            <Card>
                                <CardContent
                                    className="flex flex-col items-center justify-center py-8 sm:py-12 text-center px-4">
                                    <BookOpen className="h-10 w-10 sm:h-12 sm:w-12 text-muted-foreground mb-4"/>
                                    <h3 className="text-base sm:text-lg font-medium text-foreground mb-2">暂无课程</h3>
                                    <p className="text-xs sm:text-sm text-muted-foreground">该账号还没有关联任何课程</p>
                                </CardContent>
                            </Card>
                        )}
                        {/* 分页 */}
                        {totalPages > 1 && (
                            <div className="flex items-center justify-center gap-2 mt-6">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
                                    disabled={currentPage === 1}
                                >
                                    上一页
                                </Button>
                                <div className="flex items-center gap-1">
                                    {Array.from({ length: totalPages }, (_, i) => i + 1).map((page) => (
                                        <Button
                                            key={page}
                                            variant={currentPage === page ? "default" : "outline"}
                                            size="sm"
                                            onClick={() => setCurrentPage(page)}
                                        >
                                            {page}
                                        </Button>
                                    ))}
                                </div>
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
                                    disabled={currentPage === totalPages}
                                >
                                    下一页
                                </Button>
                            </div>
                        )}
                    </>
                )}
            </div>
        </motion.div>
    )
}
