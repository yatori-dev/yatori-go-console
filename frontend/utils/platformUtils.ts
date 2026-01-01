export const getPlatformName = (platformCode: string): string => {
    const platformMap: Record<string, string> = {
        "XUEXITONG": "学习通",
        "YINGHUA": "英华学堂",
        "QSXT": "青书学堂"
    }
    return platformMap[platformCode] || platformCode
}
