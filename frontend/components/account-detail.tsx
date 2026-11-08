"use client"

import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import { motion } from "framer-motion"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Checkbox } from "@/components/ui/checkbox"
import { Progress } from "@/components/ui/progress"
import { Switch } from "@/components/ui/switch"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import {
    ArrowLeft,
    Play,
    StopCircle,
    User,
    Check,
    BookOpen,
    Loader2,
    Eye,
    EyeOff,
    KeyRound,
    Leaf,
    Trash2,
    Lock,
    RefreshCcw,
    ShieldCheck,
    ListChecks,
    SlidersHorizontal,
    BarChart3,
    Terminal,
} from "lucide-react"
import type { Account, AccountConfig, AccountStatus, LoginCodeChallenge } from "@/api/accountApi"
import {
    checkAccountLogin,
    getAccountLog,
    getAccountLoginCode,
    getAccountStatus,
    normalizeConfig,
    sendAccountLoginCodeSms,
    startBrushForUid,
    stopBrushForUid,
    submitAccountLoginCode,
    updateAccount,
} from "@/api/accountApi"
import type { Course } from "@/api/courseApi"
import { getCourseList } from "@/api/courseApi"
import {
    getAnswerModeDescription,
    getAnswerModeLabel,
    getPlatformName,
    getVideoModeDescription,
    getVideoModeLabel,
    isConcurrentVideoMode,
    PLATFORMS,
} from "@/utils/platformUtils"
import { toast } from "@/components/ui/use-toast"

type AccountDetailProps = {
    account: Account
    onBack: () => void
    onChanged: () => void
}

const isRunningState = (state?: string): boolean => {
    return state === "starting" || state === "running" || state === "stopping"
}

const STATUS_LABELS: Record<string, string> = {
    idle: "未进行",
    starting: "启动中",
    running: "进行中",
    stopping: "停止中",
    succeeded: "已完成",
    failed: "失败",
    canceled: "已停止",
}

export function AccountDetail({ account, onBack, onChanged }: AccountDetailProps) {
    const [config, setConfig] = useState<AccountConfig>(() => normalizeConfig(account.config))
    const [courseList, setCourseList] = useState<Course[]>([])
    const [selectedCourses, setSelectedCourses] = useState<string[]>([])
    const [isLoadingCourses, setIsLoadingCourses] = useState(false)
    const [status, setStatus] = useState<AccountStatus>(
        account.status || {
            uid: account.uid,
            state: account.isRunning ? "running" : "idle",
        },
    )
    const [isRunning, setIsRunning] = useState(isRunningState(account.status?.state) || account.isRunning)
    const [lastError, setLastError] = useState(account.status?.lastError || account.lastError || "")
    const [logs, setLogs] = useState<string[]>([])
    const [hasStarted, setHasStarted] = useState(
        account.isRunning || (account.status?.state ? account.status.state !== "idle" : false),
    )
    const [isSaving, setIsSaving] = useState(false)
    const [isStarting, setIsStarting] = useState(false)
    const [isStopping, setIsStopping] = useState(false)
    const [isChecking, setIsChecking] = useState(false)
    const [showAccount, setShowAccount] = useState(true)
    const [showPassword, setShowPassword] = useState(false)
    const [loginChallenge, setLoginChallenge] = useState<LoginCodeChallenge | null>(null)
    const [loginCodeInput, setLoginCodeInput] = useState("")
    const [isSendingLoginCode, setIsSendingLoginCode] = useState(false)
    const [isSubmittingLoginCode, setIsSubmittingLoginCode] = useState(false)
    const logContainerRef = useRef<HTMLDivElement | null>(null)
    const stickToBottomRef = useRef(true)
    const didLoadCourses = useRef(false)
    const platform = useMemo(
        () => PLATFORMS.find((item) => item.code === config.accountType),
        [config.accountType],
    )

    useEffect(() => {
        setConfig(normalizeConfig(account.config))
        // 只在切换账号时重置表单，避免轮询刷新覆盖未保存的修改
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [account.uid])

    useEffect(() => {
        if (account.status) {
            setStatus((prev) =>
                prev.state === account.status!.state &&
                prev.taskId === account.status!.taskId &&
                prev.lastError === account.status!.lastError
                    ? prev
                    : account.status!,
            )
            setIsRunning((prev) => (prev === isRunningState(account.status!.state) ? prev : isRunningState(account.status!.state)))
            setLastError((prev) =>
                (prev || "") === (account.status!.lastError || "") ? prev : account.status!.lastError || "",
            )
            if (account.status.state !== "idle") {
                setHasStarted(true)
            }
            return
        }
        setIsRunning(account.isRunning)
    }, [account.uid, account.isRunning, account.status])

    // 任务状态以 /accountStatus 为准，避免前端靠布尔值猜测
    useEffect(() => {
        let active = true
        const poll = async () => {
            const response = await getAccountStatus(account.uid)
            if (!active || response.code !== 200 || !response.data?.status) {
                return
            }
            const next = response.data.status
            setStatus((prev) =>
                prev.state === next.state &&
                prev.taskId === next.taskId &&
                prev.lastError === next.lastError
                    ? prev
                    : next,
            )
            setIsRunning((prev) => (prev === isRunningState(next.state) ? prev : isRunningState(next.state)))
            setLastError((prev) => ((prev || "") === (next.lastError || "") ? prev : next.lastError || ""))
            if (next.state !== "idle") {
                setHasStarted(true)
            }
        }
        poll()
        const timer = setInterval(poll, 2000)
        return () => {
            active = false
            clearInterval(timer)
        }
    }, [account.uid])

    // 选课结果按账号持久化：返回列表再进来不会丢，也不会退回“全部课程”
    useEffect(() => {
        let restored: string[] = []
        try {
            const raw = window.localStorage.getItem(`yatori:selectedCourses:${account.uid}`)
            if (raw) {
                const parsed = JSON.parse(raw)
                if (Array.isArray(parsed)) {
                    restored = parsed.filter((item): item is string => typeof item === "string")
                }
            }
        } catch {
            restored = []
        }
        setSelectedCourses(restored)
    }, [account.uid])

    const updateSelection = useCallback(
        (next: string[]) => {
            setSelectedCourses(next)
            try {
                window.localStorage.setItem(`yatori:selectedCourses:${account.uid}`, JSON.stringify(next))
            } catch {
                // localStorage 不可用时只保留内存态
            }
        },
        [account.uid],
    )

    // 登录验证码轮询：学习通 APP 登录触发短信/图形验证码时弹窗输入
    useEffect(() => {
        let active = true
        const poll = async () => {
            const response = await getAccountLoginCode(account.uid)
            if (!active) {
                return
            }
            if (response.code === 200 && response.data?.pending && response.data.challenge) {
                setLoginChallenge(response.data.challenge)
            } else {
                setLoginChallenge(null)
            }
        }
        poll()
        const timer = setInterval(poll, 3000)
        return () => {
            active = false
            clearInterval(timer)
        }
    }, [account.uid])

    const loadCourses = useCallback(async (silent = false) => {
        if (!silent) {
            setIsLoadingCourses(true)
        }
        try {
            const response = await getCourseList(account.uid)
            if (response.code === 200 && response.data) {
                setCourseList(response.data.courseList || [])
            } else if (!silent) {
                toast({
                    title: "课程加载失败",
                    description: response.message,
                    variant: "destructive",
                })
            }
        } finally {
            if (!silent) {
                setIsLoadingCourses(false)
            }
        }
    }, [account.uid])

    useEffect(() => {
        // 避免 React StrictMode/重复渲染导致首次打开同时发起两次登录和课程请求
        if (didLoadCourses.current) {
            return
        }
        didLoadCourses.current = true
        loadCourses()
    }, [loadCourses])

    // 运行中定时刷新课程进度
    useEffect(() => {
        if (!isRunning) {
            return
        }
        const timer = setInterval(() => {
            loadCourses(true)
        }, 5000)
        return () => clearInterval(timer)
    }, [isRunning, loadCourses])

    // 账号日志轮询
    useEffect(() => {
        if (!hasStarted && !isRunning) {
            return
        }
        let active = true
        const poll = async () => {
            const response = await getAccountLog(account.uid, 300)
            if (!active || response.code !== 200 || !response.data) {
                return
            }
            const next = response.data.lines || []
            setLogs((prev) => {
                if (
                    prev.length === next.length &&
                    (next.length === 0 || prev[next.length - 1] === next[next.length - 1])
                ) {
                    return prev
                }
                return next
            })
        }
        poll()
        const timer = setInterval(poll, 2000)
        return () => {
            active = false
            clearInterval(timer)
        }
    }, [account.uid, hasStarted, isRunning])

    // 只滚动日志容器自身；用户往回翻日志时不再强行拉到底部，更不会影响整页滚动位置
    useEffect(() => {
        const container = logContainerRef.current
        if (!container || !stickToBottomRef.current) {
            return
        }
        container.scrollTop = container.scrollHeight
    }, [logs])

    // 任务结束后刷新一次课程列表，让本地记录的节点进度立刻反映到进度条上
    const prevTaskStateRef = useRef(status.state)
    useEffect(() => {
        const prev = prevTaskStateRef.current
        prevTaskStateRef.current = status.state
        if (prev === status.state) {
            return
        }
        if (status.state === "succeeded" || status.state === "failed" || status.state === "canceled") {
            const timer = setTimeout(() => loadCourses(true), 800)
            return () => clearTimeout(timer)
        }
    }, [status.state, loadCourses])

    const updateConfig = (patch: Partial<AccountConfig>) => {
        setConfig((prev) => ({ ...prev, ...patch }))
    }

    const updateCoursesCustom = (patch: Partial<AccountConfig["coursesCustom"]>) => {
        setConfig((prev) => ({
            ...prev,
            coursesCustom: { ...prev.coursesCustom, ...patch },
        }))
    }

    const handleSave = async () => {
        setIsSaving(true)
        try {
            // 课程选择改为仅通过“开始任务”时的勾选生效，清理旧的 include/exclude 配置
            const payload: AccountConfig = {
                ...config,
                coursesCustom: {
                    ...config.coursesCustom,
                    includeCourses: [],
                    excludeCourses: [],
                },
            }
            const response = await updateAccount(account.uid, payload)
            if (response.code === 200 && response.data) {
                setConfig(normalizeConfig(response.data.config))
                toast({ title: "保存成功", description: "账号配置已更新" })
                onChanged()
            } else {
                toast({
                    title: "保存失败",
                    description: response.message,
                    variant: "destructive",
                })
            }
        } finally {
            setIsSaving(false)
        }
    }

    const handleCheckLogin = async () => {
        setIsChecking(true)
        try {
            const response = await checkAccountLogin(account.uid)
            toast({
                title: response.code === 200 ? "登录正常" : "登录失败",
                description: response.code === 200 ? `平台：${getPlatformName(account.accountType)}` : response.message,
                variant: response.code === 200 ? "default" : "destructive",
            })
        } finally {
            setIsChecking(false)
        }
    }

    const handleStart = async () => {
        setIsStarting(true)
        try {
            const selectedNames = selectedCourses
                .map((courseId) => courseList.find((course) => course.courseId === courseId)?.courseName)
                .filter((name): name is string => Boolean(name))
            const response = await startBrushForUid(account.uid, selectedNames)
            if (response.code === 200) {
                setLastError("")
                setHasStarted(true)
                setStatus((prev) => ({ ...prev, uid: account.uid, state: "starting", lastError: "" }))
                setIsRunning(true)
                toast({
                    title: "任务已启动",
                    description:
                        selectedNames.length > 0
                            ? `本次只执行 ${selectedNames.length} 门课程`
                            : "将执行账号配置内的全部课程",
                })
                setTimeout(() => loadCourses(true), 1000)
            } else {
                setLastError(response.message)
                toast({
                    title: "启动失败",
                    description: response.message,
                    variant: "destructive",
                })
            }
        } finally {
            setIsStarting(false)
        }
    }

    const handleStop = async () => {
        setIsStopping(true)
        try {
            const response = await stopBrushForUid(account.uid)
            if (response.code === 200) {
                setStatus((prev) => ({ ...prev, uid: account.uid, state: "stopping" }))
                setIsRunning(true)
                toast({ title: "已请求停止", description: "执行器会在当前节点结束后退出" })
            } else {
                setLastError(response.message)
                toast({
                    title: "停止失败",
                    description: response.message,
                    variant: "destructive",
                })
            }
        } finally {
            setIsStopping(false)
        }
    }

    const handleSendLoginCode = async () => {
        setIsSendingLoginCode(true)
        try {
            const response = await sendAccountLoginCodeSms(account.uid)
            toast({
                title: response.code === 200 ? "验证码已发送" : "发送失败",
                description: response.message,
                variant: response.code === 200 ? "default" : "destructive",
            })
        } finally {
            setIsSendingLoginCode(false)
        }
    }

    const handleSubmitLoginCode = async () => {
        const code = loginCodeInput.trim()
        if (!code) {
            return
        }
        setIsSubmittingLoginCode(true)
        try {
            const response = await submitAccountLoginCode(account.uid, code)
            if (response.code === 200) {
                toast({ title: "验证码已提交", description: "等待登录结果" })
                setLoginCodeInput("")
                setLoginChallenge(null)
            } else {
                toast({ title: "提交失败", description: response.message, variant: "destructive" })
            }
        } finally {
            setIsSubmittingLoginCode(false)
        }
    }

    const maskText = (text: string) => "*".repeat(Math.min(text.length, 12))
    const allSelected = courseList.length > 0 && selectedCourses.length === courseList.length
    // 未选课时优先展示「本次任务实际执行的课程」，避免一进详情就显示全部课程
    const taskCourseNames = status.courses && status.courses.length > 0 ? status.courses : []
    const progressCourses =
        selectedCourses.length > 0
            ? courseList.filter((course) => selectedCourses.includes(course.courseId))
            : taskCourseNames.length > 0
              ? courseList.filter((course) => taskCourseNames.includes(course.courseName))
              : courseList
    const displayName = config.remarkName?.trim() || config.account

    return (
        <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3, ease: "easeOut" }}
            className="space-y-5"
        >
            <div className="flex items-center justify-between gap-3">
                <Button
                    variant="ghost"
                    size="sm"
                    onClick={onBack}
                    className="-ml-2 gap-1.5 text-sm text-muted-foreground hover:text-foreground"
                >
                    <ArrowLeft className="h-4 w-4" />
                    返回账号列表
                </Button>
            </div>

            <Card className="overflow-hidden">
                <CardHeader>
                    <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4">
                        <div className="flex h-12 w-12 sm:h-16 sm:w-16 items-center justify-center rounded-full bg-primary/10 text-primary">
                            <User className="h-6 w-6 sm:h-8 sm:w-8" />
                        </div>
                        <div className="flex-1 w-full">
                            <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-3 mb-2">
                                <h2 className="text-xl sm:text-2xl font-semibold text-foreground flex items-center gap-1">
                                    <span>{showAccount ? displayName : maskText(displayName)}</span>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-6 w-6"
                                        onClick={() => setShowAccount(!showAccount)}
                                    >
                                        {showAccount ? <Eye className="h-3 w-3" /> : <EyeOff className="h-3 w-3" />}
                                    </Button>
                                </h2>
                                <Badge
                                    variant={
                                        status.state === "failed"
                                            ? "destructive"
                                            : isRunning
                                              ? "default"
                                              : "secondary"
                                    }
                                >
                                    {STATUS_LABELS[status.state] || status.state}
                                </Badge>
                            </div>
                            <div className="space-y-2">
                                <div className="flex items-center gap-2 text-xs sm:text-sm text-muted-foreground">
                                    <Leaf className="h-3 w-3 sm:h-4 sm:w-4" />
                                    <span>{getPlatformName(config.accountType)}</span>
                                    {config.url && <span className="truncate">· {config.url}</span>}
                                </div>
                                <div className="flex max-w-xs items-center gap-2 text-xs sm:text-sm">
                                    <div className="flex min-w-0 items-center gap-2 text-muted-foreground">
                                        <KeyRound className="h-3 w-3 sm:h-4 sm:w-4" />
                                        <span className="truncate font-mono">
                                            {showPassword ? config.password : maskText(config.password)}
                                        </span>
                                    </div>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-6 w-6 shrink-0"
                                        onClick={() => setShowPassword(!showPassword)}
                                    >
                                        {showPassword ? <Eye className="h-3 w-3" /> : <EyeOff className="h-3 w-3" />}
                                    </Button>
                                </div>
                                <div className="flex items-center gap-2 text-xs sm:text-sm text-muted-foreground">
                                    <BookOpen className="h-3 w-3 sm:h-4 sm:w-4" />
                                    <span>{courseList.length} 门课程，已选 {selectedCourses.length} 门</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="flex flex-wrap gap-3">
                        <Button
                            variant="default"
                            className="gap-2"
                            onClick={handleSave}
                            disabled={isRunning || isSaving}
                        >
                            <Check className="h-4 w-4" />
                            {isSaving ? "保存中..." : "保存配置"}
                        </Button>
                        <Button
                            variant="outline"
                            className="gap-2"
                            onClick={handleCheckLogin}
                            disabled={isChecking || isRunning}
                        >
                            {isChecking ? <Loader2 className="h-4 w-4 animate-spin" /> : <ShieldCheck className="h-4 w-4" />}
                            登录检测
                        </Button>
                        {isRunning ? (
                            <Button
                                variant="destructive"
                                className="gap-2"
                                onClick={handleStop}
                                disabled={isStopping || status.state === "stopping"}
                            >
                                {isStopping || status.state === "stopping" ? (
                                    <Loader2 className="h-4 w-4 animate-spin" />
                                ) : (
                                    <StopCircle className="h-4 w-4" />
                                )}
                                {status.state === "starting" ? "启动中..." : status.state === "stopping" ? "停止中..." : "停止任务"}
                            </Button>
                        ) : (
                            <Button className="gap-2" onClick={handleStart} disabled={isStarting}>
                                {isStarting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
                                {isStarting ? "启动中..." : "开始任务"}
                            </Button>
                        )}
                    </div>
                    {isRunning && (
                        <p className="text-xs text-muted-foreground">
                            任务运行中，配置已锁定；如需修改请先停止任务。
                        </p>
                    )}
                    {lastError && (
                        <div className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive whitespace-pre-wrap break-words">
                            {lastError}
                        </div>
                    )}
                </CardContent>
            </Card>

            <Card className="relative overflow-hidden">
                <CardHeader className="pb-4">
                    <div className="flex items-start gap-3">
                        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
                            <SlidersHorizontal className="h-4 w-4" />
                        </div>
                        <div>
                            <h3 className="text-base font-semibold text-foreground sm:text-lg">功能配置</h3>
                            <p className="mt-1 text-xs text-muted-foreground">
                                课程选择、视频模式、答题模式与通知设置
                            </p>
                        </div>
                    </div>
                </CardHeader>
                <CardContent className={`space-y-6 ${isRunning ? "opacity-60 pointer-events-none" : ""}`}>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label>备注名</Label>
                            <Input
                                value={config.remarkName || ""}
                                placeholder="用于区分多个账号"
                                onChange={(e) => updateConfig({ remarkName: e.target.value })}
                            />
                        </div>
                        <div className="flex items-center justify-between rounded-lg border p-3">
                            <div>
                                <Label className="text-sm">IP 代理</Label>
                                <p className="text-xs text-muted-foreground mt-1">从 ip.txt 随机选择代理</p>
                            </div>
                            <Switch
                                checked={config.isProxy === 1}
                                onCheckedChange={(checked) => updateConfig({ isProxy: checked ? 1 : 0 })}
                            />
                        </div>
                    </div>

                    <div className="space-y-2">
                        <Label>密码 / Cookie / Token</Label>
                        <div className="relative">
                            <Input
                                type={showPassword ? "text" : "password"}
                                value={config.password}
                                onChange={(e) => updateConfig({ password: e.target.value })}
                                className="pr-10"
                            />
                            <Button
                                type="button"
                                variant="ghost"
                                size="icon"
                                className="absolute right-0 top-0 h-full w-10 hover:bg-transparent"
                                onClick={() => setShowPassword(!showPassword)}
                            >
                                {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                            </Button>
                        </div>
                    </div>

                    {config.accountType === "XUEXITONG" && (
                        <div className="space-y-2">
                            <Label>登录方式</Label>
                            <Select
                                value={config.loginMode === "app" ? "app" : "web"}
                                onValueChange={(value) => updateConfig({ loginMode: value })}
                            >
                                <SelectTrigger className="w-full">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="web">网页登录</SelectItem>
                                    <SelectItem value="app">APP 登录（自动获取人脸签名）</SelectItem>
                                </SelectContent>
                            </Select>
                            <p className="text-xs text-muted-foreground">
                                APP 登录会走学习通 App 端协议，自动拿到人脸签名 cid/sc，不再需要手动配置；
                                触发短信/图形验证码时会在页面弹窗提示输入。该模式下多课程/多任务点会退化为串行。
                            </p>
                        </div>
                    )}

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label>视频模式</Label>
                            <Select
                                value={String(config.coursesCustom.videoModel)}
                                onValueChange={(value) => updateCoursesCustom({ videoModel: Number(value) })}
                            >
                                <SelectTrigger className="w-full">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    {(platform?.videoModes || [0, 1, 2]).map((mode) => (
                                        <SelectItem key={mode} value={String(mode)}>
                                            {getVideoModeLabel(config.accountType, mode)}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            {getVideoModeDescription(config.accountType, config.coursesCustom.videoModel) && (
                                <p className="text-xs text-muted-foreground">
                                    {getVideoModeDescription(config.accountType, config.coursesCustom.videoModel)}
                                </p>
                            )}
                            {isConcurrentVideoMode(config.accountType, config.coursesCustom.videoModel) && (
                                <div className="space-y-1 pt-1">
                                    <Label className="text-xs">同时任务点数（1-5）</Label>
                                    <Input
                                        type="number"
                                        min={1}
                                        max={5}
                                        value={config.coursesCustom.cxNode ?? 3}
                                        onChange={(e) =>
                                            updateCoursesCustom({
                                                cxNode: Math.max(1, Math.min(5, Number(e.target.value) || 1)),
                                            })
                                        }
                                    />
                                </div>
                            )}
                        </div>
                        <div className="space-y-2">
                            <Label>答题模式</Label>
                            <Select
                                value={String(config.coursesCustom.autoExam)}
                                onValueChange={(value) => updateCoursesCustom({ autoExam: Number(value) })}
                            >
                                <SelectTrigger className="w-full">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    {(platform?.answerModes || [0]).map((mode) => (
                                        <SelectItem key={mode} value={String(mode)}>
                                            {getAnswerModeLabel(config.accountType, mode)}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            {config.coursesCustom.autoExam !== 0 && (
                                <p className="text-xs text-muted-foreground">
                                    {getAnswerModeDescription(config.accountType, config.coursesCustom.autoExam)}
                                </p>
                            )}
                        </div>
                    </div>

                    {config.accountType === "XUEXITONG" && (
                        <div className="space-y-2">
                            <Label>学习通任务开关</Label>
                            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                                <div className="flex items-center justify-between rounded-lg border p-3">
                                    <div>
                                        <Label className="text-sm">章节测验</Label>
                                        <p className="text-xs text-muted-foreground mt-1">课程树里的章测任务点</p>
                                    </div>
                                    <Switch
                                        checked={(config.coursesCustom.cxChapterTestSw ?? 1) !== 0}
                                        onCheckedChange={(checked) =>
                                            updateCoursesCustom({ cxChapterTestSw: checked ? 1 : 0 })
                                        }
                                    />
                                </div>
                                <div className="flex items-center justify-between rounded-lg border p-3">
                                    <div>
                                        <Label className="text-sm">作业</Label>
                                        <p className="text-xs text-muted-foreground mt-1">作业列表</p>
                                    </div>
                                    <Switch
                                        checked={(config.coursesCustom.cxWorkSw ?? 1) !== 0}
                                        onCheckedChange={(checked) => updateCoursesCustom({ cxWorkSw: checked ? 1 : 0 })}
                                    />
                                </div>
                                <div className="flex items-center justify-between rounded-lg border p-3">
                                    <div>
                                        <Label className="text-sm">考试</Label>
                                        <p className="text-xs text-muted-foreground mt-1">考试列表</p>
                                    </div>
                                    <Switch
                                        checked={(config.coursesCustom.cxExamSw ?? 1) !== 0}
                                        onCheckedChange={(checked) => updateCoursesCustom({ cxExamSw: checked ? 1 : 0 })}
                                    />
                                </div>
                            </div>
                            <p className="text-xs text-muted-foreground">
                                章节测验/作业/考试都按上方「答题模式」作答，未开启答题模式时只做视频。
                            </p>
                        </div>
                    )}

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label>交卷方式</Label>
                            <Select
                                value={String(config.coursesCustom.examAutoSubmit)}
                                onValueChange={(value) => updateCoursesCustom({ examAutoSubmit: Number(value) })}
                            >
                                <SelectTrigger className="w-full">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="0">只保存不交卷</SelectItem>
                                    <SelectItem value="1">答完自动交卷</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="space-y-2">
                            <Label>打乱学习顺序</Label>
                            <Select
                                value={String(config.coursesCustom.shuffleSw)}
                                onValueChange={(value) => updateCoursesCustom({ shuffleSw: Number(value) })}
                            >
                                <SelectTrigger className="w-full">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="0">按平台顺序</SelectItem>
                                    <SelectItem value="1">随机顺序</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>

                    {/* 课程选择移入功能配置 */}
                    <div className="space-y-2">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                                <ListChecks className="h-4 w-4 text-muted-foreground" />
                                <Label>选择课程（不选表示全部）</Label>
                            </div>
                            <div className="flex gap-2">
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="sm"
                                    onClick={() => updateSelection(courseList.map((course) => course.courseId))}
                                    disabled={courseList.length === 0}
                                >
                                    全选
                                </Button>
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="sm"
                                    onClick={() => updateSelection([])}
                                    disabled={selectedCourses.length === 0}
                                >
                                    清空
                                </Button>
                            </div>
                        </div>
                        <div className="rounded-lg border max-h-64 overflow-y-auto overscroll-contain">
                            {isLoadingCourses ? (
                                <div className="flex items-center justify-center py-6 text-sm text-muted-foreground">
                                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                                    加载课程中...
                                </div>
                            ) : courseList.length === 0 ? (
                                <div className="py-6 text-center text-sm text-muted-foreground">
                                    暂无课程，请先刷新
                                </div>
                            ) : (
                                <div className="divide-y">
                                    {courseList.map((course) => (
                                        <label
                                            key={course.courseId}
                                            className="flex items-center gap-3 px-3 py-2 cursor-pointer hover:bg-muted/50"
                                        >
                                            <Checkbox
                                                checked={selectedCourses.includes(course.courseId)}
                                                onCheckedChange={(checked) => {
                                                    if (checked) {
                                                        updateSelection([...selectedCourses, course.courseId])
                                                    } else {
                                                        updateSelection(
                                                            selectedCourses.filter((id) => id !== course.courseId),
                                                        )
                                                    }
                                                }}
                                            />
                                            <span className="flex-1 text-sm truncate">{course.courseName}</span>
                                            <span className="text-xs text-muted-foreground">
                                                {course.progress.toFixed(1)}%
                                            </span>
                                        </label>
                                    ))}
                                </div>
                            )}
                        </div>
                        <div className="flex justify-end">
                            <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                className="gap-2"
                                onClick={() => loadCourses()}
                                disabled={isLoadingCourses}
                            >
                                <RefreshCcw className={`h-4 w-4 ${isLoadingCourses ? "animate-spin" : ""}`} />
                                刷新课程
                            </Button>
                        </div>
                    </div>

                    <div className="space-y-3">
                        <div className="flex items-center justify-between">
                            <Label>通知邮箱</Label>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() =>
                                    updateConfig({ informEmails: [...(config.informEmails || []), ""] })
                                }
                            >
                                新增邮箱
                            </Button>
                        </div>
                        <div className="space-y-2">
                            {(config.informEmails || []).map((email, index) => (
                                <div key={index} className="flex items-center gap-2">
                                    <Input
                                        type="email"
                                        placeholder="输入邮箱地址"
                                        value={email}
                                        onChange={(e) => {
                                            const next = [...(config.informEmails || [])]
                                            next[index] = e.target.value
                                            updateConfig({ informEmails: next })
                                        }}
                                    />
                                    <Button
                                        variant="outline"
                                        size="icon"
                                        className="h-9 w-9"
                                        onClick={() =>
                                            updateConfig({
                                                informEmails: (config.informEmails || []).filter((_, i) => i !== index),
                                            })
                                        }
                                    >
                                        <Trash2 className="h-4 w-4 text-muted-foreground" />
                                    </Button>
                                </div>
                            ))}
                            {(config.informEmails || []).length === 0 && (
                                <p className="text-xs text-muted-foreground">未配置通知邮箱</p>
                            )}
                        </div>
                    </div>
                </CardContent>
                {isRunning && (
                    <div className="absolute inset-0 z-[1] flex flex-col items-center justify-center gap-2 rounded-xl bg-background/40 backdrop-blur-[1px] pointer-events-none">
                        <Lock className="h-8 w-8 text-muted-foreground" />
                        <span className="text-base font-medium text-foreground/80">请先停止任务</span>
                    </div>
                )}
            </Card>

            {/* 任务进度 + 运行日志 */}
            {(hasStarted || isRunning) && (
                <div className="grid grid-cols-1 gap-5 lg:grid-cols-2">
                    <Card className="overflow-hidden">
                        <CardHeader className="pb-4">
                            <div className="flex items-center justify-between gap-2">
                                <div className="flex items-center gap-2">
                                    <BarChart3 className="h-4 w-4 text-primary" />
                                    <h3 className="text-base font-semibold text-foreground">任务进度</h3>
                                </div>
                                {isRunning && (
                                    <span className="flex items-center gap-1.5 text-xs font-medium text-primary">
                                        <span className="relative flex h-2 w-2">
                                            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-primary opacity-75" />
                                            <span className="relative inline-flex h-2 w-2 rounded-full bg-primary" />
                                        </span>
                                        进行中
                                    </span>
                                )}
                            </div>
                            <p className="text-xs text-muted-foreground">
                                {selectedCourses.length > 0
                                    ? `当前展示已选 ${selectedCourses.length} 门课程`
                                    : taskCourseNames.length > 0
                                      ? `当前展示本次任务的 ${taskCourseNames.length} 门课程`
                                      : "未选择课程时展示全部课程"}
                            </p>
                        </CardHeader>
                        <CardContent className="space-y-4 max-h-[420px] overflow-y-auto overscroll-contain">
                            {progressCourses.length === 0 ? (
                                <div className="py-8 text-center text-sm text-muted-foreground">暂无课程进度</div>
                            ) : (
                                progressCourses.map((course) => (
                                    <div key={course.courseId} className="space-y-1.5">
                                        <div className="flex items-center justify-between gap-3 text-sm">
                                            <span className="truncate">{course.courseName}</span>
                                            <span className="text-primary font-medium tabular-nums">
                                                {course.progress.toFixed(1)}%
                                            </span>
                                        </div>
                                        <div className="relative">
                                            <Progress
                                                value={course.progress}
                                                className={`h-2 transition-all duration-500 ${
                                                    isRunning ? "animate-pulse" : ""
                                                }`}
                                            />
                                            {isRunning && (
                                                <span className="pointer-events-none absolute inset-0 overflow-hidden rounded-full">
                                                    <span className="absolute inset-y-0 w-1/3 animate-[progress-shimmer_1.8s_ease-in-out_infinite] rounded-full bg-white/50 blur-[2px] dark:bg-white/30" />
                                                </span>
                                            )}
                                        </div>
                                    </div>
                                ))
                            )}
                        </CardContent>
                    </Card>

                    <Card className="flex flex-col overflow-hidden">
                        <CardHeader className="pb-4">
                            <div className="flex items-center gap-2">
                                <Terminal className="h-4 w-4 text-primary" />
                                <h3 className="text-base font-semibold text-foreground">运行日志</h3>
                            </div>
                            <p className="text-xs text-muted-foreground">最近 300 行，每 2 秒刷新</p>
                        </CardHeader>
                        <CardContent className="flex-1">
                            <div
                                ref={logContainerRef}
                                onScroll={(event) => {
                                    const el = event.currentTarget
                                    stickToBottomRef.current =
                                        el.scrollHeight - el.scrollTop - el.clientHeight < 48
                                }}
                                className="h-[360px] overflow-y-auto overscroll-contain rounded-lg bg-muted/40 p-3 font-mono text-xs leading-5"
                            >
                                {logs.length === 0 ? (
                                    <div className="text-muted-foreground">暂无日志</div>
                                ) : (
                                    logs.map((line, index) => (
                                        <div key={index} className="whitespace-pre-wrap break-all">
                                            {line}
                                        </div>
                                    ))
                                )}
                            </div>
                        </CardContent>
                    </Card>
                </div>
            )}

            <Dialog
                open={!!loginChallenge}
                onOpenChange={(open) => {
                    if (!open) {
                        setLoginChallenge(null)
                    }
                }}
            >
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>需要验证码</DialogTitle>
                        <DialogDescription>{loginChallenge?.prompt || "学习通登录需要验证码"}</DialogDescription>
                    </DialogHeader>
                    {loginChallenge?.phone && (
                        <p className="text-sm text-muted-foreground">手机号：{loginChallenge.phone}</p>
                    )}
                    {loginChallenge?.image && (
                        <img
                            src={`data:image/png;base64,${loginChallenge.image}`}
                            alt="验证码"
                            className="mx-auto rounded-md border"
                        />
                    )}
                    <Input
                        value={loginCodeInput}
                        onChange={(e) => setLoginCodeInput(e.target.value)}
                        placeholder="请输入验证码"
                        onKeyDown={(e) => {
                            if (e.key === "Enter") {
                                void handleSubmitLoginCode()
                            }
                        }}
                    />
                    <div className="flex justify-end gap-2">
                        {loginChallenge?.canSendSms && (
                            <Button
                                type="button"
                                variant="outline"
                                onClick={handleSendLoginCode}
                                disabled={isSendingLoginCode}
                            >
                                {isSendingLoginCode ? "发送中..." : "发送验证码"}
                            </Button>
                        )}
                        <Button
                            type="button"
                            onClick={handleSubmitLoginCode}
                            disabled={isSubmittingLoginCode || !loginCodeInput.trim()}
                        >
                            {isSubmittingLoginCode ? "提交中..." : "确定"}
                        </Button>
                    </div>
                </DialogContent>
            </Dialog>
        </motion.div>
    )
}

