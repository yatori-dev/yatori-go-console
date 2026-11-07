import axios from 'axios'

// 默认使用相对路径：console 的 Gin 服务在 /api 下提供接口。
// 本地单独跑 Next 开发服务器时，可通过 NEXT_PUBLIC_API_BASE 指向 http://localhost:8080/api。
const apiBase = process.env.NEXT_PUBLIC_API_BASE || '/api'

const apiClient = axios.create({
    baseURL: apiBase,
    timeout: 240000,
    headers: {
        'Content-Type': 'application/json',
    },
})

apiClient.interceptors.response.use(
    (response) => response.data,
    (error) => {
        console.error('API请求错误:', error)
        return Promise.reject(error)
    },
)

export default apiClient
