export type PlatformOption = {
    code: string
    name: string
    requiresUrl: boolean
    videoModes: number[]
    answerModes: number[]
    description?: string
}

// 视频模式编号沿用老版 yatori-go-console 的约定，各平台含义不同：
//
//	学习通：0 不刷 / 1 普通模式 / 2 多课程模式 / 3 多任务点模式
//	英华：  0 不刷 / 1 普通模式 / 2 暴力模式 / 3 去红模式
//	海旗：  0 不刷 / 1 普通模式 / 2 秒刷模式
//	学习公社：0 不刷 / 1 普通模式 / 2 暴力模式
//	重庆工程：0 不刷 / 1 常规模式 / 2 暴力模式（秒刷）
//	WeLearn：0 不刷 / 1 刷学时 / 2 刷完成度
//	智慧职教 / 码上研训 / 青书学堂：0 不刷 / 1 普通模式（刷学时）
export const PLATFORMS: PlatformOption[] = [
    {
        code: 'XUEXITONG',
        name: '超星学习通',
        requiresUrl: false,
        videoModes: [0, 1, 2, 3],
        answerModes: [0, 1, 2, 3],
    },
    {
        code: 'YINGHUA',
        name: '英华学堂',
        requiresUrl: true,
        videoModes: [0, 1, 2, 3],
        answerModes: [0, 1, 2],
    },
    { code: 'HQKJ', name: '海旗科技', requiresUrl: true, videoModes: [0, 1, 2], answerModes: [0, 1, 2] },
    { code: 'ICVE', name: '智慧职教', requiresUrl: false, videoModes: [0, 1], answerModes: [0] },
    { code: 'CQIE', name: '重庆工程学院', requiresUrl: false, videoModes: [0, 1, 2], answerModes: [0] },
    { code: 'ENAEA', name: '学习公社', requiresUrl: false, videoModes: [0, 1, 2], answerModes: [0] },
    { code: 'WELEARN', name: '随行课堂 WeLearn', requiresUrl: false, videoModes: [0, 1, 2], answerModes: [0] },
    { code: 'QSXT', name: '青书学堂', requiresUrl: false, videoModes: [0, 1], answerModes: [0, 1] },
    { code: 'KETANGX', name: '码上研训', requiresUrl: false, videoModes: [0, 1], answerModes: [0] },
]

const platformMap = new Map(PLATFORMS.map((item) => [item.code, item]))

export const getPlatformName = (platformCode: string): string => {
    return platformMap.get(platformCode)?.name || platformCode
}

export const getPlatform = (platformCode: string): PlatformOption | undefined => {
    return platformMap.get(platformCode)
}

export const platformRequiresUrl = (platformCode: string): boolean => {
    return getPlatform(platformCode)?.requiresUrl ?? false
}

const VIDEO_MODE_LABELS: Record<string, Record<number, string>> = {
    XUEXITONG: { 0: '不刷', 1: '普通模式', 2: '多课程模式', 3: '多任务点模式' },
    YINGHUA: { 0: '不刷', 1: '普通模式', 2: '暴力模式', 3: '去红模式' },
    HQKJ: { 0: '不刷', 1: '普通模式', 2: '秒刷模式' },
    ENAEA: { 0: '不刷', 1: '普通模式', 2: '暴力模式' },
    CQIE: { 0: '不刷', 1: '常规模式', 2: '暴力模式（秒刷）' },
    WELEARN: { 0: '不刷', 1: '刷学时', 2: '刷完成度' },
    ICVE: { 0: '不刷', 1: '刷学时' },
    KETANGX: { 0: '不刷', 1: '普通模式' },
    QSXT: { 0: '不刷', 1: '刷学时' },
}

export const getVideoModeLabel = (platformCode: string, mode: number): string => {
    const label = VIDEO_MODE_LABELS[platformCode]?.[mode]
    if (label) {
        return label
    }
    switch (mode) {
        case 0:
            return '不刷'
        case 1:
            return '普通模式'
        default:
            return `模式 ${mode}`
    }
}

// 该模式是否属于“多任务点并发”，需要展示“同时任务点数”
export const isConcurrentVideoMode = (platformCode: string, mode: number): boolean => {
    return platformCode === 'XUEXITONG' && mode === 3
}

export const getVideoModeDescription = (platformCode: string, mode: number): string => {    if (mode === 0) {
        return '只处理作业/考试，不学习视频任务点'
    }
    if (platformCode === 'XUEXITONG') {
        if (mode === 2) {
            return '最多 3 门课程同时推进，每门课使用一个独立会话'
        }
        if (mode === 3) {
            return '同一课程内多个任务点同时推进，并发数取「同时任务点数」（默认 3）'
        }
        if (mode === 1) {
            return '按章节顺序逐个任务点学习，节点之间随机等待'
        }
    }
    if (platformCode === 'YINGHUA') {
        if (mode === 2) {
            return '暴力模式：快速提交任务点，跑完后自动执行一次去红'
        }
        if (mode === 3) {
            return '去红模式：对已有进度但未完成的任务点重复提交当前进度，清理红标（不学习新节点）'
        }
    }
    if (platformCode === 'HQKJ' && mode === 2) {
        return '秒刷模式：每个任务点单次提交完成'
    }
    if (platformCode === 'CQIE' && mode === 2) {
        return '暴力模式（秒刷）：每个任务点单次提交完成'
    }
    if (platformCode === 'ENAEA' && mode === 2) {
        return '暴力模式：直接提交学习时长'
    }
    if (platformCode === 'WELEARN') {
        if (mode === 1) {
            return '按任务点提交学习时长'
        }
        if (mode === 2) {
            return '按任务点提交完成度'
        }
    }
    return ''
}

export const getAnswerModeLabel = (platformCode: string, mode: number): string => {
    switch (mode) {
        case 0:
            return '不答题'
        case 1:
            return 'AI 大模型答题'
        case 2:
            return platformCode === 'HQKJ' ? '平台内置标答' : '外置题库答题'
        case 3:
            return '学习通内置 AI'
        default:
            return String(mode)
    }
}

export const getAnswerModeDescription = (platformCode: string, mode: number): string => {
    if (platformCode === 'HQKJ') {
        if (mode === 1) {
            return '平台有标答时优先使用标答，缺失时才调用 AI'
        }
        if (mode === 2) {
            return '仅使用平台随题返回的内置标答，不调用 AI；没有标答的题目会跳过'
        }
    }
    if (platformCode === 'XUEXITONG' && mode === 3) {
        return 'v3 SDK 暂未提供学习通内置 AI 接口，实际回退到 AI 大模型答题'
    }
    if (mode === 1) {
        return '调用全局 AI 配置进行答题'
    }
    if (mode === 2) {
        return '通过外置题库服务获取答案'
    }
    return ''
}

export const getSubmitModeLabel = (mode: number): string => {
    return mode === 1 ? '答完自动交卷' : '只保存不交卷'
}
