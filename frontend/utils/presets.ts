export interface CoursesCustomPreset {
    cxChapterTestSw?: number
    cxWorkSw?: number
    cxExamSw?: number
    autoExam?: number
    examAutoSubmit?: number
    cxNode?: number
    videoModel?: number
}

export interface PresetOption {
    label: string
    description: string
    coursesCustom: CoursesCustomPreset
}

export const COURSE_PRESETS: Record<string, PresetOption> = {
    "video-only": {
        label: "仅刷视频",
        description: "所有课程仅刷视频，不做题",
        coursesCustom: {
            cxChapterTestSw: 0,
            cxWorkSw: 0,
            cxExamSw: 0,
            autoExam: 0,
            examAutoSubmit: 0,
            videoModel: 1,
            cxNode: 3,
        },
    },
    "video-and-test": {
        label: "刷视频+章测自动提交",
        description: "刷视频同时自动做章测并提交",
        coursesCustom: {
            cxChapterTestSw: 1,
            cxWorkSw: 0,
            cxExamSw: 0,
            autoExam: 1,
            examAutoSubmit: 1,
            videoModel: 1,
            cxNode: 3,
        },
    },
}

// 默认（不设置任何特殊选项，程序用默认值）
export const DEFAULT_PRESET_KEY = "default"
