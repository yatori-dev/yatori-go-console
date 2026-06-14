export const getPlatformName = (platformCode: string): string => {
    const platformMap: Record<string, string> = {
        "XUEXITONG": "学习通",
        "YINGHUA":   "英华学堂",
        "CANGHUI":   "仓辉实训",
        "ENAEA":     "学习公社",
        "CQIE":      "重庆工程学院",
        "KETANGX":   "码上研训",
        "ICVE":      "智慧职教",
        "QSXT":      "青书学堂",
        "WELEARN":   "WeLearn",
        "HQKJ":      "海旗科技",
    }
    return platformMap[platformCode] || platformCode
}

// 可用于下拉选择器的平台列表
export const PLATFORM_OPTIONS = [
    { value: "XUEXITONG", label: "学习通" },
    { value: "YINGHUA",   label: "英华学堂" },
    { value: "CANGHUI",   label: "仓辉实训" },
    { value: "ENAEA",     label: "学习公社" },
    { value: "CQIE",      label: "重庆工程学院" },
    { value: "KETANGX",   label: "码上研训" },
    { value: "ICVE",      label: "智慧职教" },
    { value: "QSXT",      label: "青书学堂" },
    { value: "WELEARN",   label: "WeLearn" },
    { value: "HQKJ",      label: "海旗科技" },
]
