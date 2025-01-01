import http from './http';
import { LoginRequest, RegisterRequest, AuthResponse, ChangePasswordRequest, ChangePasswordResponse } from '@/store/types';
import { store } from '@/store';
import { setAuth } from '@/store/slices/authSlice';

export const login = async (data: LoginRequest): Promise<AuthResponse> => {
  const response = await http.post<AuthResponse>('/auth/login', data);
  if (!response.data.user) {
    throw new Error('User data is required');
  }
  store.dispatch(setAuth({
    token: response.data.token,
    user: response.data.user,
    needChangePwd: response.data.needChangePwd
  }));
  
  return response.data;
};

export const register = async (data: RegisterRequest): Promise<AuthResponse> => {
  const response = await http.post('/auth/register', data);
  return response.data;
};

export const checkToken = async (): Promise<AuthResponse> => {
  const response = await http.get<AuthResponse>('/auth/check');
  if (!response.data.user) {
    throw new Error('User data is required');
  }
  store.dispatch(setAuth({
    token: response.data.token,
    user: response.data.user,
    needChangePwd: response.data.needChangePwd
  }));
  return response.data;
};

export const refreshToken = async (): Promise<AuthResponse> => {
  try {
    const response = await http.post('/auth/refresh');
    if (!response.data.user) {
      throw new Error('User data is required');
    }
    store.dispatch(setAuth({
      token: response.data.token,
      user: response.data.user,
      needChangePwd: response.data.needChangePwd
    }));
    return response.data;
  } catch (error) {
    throw error;
  }
};

export const validateTokenAndGetUser = async (): Promise<AuthResponse> => {
  try {
    const response = await http.get<AuthResponse>('/auth/me');
    if (!response.data.user) {
      throw new Error('User data is required');
    }
    store.dispatch(setAuth({
      token: response.data.token,
      user: response.data.user,
      needChangePwd: response.data.needChangePwd
    }));
    return response.data;
  } catch (error: any) {
    if (error.response?.status === 401) {
      try {
        const retryResponse = await http.get<AuthResponse>('/auth/me');
        if (!retryResponse.data.user) {
          throw new Error('User data is required');
        }
        store.dispatch(setAuth({
          token: retryResponse.data.token,
          user: retryResponse.data.user,
          needChangePwd: retryResponse.data.needChangePwd
        }));
        return retryResponse.data;
      } catch (refreshError) {
        throw refreshError;
      }
    }
    throw error;
  }
};

export const changePassword = async (data: ChangePasswordRequest): Promise<ChangePasswordResponse> => {
  const response = await http.post('/auth/change-password', data);
  console.log(response.data);
  return response.data;
};
