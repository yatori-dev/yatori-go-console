"use client"

import {useCallback, useEffect, useState} from "react"
import {motion} from "framer-motion"
import {ArrowLeft, BookOpen, Check, Clock3, Eye, EyeOff, FileText, Gauge, KeyRound, Mail, Pause, Play, Plus, RefreshCcw, Settings2, Trash2, User} from "lucide-react"
import {Button} from "@/components/ui/button"
import {Card, CardContent, CardHeader} from "@/components/ui/card"
import {Badge} from "@/components/ui/badge"
import {Progress} from "@/components/ui/progress"
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs"
import {Input} from "@/components/ui/input"
import {Label} from "@/components/ui/label"
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select"
import {Switch} from "@/components/ui/switch"
import {Textarea} from "@/components/ui/textarea"
import {toast} from "@/components/ui/use-toast"
import {getCourseList} from "@/api/courseApi"
import {getAccountDetail, getAccountLogs, startBrushForUid, stopBrushForUid, updateAccount} from "@/api/accountApi"
import {getPlatformName} from "@/utils/platformUtils"
import type {Account, Course} from "@/components/account-list"

type AccountDetailProps = {
    account: Account
    onBack: () => void
    onUpdated?: () => Promise<void> | void
}

type EditableForm = {
    uid: string
    accountType: string
    url: string
    account: string
    password: string
    remarkName: string
    informEmailsText: string
    includeCourses: string[]
    excludeCourses: string[]
    studyTime: string
    autoRunStartTime: string
    autoRunEndTime: string
    videoModel: string
    autoExam: string
    examAutoSubmit: string
    cxNode: string
    cxChapterTestSw: string
    cxWorkSw: string
    cxExamSw: string
    shuffleSw: string
}

type Option = {
    value: string
    label: string
    description: string
}

const videoModelOptions: Option[] = [
    {value: "0", label: "不刷视频", description: "不执行视频学习，适合只维护账号配置时使用。"},
    {value: "1", label: "普通模式", description: "常规学习模式，整体最稳妥，适合大多数账号。"},
    {value: "2", label: "暴力模式", description: "更激进，速度更快，但部分平台更容易触发风控。"},
    {value: "3", label: "去红模式", description: "学习通多任务点模式，可配合并发节点一起使用。"},
]

const autoExamOptions: Option[] = [
    {value: "0", label: "不自动答题", description: "不自动处理考试或作业。"},
    {value: "1", label: "AI 自动答题", description: "使用当前 AI 配置自动答题。"},
    {value: "2", label: "外部题库答题", description: "通过外部题库接口自动答题。"},
    {value: "3", label: "学习通内置 AI", description: "调用学习通内置 AI 处理答题，仅适用于学习通。"},
]

const cxNodeOptions: Option[] = [
    {value: "0", label: "默认 3 个节点", description: "使用系统默认并发数，适合大多数用户。"},
    {value: "1", label: "1 个节点", description: "最稳妥，但处理速度最慢。"},
    {value: "2", label: "2 个节点", description: "比单节点更快，风险较低。"},
    {value: "3", label: "3 个节点", description: "常用平衡配置。"},
    {value: "4", label: "4 个节点", description: "速度更快，但更容易触发风控。"},
    {value: "5", label: "5 个节点", description: "高并发模式，仅建议熟悉风险时使用。"},
    {value: "-1", label: "无限制模式", description: "会大量重复登录，存在封号或封 IP 风险。"},
]

const enabledValue = "1"
const disabledValue = "0"

const examAutoSubmitOptions: Option[] = [
    {value: disabledValue, label: "手动交卷", description: "答题完成后不自动提交，适合先人工确认。"},
    {value: enabledValue, label: "自动交卷", description: "答题完成后自动提交试卷，适合全自动运行。"},
]

const switchOptions: Option[] = [
    {value: disabledValue, label: "关闭", description: "当前功能不启用。"},
    {value: enabledValue, label: "开启", description: "当前功能已启用。"},
]

function safeParseUserConfig(userConfigJson: string) {
    try {
        return JSON.parse(userConfigJson)
    } catch {
        return {}
    }
}

function formatProgress(value: number | undefined) {
    if (typeof value !== "number" || Number.isNaN(value)) {
        return 0
    }
    return Math.max(0, Math.min(100, Math.round(value)))
}

function maskText(text: string) {
    return "*".repeat(text.length || 8)
}

function toStringArray(value: unknown) {
    return Array.isArray(value)
        ? value.map((item) => String(item).trim()).filter(Boolean)
        : []
}

function parseLineText(value: string) {
    return value
        .split(/\n|,|，/)
        .map((item) => item.trim())
        .filter(Boolean)
}

function toNumberString(value: unknown) {
    return value === undefined || value === null ? "" : String(value)
}

function isValidClockTime(value: string) {
    return /^(?:[01]\d|2[0-3]):[0-5]\d$/.test(value)
}

function getOptionLabel(options: Option[], value: unknown, fallback = "未设置") {
    const match = options.find((item) => item.value === String(value))
    return match?.label ?? fallback
}

function getOptionDescription(options: Option[], value: unknown) {
    return options.find((item) => item.value === String(value))?.description ?? ""
}

function getSwitchLabel(value: unknown) {
    return String(value) === enabledValue ? "开启" : "关闭"
}

function renderOptionHint(options: Option[], value: string) {
    const description = getOptionDescription(options, value)
    if (!description) {
        return null
    }
    return <p className="text-xs leading-5 text-muted-foreground">{description}</p>
}

function buildUserConfigJson(user: any) {
    return JSON.stringify({
        accountType: user.accountType,
        url: user.url || "",
        account: user.account,
        password: user.password,
        isProxy: user.isProxy ?? 0,
        informEmails: user.informEmails ?? [],
        coursesCustom: user.coursesCustom ?? {},
        remarkName: user.remarkName ?? "",
    })
}

function buildEditableForm(user: any): EditableForm {
    const coursesCustom = user?.coursesCustom ?? {}
    return {
        uid: user?.uid ?? "",
        accountType: user?.accountType ?? "",
        url: user?.url ?? "",
        account: user?.account ?? "",
        password: user?.password ?? "",
        remarkName: user?.remarkName ?? "",
        informEmailsText: toStringArray(user?.informEmails).join("\n"),
        includeCourses: toStringArray(coursesCustom?.includeCourses),
        excludeCourses: toStringArray(coursesCustom?.excludeCourses),
        studyTime: coursesCustom?.studyTime ?? "",
        autoRunStartTime: coursesCustom?.autoRunStartTime ?? "",
        autoRunEndTime: coursesCustom?.autoRunEndTime ?? "",
        videoModel: toNumberString(coursesCustom?.videoModel),
        autoExam: toNumberString(coursesCustom?.autoExam),
        examAutoSubmit: toNumberString(coursesCustom?.examAutoSubmit),
        cxNode: toNumberString(coursesCustom?.cxNode),
        cxChapterTestSw: toNumberString(coursesCustom?.cxChapterTestSw),
        cxWorkSw: toNumberString(coursesCustom?.cxWorkSw),
        cxExamSw: toNumberString(coursesCustom?.cxExamSw),
        shuffleSw: toNumberString(coursesCustom?.shuffleSw),
    }
}

export function AccountDetail({account, onBack, onUpdated}: AccountDetailProps) {
    const [activeTab, setActiveTab] = useState("courses")
    const [courseList, setCourseList] = useState<Course[]>([])
    const [detailAccount, setDetailAccount] = useState<Account>(account)
    const [accountConfig, setAccountConfig] = useState<any>(() => safeParseUserConfig(account.userConfigJson))
    const [form, setForm] = useState<EditableForm>(() => buildEditableForm(safeParseUserConfig(account.userConfigJson)))
    const [logs, setLogs] = useState("")
    const [lastRefreshAt, setLastRefreshAt] = useState("")
    const [showAccount, setShowAccount] = useState(true)
    const [showPassword, setShowPassword] = useState(false)
    const [isLoadingDetail, setIsLoadingDetail] = useState(false)
    const [isLoadingCourses, setIsLoadingCourses] = useState(false)
    const [isLoadingLogs, setIsLoadingLogs] = useState(false)
    const [isProcessing, setIsProcessing] = useState(false)
    const [isSaving, setIsSaving] = useState(false)

    const setToggleField = (field: keyof EditableForm, checked: boolean) => {
        updateForm(field, checked ? enabledValue : disabledValue)
    }

    const updateCourseField = (field: "includeCourses" | "excludeCourses", index: number, value: string) => {
        setForm((prev) => {
            const nextItems = [...prev[field]]
            nextItems[index] = value
            return {...prev, [field]: nextItems}
        })
    }

    const addCourseField = (field: "includeCourses" | "excludeCourses") => {
        setForm((prev) => ({...prev, [field]: [...prev[field], ""]}))
    }

    const removeCourseField = (field: "includeCourses" | "excludeCourses", index: number) => {
        setForm((prev) => ({...prev, [field]: prev[field].filter((_, itemIndex) => itemIndex !== index)}))
    }

    const handleLoadDetail = useCallback(async () => {
        try {
            setIsLoadingDetail(true)
            const response = await getAccountDetail(account.uid)
            if (response.code !== 200 || !response.data?.user) {
                toast({
                    title: "加载账号详情失败",
                    description: response.msg || "未获取到账号详情",
                    variant: "destructive",
                })
                return
            }

            const user = response.data.user
            const userConfigJson = user.userConfigJson ?? buildUserConfigJson(user)
            setDetailAccount((prev) => ({
                ...prev,
                uid: user.uid,
                accountType: user.accountType,
                url: user.url,
                account: user.account,
                password: user.password,
                isRunning: user.isRunning ?? prev.isRunning,
                userConfigJson,
            }))
            setAccountConfig(safeParseUserConfig(userConfigJson))
            setForm(buildEditableForm(user))
        } catch (error) {
            console.error("账号详情加载失败:", error)
            toast({
                title: "网络错误",
                description: "无法加载账号详情，请稍后重试",
                variant: "destructive",
            })
        } finally {
            setIsLoadingDetail(false)
        }
    }, [account.uid])

    const handleLoadCourses = useCallback(async () => {
        try {
            setIsLoadingCourses(true)
            const response = await getCourseList(account.uid)
            if (response.code === 200) {
                setCourseList(response.data?.courseList ?? [])
                setLastRefreshAt(new Date().toLocaleString("zh-CN"))
                return
            }
            toast({
                title: "课程加载失败",
                description: response.message || "课程列表加载失败",
                variant: "destructive",
            })
        } catch (error) {
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

    const handleLoadLogs = useCallback(async (silent = false) => {
        try {
            if (!silent) {
                setIsLoadingLogs(true)
            }
            const response = await getAccountLogs(account.uid)
            if (response.code === 200) {
                setLogs(response.data?.logs ?? "")
                return
            }
            if (!silent) {
                toast({
                    title: "日志加载失败",
                    description: response.message || "未获取到账号日志",
                    variant: "destructive",
                })
            }
        } catch (error) {
            console.error("日志加载失败:", error)
            if (!silent) {
                toast({
                    title: "网络错误",
                    description: "无法加载账号日志，请稍后重试",
                    variant: "destructive",
                })
            }
        } finally {
            if (!silent) {
                setIsLoadingLogs(false)
            }
        }
    }, [account.uid])

    useEffect(() => {
        handleLoadDetail()
        handleLoadLogs()
    }, [handleLoadDetail, handleLoadLogs])

    useEffect(() => {
        if (activeTab === "courses") {
            handleLoadCourses()
        }
    }, [activeTab, handleLoadCourses])

    useEffect(() => {
        const timer = window.setInterval(() => {
            handleLoadLogs(true)
        }, 5000)
        return () => window.clearInterval(timer)
    }, [handleLoadLogs])

    const handleToggleRunning = async () => {
        try {
            setIsProcessing(true)
            const response = detailAccount.isRunning
                ? await stopBrushForUid(detailAccount.uid)
                : await startBrushForUid(detailAccount.uid)

            if (response.code !== 200) {
                toast({
                    title: detailAccount.isRunning ? "停止失败" : "启动失败",
                    description: response.message,
                    variant: "destructive",
                })
                return
            }

            setDetailAccount((prev) => ({...prev, isRunning: !prev.isRunning}))
            toast({
                title: detailAccount.isRunning ? "已停止" : "已启动",
                description: response.message,
            })
        } catch (error) {
            console.error("切换运行状态失败:", error)
            toast({
                title: "操作失败",
                description: "无法连接到服务器，请稍后重试",
                variant: "destructive",
            })
        } finally {
            setIsProcessing(false)
        }
    }

    const handleSave = async () => {
        try {
            setIsSaving(true)
            const cxNodeValue = Number(form.cxNode || 0)
            if (!Number.isInteger(cxNodeValue) || cxNodeValue < -1) {
                toast({
                    title: "保存失败",
                    description: "并发节点只能是 -1 或不小于 0 的整数",
                    variant: "destructive",
                })
                return
            }

            const autoRunStartTime = form.autoRunStartTime.trim()
            const autoRunEndTime = form.autoRunEndTime.trim()
            if ((autoRunStartTime && !autoRunEndTime) || (!autoRunStartTime && autoRunEndTime)) {
                toast({
                    title: "保存失败",
                    description: "自动执行时间段需要同时设置开始时间和结束时间",
                    variant: "destructive",
                })
                return
            }
            if ((autoRunStartTime && !isValidClockTime(autoRunStartTime)) || (autoRunEndTime && !isValidClockTime(autoRunEndTime))) {
                toast({
                    title: "保存失败",
                    description: "自动执行时间格式必须为 HH:MM",
                    variant: "destructive",
                })
                return
            }

            const response = await updateAccount({
                uid: form.uid,
                accountType: form.accountType,
                url: form.url.trim(),
                account: form.account.trim(),
                password: form.password,
                remarkName: form.remarkName.trim(),
                informEmails: parseLineText(form.informEmailsText),
                coursesCustom: {
                    studyTime: form.studyTime.trim(),
                    autoRunStartTime,
                    autoRunEndTime,
                    videoModel: Number(form.videoModel || 0),
                    autoExam: Number(form.autoExam || 0),
                    examAutoSubmit: Number(form.examAutoSubmit || 0),
                    cxNode: cxNodeValue,
                    cxChapterTestSw: Number(form.cxChapterTestSw || 0),
                    cxWorkSw: Number(form.cxWorkSw || 0),
                    cxExamSw: Number(form.cxExamSw || 0),
                    shuffleSw: Number(form.shuffleSw || 0),
                    includeCourses: form.includeCourses.map((item) => item.trim()).filter(Boolean),
                    excludeCourses: form.excludeCourses.map((item) => item.trim()).filter(Boolean),
                },
            })

            if (response.code !== 200) {
                toast({
                    title: "保存失败",
                    description: response.message || response.msg || "账号配置保存失败",
                    variant: "destructive",
                })
                return
            }

            await handleLoadDetail()
            if (onUpdated) {
                await onUpdated()
            }
            toast({
                title: "保存成功",
                description: "账号配置已保存",
            })
        } catch (error) {
            console.error("保存账号配置失败:", error)
            toast({
                title: "保存失败",
                description: "无法连接到服务器，请稍后重试",
                variant: "destructive",
            })
        } finally {
            setIsSaving(false)
        }
    }

    const handleRefreshAll = () => {
        handleLoadDetail()
        handleLoadLogs()
        if (activeTab === "courses") {
            handleLoadCourses()
        }
    }

    const updateForm = (field: keyof EditableForm, value: string) => {
        setForm((prev) => ({...prev, [field]: value}))
    }

    const courseProgressList = courseList.map((course) => ({
        ...course,
        normalizedProgress: formatProgress(course.progress),
    }))
    const completedCourses = courseProgressList.filter((course) => course.normalizedProgress >= 100).length
    const averageProgress = courseProgressList.length
        ? Math.round(courseProgressList.reduce((sum, course) => sum + course.normalizedProgress, 0) / courseProgressList.length)
        : 0
    const inProgressCourses = courseProgressList.filter((course) => course.normalizedProgress > 0 && course.normalizedProgress < 100).length
    const configCourses = accountConfig?.coursesCustom ?? {}
    const currentVideoModeLabel = getOptionLabel(videoModelOptions, configCourses.videoModel, "普通模式")
    const currentAutoExamLabel = getOptionLabel(autoExamOptions, form.autoExam, "不自动答题")
    const currentCxNodeLabel = getOptionLabel(cxNodeOptions, form.cxNode, "默认 3 个节点")

    return (
        <motion.div initial={{opacity: 0, y: 20}} animate={{opacity: 1, y: 0}} transition={{duration: 0.4}} className="space-y-6">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <Button variant="ghost" onClick={onBack} className="w-fit gap-2 text-sm sm:text-base">
                    <ArrowLeft className="h-4 w-4"/>
                    返回账号列表
                </Button>
                <div className="flex flex-wrap gap-3">
                    <Button onClick={handleSave} disabled={isSaving || isLoadingDetail} className="gap-2">
                        <Check className="h-4 w-4"/>
                        {isSaving ? "保存中..." : "保存配置"}
                    </Button>
                    <Button variant={detailAccount.isRunning ? "destructive" : "default"} className="gap-2" onClick={handleToggleRunning} disabled={isProcessing || isLoadingDetail}>
                        {detailAccount.isRunning ? <Pause className="h-4 w-4"/> : <Play className="h-4 w-4"/>}
                        {isProcessing ? "处理中..." : detailAccount.isRunning ? "停止任务" : "开始任务"}
                    </Button>
                    <Button variant="outline" className="gap-2" onClick={handleRefreshAll} disabled={isLoadingDetail || isLoadingCourses || isLoadingLogs}>
                        <RefreshCcw className="h-4 w-4"/>
                        刷新
                    </Button>
                </div>
            </div>

            <Card className="overflow-hidden border-0 shadow-lg shadow-primary/5">
                <div className="bg-[linear-gradient(135deg,rgba(68,97,242,0.10),rgba(15,196,140,0.08),rgba(255,255,255,0.95))]">
                    <CardHeader className="pb-8">
                        <div className="flex flex-col gap-5 lg:flex-row lg:items-center">
                            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/10 text-primary ring-1 ring-primary/15">
                                <User className="h-8 w-8"/>
                            </div>
                            <div className="min-w-0 flex-1 space-y-3">
                                <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                                    <div className="min-w-0">
                                        <div className="flex items-center gap-2">
                                            <h2 className="truncate text-2xl font-semibold tracking-tight sm:text-3xl">
                                                {showAccount ? detailAccount.account : maskText(detailAccount.account)}
                                            </h2>
                                            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setShowAccount((prev) => !prev)}>
                                                {showAccount ? <Eye className="h-3.5 w-3.5"/> : <EyeOff className="h-3.5 w-3.5"/>}
                                            </Button>
                                        </div>
                                        <p className="mt-1 text-sm text-muted-foreground">
                                            {getPlatformName(detailAccount.accountType)}
                                            {detailAccount.url ? ` · ${detailAccount.url}` : ""}
                                        </p>
                                    </div>
                                    <div className="flex flex-wrap gap-2">
                                        <Badge variant={detailAccount.isRunning ? "default" : "secondary"}>{detailAccount.isRunning ? "运行中" : "未运行"}</Badge>
                                        <Badge variant="outline">平均进度 {averageProgress}%</Badge>
                                        <Badge variant="outline">{form.remarkName || "未设置备注"}</Badge>
                                    </div>
                                </div>
                                <div className="flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
                                    <div className="flex items-center gap-2">
                                        <KeyRound className="h-4 w-4"/>
                                        <span className="font-mono">{showPassword ? detailAccount.password : maskText(detailAccount.password)}</span>
                                        <Button variant="ghost" size="icon" className="h-6 w-6" onClick={() => setShowPassword((prev) => !prev)}>
                                            {showPassword ? <Eye className="h-3 w-3"/> : <EyeOff className="h-3 w-3"/>}
                                        </Button>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <Clock3 className="h-4 w-4"/>
                                        <span>{lastRefreshAt || "尚未刷新"}</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </CardHeader>
                </div>
            </Card>

            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                <Card className="border-0 bg-[linear-gradient(180deg,rgba(68,97,242,0.10),rgba(255,255,255,1))] shadow-sm">
                    <CardContent className="flex items-center justify-between p-5">
                        <div>
                            <p className="text-xs uppercase tracking-[0.24em] text-muted-foreground">课程总数</p>
                            <p className="mt-2 text-3xl font-semibold">{courseProgressList.length}</p>
                        </div>
                        <BookOpen className="h-8 w-8 text-primary"/>
                    </CardContent>
                </Card>
                <Card className="border-0 bg-[linear-gradient(180deg,rgba(15,196,140,0.10),rgba(255,255,255,1))] shadow-sm">
                    <CardContent className="flex items-center justify-between p-5">
                        <div>
                            <p className="text-xs uppercase tracking-[0.24em] text-muted-foreground">已完成</p>
                            <p className="mt-2 text-3xl font-semibold">{completedCourses}</p>
                        </div>
                        <Check className="h-8 w-8 text-emerald-600"/>
                    </CardContent>
                </Card>
                <Card className="border-0 bg-[linear-gradient(180deg,rgba(255,184,77,0.16),rgba(255,255,255,1))] shadow-sm">
                    <CardContent className="flex items-center justify-between p-5">
                        <div>
                            <p className="text-xs uppercase tracking-[0.24em] text-muted-foreground">进行中</p>
                            <p className="mt-2 text-3xl font-semibold">{inProgressCourses}</p>
                        </div>
                        <Gauge className="h-8 w-8 text-amber-600"/>
                    </CardContent>
                </Card>
                <Card className="border-0 bg-[linear-gradient(180deg,rgba(95,129,255,0.10),rgba(255,255,255,1))] shadow-sm">
                    <CardContent className="flex items-center justify-between p-5">
                        <div>
                            <p className="text-xs uppercase tracking-[0.24em] text-muted-foreground">当前视频模式</p>
                            <p className="mt-2 text-2xl font-semibold">{currentVideoModeLabel}</p>
                        </div>
                        <Settings2 className="h-8 w-8 text-primary"/>
                    </CardContent>
                </Card>
                <Card className="border-0 bg-[linear-gradient(180deg,rgba(255,122,69,0.12),rgba(255,255,255,1))] shadow-sm">
                    <CardContent className="flex items-center justify-between p-5">
                        <div>
                            <p className="text-xs uppercase tracking-[0.24em] text-muted-foreground">自动答题策略</p>
                            <p className="mt-2 text-2xl font-semibold">{currentAutoExamLabel}</p>
                            <p className="mt-1 text-sm text-muted-foreground">{currentCxNodeLabel}</p>
                        </div>
                        <Gauge className="h-8 w-8 text-orange-500"/>
                    </CardContent>
                </Card>
            </div>

            <Tabs value={activeTab} onValueChange={setActiveTab} className="gap-4">
                <TabsList className="h-auto w-full justify-start rounded-2xl bg-muted/70 p-1.5">
                    <TabsTrigger value="courses" className="gap-2 px-4 py-2">
                        <BookOpen className="h-4 w-4"/>
                        课程进度
                    </TabsTrigger>
                    <TabsTrigger value="config" className="gap-2 px-4 py-2">
                        <Settings2 className="h-4 w-4"/>
                        在线编辑
                    </TabsTrigger>
                    <TabsTrigger value="logs" className="gap-2 px-4 py-2">
                        <FileText className="h-4 w-4"/>
                        账号日志
                    </TabsTrigger>
                </TabsList>

                <TabsContent value="courses" className="mt-0">
                    <Card className="border-0 shadow-sm">
                        <CardHeader className="border-b bg-muted/20">
                            <h3 className="text-lg font-semibold">课程完成进度</h3>
                            <p className="text-sm text-muted-foreground">按课程查看当前账号的完成情况。</p>
                        </CardHeader>
                        <CardContent className="p-4 sm:p-6">
                            {isLoadingCourses ? (
                                <div className="rounded-2xl border border-dashed p-10 text-center text-sm text-muted-foreground">正在拉取课程进度...</div>
                            ) : courseProgressList.length === 0 ? (
                                <div className="rounded-2xl border border-dashed p-10 text-center text-sm text-muted-foreground">暂无课程数据</div>
                            ) : (
                                <div className="grid gap-4 lg:grid-cols-2">
                                    {courseProgressList.map((course) => (
                                        <div key={course.courseId} className="rounded-2xl border bg-[linear-gradient(180deg,rgba(255,255,255,1),rgba(246,248,255,1))] p-4 shadow-sm">
                                            <div className="flex items-start justify-between gap-3">
                                                <div className="min-w-0">
                                                    <h4 className="truncate text-base font-semibold">{course.courseName}</h4>
                                                    <p className="mt-1 text-sm text-muted-foreground">{course.instructor || "暂无教师信息"}</p>
                                                </div>
                                                <Badge variant={course.normalizedProgress >= 100 ? "default" : "secondary"}>
                                                    {course.normalizedProgress >= 100 ? "已完成" : "学习中"}
                                                </Badge>
                                            </div>
                                            <div className="mt-4 space-y-3">
                                                <div className="flex items-center justify-between text-sm">
                                                    <span className="text-muted-foreground">完成率</span>
                                                    <span className="font-medium">{course.normalizedProgress}%</span>
                                                </div>
                                                <Progress value={course.normalizedProgress} className="h-2.5"/>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="config" className="mt-0">
                    <div className="grid gap-4 lg:grid-cols-[1.1fr_0.9fr]">
                        <Card className="border-0 shadow-sm">
                            <CardHeader className="border-b bg-muted/20">
                                <h3 className="text-lg font-semibold">基础信息</h3>
                                <p className="text-sm text-muted-foreground">这里修改的内容会直接保存到 `config.yaml`。</p>
                            </CardHeader>
                            <CardContent className="grid gap-4 p-4 sm:grid-cols-2 sm:p-6">
                                <div className="space-y-2 rounded-2xl border p-4">
                                    <Label htmlFor="remarkName">备注名</Label>
                                    <Input id="remarkName" value={form.remarkName} onChange={(e) => updateForm("remarkName", e.target.value)}/>
                                </div>
                                <div className="space-y-2 rounded-2xl border p-4">
                                    <Label htmlFor="url">平台地址</Label>
                                    <Input id="url" value={form.url} onChange={(e) => updateForm("url", e.target.value)}/>
                                </div>
                                <div className="space-y-2 rounded-2xl border p-4">
                                    <Label htmlFor="accountValue">账号</Label>
                                    <Input id="accountValue" value={form.account} onChange={(e) => updateForm("account", e.target.value)}/>
                                </div>
                                <div className="space-y-2 rounded-2xl border p-4">
                                    <Label htmlFor="passwordValue">密码</Label>
                                    <Input id="passwordValue" value={form.password} onChange={(e) => updateForm("password", e.target.value)}/>
                                </div>
                                <div className="space-y-2 rounded-2xl border p-4">
                                    <Label htmlFor="studyTime">学习时段</Label>
                                    <Input id="studyTime" value={form.studyTime} onChange={(e) => updateForm("studyTime", e.target.value)}/>
                                    <p className="text-xs leading-5 text-muted-foreground">按平台支持的时间范围填写，用于限制刷课时间段。</p>
                                </div>
                                <div className="space-y-4 rounded-2xl border p-4 sm:col-span-2">
                                    <div>
                                        <p className="text-sm font-medium">自动执行时间段</p>
                                        <p className="mt-1 text-xs leading-5 text-muted-foreground">设置后系统会每天仅在这个时间段内自动执行。支持跨天，例如 22:00 到 02:00。</p>
                                    </div>
                                    <div className="grid gap-4 sm:grid-cols-2">
                                        <div className="space-y-2 rounded-2xl bg-muted/30 p-4">
                                            <Label htmlFor="autoRunStartTime">开始时间</Label>
                                            <Input
                                                id="autoRunStartTime"
                                                type="time"
                                                value={form.autoRunStartTime}
                                                onChange={(e) => updateForm("autoRunStartTime", e.target.value)}
                                            />
                                        </div>
                                        <div className="space-y-2 rounded-2xl bg-muted/30 p-4">
                                            <Label htmlFor="autoRunEndTime">结束时间</Label>
                                            <Input
                                                id="autoRunEndTime"
                                                type="time"
                                                value={form.autoRunEndTime}
                                                onChange={(e) => updateForm("autoRunEndTime", e.target.value)}
                                            />
                                        </div>
                                    </div>
                                </div>
                                <div className="space-y-4 rounded-2xl border p-4 sm:col-span-2">
                                    <div>
                                        <p className="text-sm font-medium">学习策略</p>
                                        <p className="mt-1 text-xs leading-5 text-muted-foreground">这里用中文选项替代原来的数字配置，保存后仍会写回原始数值。</p>
                                    </div>
                                    <div className="grid gap-4 sm:grid-cols-2">
                                        <div className="space-y-2 rounded-2xl bg-muted/30 p-4">
                                            <Label htmlFor="videoModel">视频模式</Label>
                                            <Select value={form.videoModel || "0"} onValueChange={(value) => updateForm("videoModel", value)}>
                                                <SelectTrigger id="videoModel">
                                                    <SelectValue placeholder="请选择视频模式"/>
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {videoModelOptions.map((option) => (
                                                        <SelectItem key={option.value} value={option.value}>
                                                            {option.label}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                            {renderOptionHint(videoModelOptions, form.videoModel || "0")}
                                        </div>
                                        <div className="space-y-2 rounded-2xl bg-muted/30 p-4">
                                            <Label htmlFor="autoExam">答题方式</Label>
                                            <Select value={form.autoExam || "0"} onValueChange={(value) => updateForm("autoExam", value)}>
                                                <SelectTrigger id="autoExam">
                                                    <SelectValue placeholder="请选择答题方式"/>
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {autoExamOptions.map((option) => (
                                                        <SelectItem key={option.value} value={option.value}>
                                                            {option.label}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                            {renderOptionHint(autoExamOptions, form.autoExam || "0")}
                                        </div>
                                        <div className="space-y-2 rounded-2xl bg-muted/30 p-4">
                                            <Label htmlFor="examAutoSubmit">交卷方式</Label>
                                            <Select value={form.examAutoSubmit || disabledValue} onValueChange={(value) => updateForm("examAutoSubmit", value)}>
                                                <SelectTrigger id="examAutoSubmit">
                                                    <SelectValue placeholder="请选择交卷方式"/>
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {examAutoSubmitOptions.map((option) => (
                                                        <SelectItem key={option.value} value={option.value}>
                                                            {option.label}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                            {renderOptionHint(examAutoSubmitOptions, form.examAutoSubmit || disabledValue)}
                                        </div>
                                        <div className="space-y-2 rounded-2xl bg-muted/30 p-4">
                                            <Label htmlFor="shuffleSw">课程顺序</Label>
                                            <div className="flex items-center justify-between rounded-xl border bg-background px-3 py-3">
                                                <div>
                                                    <p className="text-sm font-medium">{form.shuffleSw === enabledValue ? "随机学习" : "按原顺序学习"}</p>
                                                    <p className="mt-1 text-xs text-muted-foreground">开启后会打乱课程执行顺序，适合降低固定轨迹。</p>
                                                </div>
                                                <Switch
                                                    id="shuffleSw"
                                                    checked={form.shuffleSw === enabledValue}
                                                    onCheckedChange={(checked) => setToggleField("shuffleSw", checked)}
                                                />
                                            </div>
                                        </div>
                                    </div>
                                </div>
                                <div className="space-y-4 rounded-2xl border p-4 sm:col-span-2">
                                    <div className="flex items-start justify-between gap-3">
                                        <div>
                                            <p className="text-sm font-medium">学习通专属设置</p>
                                            <p className="mt-1 text-xs leading-5 text-muted-foreground">仅学习通相关账号会用到这些配置，其他平台通常可以保持默认。</p>
                                        </div>
                                        <Badge variant="outline">学习通</Badge>
                                    </div>
                                    <div className="rounded-2xl bg-muted/30 p-4">
                                        <Label htmlFor="cxNode">并发节点</Label>
                                        <Select value={form.cxNode || "3"} onValueChange={(value) => updateForm("cxNode", value)}>
                                            <SelectTrigger id="cxNode" className="mt-2">
                                                <SelectValue placeholder="请选择并发节点"/>
                                            </SelectTrigger>
                                            <SelectContent>
                                                {cxNodeOptions.map((option) => (
                                                    <SelectItem key={option.value} value={option.value}>
                                                        {option.label}
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                        <div className="mt-2 space-y-2">
                                            {renderOptionHint(cxNodeOptions, form.cxNode || "3")}
                                            {form.cxNode === "-1" ? (
                                                <p className="rounded-xl border border-destructive/30 bg-destructive/5 px-3 py-2 text-xs leading-5 text-destructive">
                                                    无限模式风险很高，会大量重复登录，可能触发封号或封 IP。
                                                </p>
                                            ) : null}
                                        </div>
                                    </div>
                                    <div className="grid gap-4 sm:grid-cols-3">
                                        <div className="space-y-3 rounded-2xl bg-muted/30 p-4">
                                            <div className="space-y-1">
                                                <Label htmlFor="cxChapterTestSw">章测</Label>
                                                <p className="text-xs text-muted-foreground">{getSwitchLabel(form.cxChapterTestSw)}</p>
                                            </div>
                                            <Switch
                                                id="cxChapterTestSw"
                                                checked={form.cxChapterTestSw === enabledValue}
                                                onCheckedChange={(checked) => setToggleField("cxChapterTestSw", checked)}
                                            />
                                            {renderOptionHint(switchOptions, form.cxChapterTestSw || disabledValue)}
                                        </div>
                                        <div className="space-y-3 rounded-2xl bg-muted/30 p-4">
                                            <div className="space-y-1">
                                                <Label htmlFor="cxWorkSw">作业</Label>
                                                <p className="text-xs text-muted-foreground">{getSwitchLabel(form.cxWorkSw)}</p>
                                            </div>
                                            <Switch
                                                id="cxWorkSw"
                                                checked={form.cxWorkSw === enabledValue}
                                                onCheckedChange={(checked) => setToggleField("cxWorkSw", checked)}
                                            />
                                            {renderOptionHint(switchOptions, form.cxWorkSw || disabledValue)}
                                        </div>
                                        <div className="space-y-3 rounded-2xl bg-muted/30 p-4">
                                            <div className="space-y-1">
                                                <Label htmlFor="cxExamSw">考试</Label>
                                                <p className="text-xs text-muted-foreground">{getSwitchLabel(form.cxExamSw)}</p>
                                            </div>
                                            <Switch
                                                id="cxExamSw"
                                                checked={form.cxExamSw === enabledValue}
                                                onCheckedChange={(checked) => setToggleField("cxExamSw", checked)}
                                            />
                                            {renderOptionHint(switchOptions, form.cxExamSw || disabledValue)}
                                        </div>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>

                        <Card className="border-0 shadow-sm">
                            <CardHeader className="border-b bg-muted/20">
                                <h3 className="text-lg font-semibold">邮箱与课程筛选</h3>
                                <p className="text-sm text-muted-foreground">邮箱支持多行编辑，课程筛选支持逐条新增和删除。</p>
                            </CardHeader>
                            <CardContent className="space-y-4 p-4 sm:p-6">
                                <div className="rounded-2xl border p-4">
                                    <div className="flex items-center gap-2">
                                        <Mail className="h-4 w-4 text-primary"/>
                                        <p className="font-medium">通知邮箱</p>
                                    </div>
                                    <Textarea className="mt-3 min-h-[120px]" value={form.informEmailsText} onChange={(e) => updateForm("informEmailsText", e.target.value)} placeholder="每行一个邮箱"/>
                                </div>
                                <div className="rounded-2xl border p-4">
                                    <p className="font-medium">包含课程</p>
                                    <p className="mt-1 text-xs leading-5 text-muted-foreground">只学习这里列出的课程，不填则默认不过滤。</p>
                                    <div className="mt-3 space-y-3">
                                        {form.includeCourses.map((course, index) => (
                                            <div key={`include-${index}`} className="flex items-center gap-2">
                                                <Input
                                                    value={course}
                                                    onChange={(e) => updateCourseField("includeCourses", index, e.target.value)}
                                                    placeholder="输入一个课程名"
                                                />
                                                <Button type="button" variant="outline" size="icon" onClick={() => removeCourseField("includeCourses", index)}>
                                                    <Trash2 className="h-4 w-4"/>
                                                </Button>
                                            </div>
                                        ))}
                                        <Button type="button" variant="outline" className="gap-2" onClick={() => addCourseField("includeCourses")}>
                                            <Plus className="h-4 w-4"/>
                                            新增包含课程
                                        </Button>
                                    </div>
                                </div>
                                <div className="rounded-2xl border p-4">
                                    <p className="font-medium">排除课程</p>
                                    <p className="mt-1 text-xs leading-5 text-muted-foreground">这里列出的课程会被跳过，优先级低于“包含课程”。</p>
                                    <div className="mt-3 space-y-3">
                                        {form.excludeCourses.map((course, index) => (
                                            <div key={`exclude-${index}`} className="flex items-center gap-2">
                                                <Input
                                                    value={course}
                                                    onChange={(e) => updateCourseField("excludeCourses", index, e.target.value)}
                                                    placeholder="输入一个课程名"
                                                />
                                                <Button type="button" variant="outline" size="icon" onClick={() => removeCourseField("excludeCourses", index)}>
                                                    <Trash2 className="h-4 w-4"/>
                                                </Button>
                                            </div>
                                        ))}
                                        <Button type="button" variant="outline" className="gap-2" onClick={() => addCourseField("excludeCourses")}>
                                            <Plus className="h-4 w-4"/>
                                            新增排除课程
                                        </Button>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    </div>
                </TabsContent>

                <TabsContent value="logs" className="mt-0">
                    <Card className="border-0 shadow-sm">
                        <CardHeader className="border-b bg-muted/20">
                            <h3 className="text-lg font-semibold">账号日志</h3>
                            <p className="text-sm text-muted-foreground">当前按账号关键字过滤日志，每 5 秒自动刷新一次。</p>
                        </CardHeader>
                        <CardContent className="p-0">
                            <pre className="max-h-[34rem] overflow-auto whitespace-pre-wrap break-words bg-slate-950/95 p-5 font-mono text-xs leading-6 text-slate-100">
                                {isLoadingLogs ? "日志加载中..." : logs || "暂无该账号相关日志"}
                            </pre>
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </motion.div>
    )
}
