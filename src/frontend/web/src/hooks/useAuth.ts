import { useEffect, useCallback, useRef, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useSelector, useDispatch } from 'react-redux';
import { RootState, store } from '@/store';
import { setAuth, clearAuth } from '@/store/slices/authSlice';
import { validateTokenAndGetUser, login as loginApi } from '@/api/auth';
import type { LoginRequest } from '@/store/types';

export const useAuth = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const dispatch = useDispatch();
  const { token, user } = useSelector((state: RootState) => state.auth);
  const validatingRef = useRef(false);
  const [isInitialCheckDone, setIsInitialCheckDone] = useState(false);

  const validateAuth = useCallback(async () => {
    if (validatingRef.current || location.pathname.includes('/login')) {
      setIsInitialCheckDone(true);
      return;
    }

    if (!token) {
      setIsInitialCheckDone(true);
      navigate(`/login?redirect=${encodeURIComponent(location.pathname)}`);
      return;
    }

    if (!user) {
      try {
        validatingRef.current = true;
        const userData = await validateTokenAndGetUser();
        if (userData.user) {
          dispatch(setAuth({
            token,
            user: userData.user,
            needChangePwd: !!userData.needChangePwd
          }));
        }
      } catch (error) {
        dispatch(clearAuth());
        navigate(`/login?redirect=${encodeURIComponent(location.pathname)}`);
      } finally {
        validatingRef.current = false;
      }
    }
    
    setIsInitialCheckDone(true);
  }, [token, user, dispatch, navigate, location]);

  const login = async (data: LoginRequest) => {
    try {
      await loginApi(data);
      
      const params = new URLSearchParams(location.search);
      const redirect = params.get('redirect') || '/clips';
      setIsInitialCheckDone(true);
      validatingRef.current = false;
      navigate(redirect);
    } catch (error) {
      dispatch(clearAuth());
      throw error;
    }
  };

  const logout = () => {
    dispatch(clearAuth());
    setIsInitialCheckDone(false);
    validatingRef.current = false;
    navigate('/login');
  };

  useEffect(() => {
    validateAuth();
  }, [validateAuth]);

  return {
    user,
    token,
    login,
    logout,
    isInitialCheckDone
  };
};

export default useAuth;