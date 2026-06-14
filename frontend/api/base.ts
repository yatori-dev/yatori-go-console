import axios from 'axios'

// 创建Axios实例
const apiClient = axios.create({
  // 使用相对路径，便于挂域名、反向代理和同源部署
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
apiClient.interceptors.request.use(
  (config) => {
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => {
    // 统一处理响应
    return response.data
  },
  (error) => {
    // 统一处理错误
    console.error('API请求错误:', error)
    return Promise.reject(error)
  }
)

export default apiClient
