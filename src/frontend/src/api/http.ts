import axios from 'axios';
import { message } from 'antd';
import { store } from '@/store';
import { clearAuth } from '@/store/slices/authSlice';

const http = axios.create({
  baseURL: window.location.origin + '/api/v1/nlip',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
});

// 请求拦截器
http.interceptors.request.use(
  (config) => {
    // 从 Redux store 获取 token
    const token = store.getState().auth.token;
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
    // 如果是 blob 类型的响应，直接返回
    if (response.config.responseType === 'blob') {
      return response;
    }
    
    return response.data;
  },
  async (error) => {
    if (error.response) {
      const { status, data } = error.response;
      
      // 处理401错误
      if (status === 401) {
        // 清除认证信息
        store.dispatch(clearAuth());
        
        // 获取当前路由
        const currentPath = window.location.pathname;
        
        // 如果不是登录页且不是验证token的请求，则跳转到登录页
        if (!currentPath.includes('/login') && !error.config.url.includes('/auth/me')) {
          window.location.href = `/login?redirect=${encodeURIComponent(currentPath)}`;
        }
        
        return Promise.reject(new Error(data.message || '登录已过期，请重新登录'));
      }

      // 处理其他错误
      const errorMessage = data.message || '请求失败';
      // message.error(errorMessage);
      return Promise.reject(new Error(errorMessage));
    }

    // 处理网络错误
    message.error('网络错误，请检查网络连接');
    return Promise.reject(error);
  }
);

export default http; 