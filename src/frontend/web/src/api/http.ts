import axios from 'axios';
import { message } from 'antd';
import { store } from '@/store';
import { clearAuth } from '@/store/slices/authSlice';

const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
http.interceptors.request.use(
  (config) => {
    // 优先从 localStorage 获取 token
    const token = localStorage.getItem('token') || store.getState().auth.token;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
http.interceptors.response.use(
  (response) => {
    return response.data;
  },
  async (error) => {
    if (error.response) {
      const { status, data } = error.response;
      
      // 处理401错误
      if (status === 401) {
        // 清除认证信息
        localStorage.removeItem('token');
        store.dispatch(clearAuth());
        
        // 获取当前路由
        const currentPath = window.location.pathname;
        
        // 如果不是登录页,则跳转到登录页并携带 redirect 参数
        if (!currentPath.includes('/login')) {
          window.location.href = `/login?redirect=${encodeURIComponent(currentPath)}`;
        }
        
        return Promise.reject(new Error('登录已过期，请重新登录'));
      }

      // 处理其他错误
      const errorMessage = data.message || '请求失败';
      message.error(errorMessage);
      return Promise.reject(new Error(errorMessage));
    }

    // 处理网络错误
    message.error('网络错误，请检查网络连接');
    return Promise.reject(error);
  }
);

export default http; 