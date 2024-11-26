import http from './http';
import { LoginRequest, RegisterRequest, AuthResponse } from '@/store/types';
import { api } from '@/config/index';

export const login = async (data: LoginRequest): Promise<AuthResponse> => {
  return http.post('/auth/login', data);
};

export const register = async (data: RegisterRequest): Promise<AuthResponse> => {
  return http.post('/auth/register', data);
};

export const checkToken = async (): Promise<AuthResponse> => {
  return http.get('/auth/check');
};

export const refreshToken = async (): Promise<AuthResponse> => {
  try {
    const response = await http.post('/auth/refresh');
    return response.data;
  } catch (error) {
    throw error;
  }
};

export const validateTokenAndGetUser = async (token: string): Promise<AuthResponse> => {
  try {
    const response = await http.get('/auth/me', {
      headers: { Authorization: `Bearer ${token}` }
    });
    return response.data;
  } catch (error: any) {
    if (error.response?.status === 401) {
      try {
        const newToken = await refreshToken();
        const retryResponse = await http.get('/auth/me', {
          headers: { Authorization: `Bearer ${newToken.token}` }
        });
        return retryResponse.data;
      } catch (refreshError) {
        throw refreshError;
      }
    }
    throw error;
  }
}; 