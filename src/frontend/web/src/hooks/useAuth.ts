import { useEffect, useCallback, useRef, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useSelector, useDispatch } from 'react-redux';
import { RootState } from '@/store';
import { setAuth, clearAuth } from '@/store/slices/authSlice';
import { validateTokenAndGetUser, login as loginApi } from '@/api/auth';
import type { LoginRequest, User } from '@/store/types';

export const useAuth = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const dispatch = useDispatch();
  const { token, user } = useSelector((state: RootState) => state.auth);
  const validatingRef = useRef(false);
  const [isInitialCheckDone, setIsInitialCheckDone] = useState(false);

  const validateAuth = useCallback(async () => {
    // 如果正在验证中，则跳过
    if (validatingRef.current) {
      return;
    }

    // 如果在登录页面，直接标记为已完成
    if (location.pathname.includes('/login')) {
      setIsInitialCheckDone(true);
      return;
    }

    try {
      // 如果没有 token，重定向到登录页
      if (!token) {
        setIsInitialCheckDone(true);
        if (!location.pathname.includes('/login')) {
          navigate(`/login?redirect=${encodeURIComponent(location.pathname)}`);
        }
        return;
      }

      // 如果有 token 但没有 user，则进行验证
      if (!user) {
        validatingRef.current = true;
        const userData = await validateTokenAndGetUser(token);
        dispatch(setAuth({ token, user: userData.user }));
      }
      
      setIsInitialCheckDone(true);
    } catch (error) {
      // 验证失败，清除认证信息
      localStorage.removeItem('token');
      dispatch(clearAuth());
      if (!location.pathname.includes('/login')) {
        navigate(`/login?redirect=${encodeURIComponent(location.pathname)}`);
      }
    } finally {
      validatingRef.current = false;
      setIsInitialCheckDone(true);
    }
  }, [token, user, dispatch, navigate, location]);

  // 添加登录方法
  const login = async (data: LoginRequest) => {
    try {
      const response = await loginApi(data);
      // 保存 token 到 localStorage
      localStorage.setItem('token', response.token);
      // 设置认证信息
      dispatch(setAuth({ token: response.token, user: response.user }));
      // 获取重定向地址
      const params = new URLSearchParams(location.search);
      const redirect = params.get('redirect') || '/clips';
      setIsInitialCheckDone(true);
      validatingRef.current = false;
      // 导航到目标页面
      navigate(redirect);
    } catch (error) {
      throw error;
    }
  };

  const logout = () => {
    localStorage.removeItem('token');
    dispatch(clearAuth());
    setIsInitialCheckDone(false);
    validatingRef.current = false;
    navigate('/login');
  };

  // 在组件挂载时进行验证
  useEffect(() => {
    const storedToken = localStorage.getItem('token');
    if (storedToken && !token) {
      dispatch(setAuth({ token: storedToken, user: undefined }));
    }
    validateAuth();
  }, [validateAuth, dispatch, token]);

  return {
    user,
    token,
    login,
    logout,
    isInitialCheckDone
  };
};

export default useAuth; 