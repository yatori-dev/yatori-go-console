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
    // 若用户设置了管理密码，则通过请求头携带，用于后端 /api/v1 鉴权
    if (typeof window !== 'undefined') {
      const adminPass = window.localStorage.getItem('yatori-admin-pass')
      if (adminPass) {
        config.headers['X-Admin-Pass'] = adminPass
      }
    }
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
    if (error.response?.status === 401) {
      console.error('API 鉴权失败，请在侧边栏输入管理密码')
    }
    console.error('API请求错误:', error)
    return Promise.reject(error)
  }
)

export default apiClient
export const ADMIN_PASS_KEY = 'yatori-admin-pass'
