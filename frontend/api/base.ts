import axios from 'axios'

// 创建Axios实例
const apiClient = axios.create({
  baseURL: 'http://localhost:8080/api', // 根据你的实际后端API地址配置
  // baseURL: './api', // 根据你的实际后端API地址配置
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
apiClient.interceptors.request.use(
  (config) => {
    // 可以在这里添加认证信息，如token
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
