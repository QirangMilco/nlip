import http from './http';
import { CreateTokenRequest, CreateTokenResponse, GetTokenResponse, ListTokensResponse } from '@/store/types';


// 创建Token
export const createToken = async (data: CreateTokenRequest): Promise<CreateTokenResponse> => {
  const response = await http.post('/token/create', data);
  return response.data.data;
};

// 获取Token列表
export const listTokens = async (): Promise<ListTokensResponse> => {
  const response = await http.get('/token/list');
  return response.data.data;
};

// 获取单个Token
export const getToken = async (tokenId: string): Promise<GetTokenResponse> => {
  const response = await http.get(`/token/${tokenId}`);
  return response.data.data;
};

// 删除Token
export const deleteToken = async (tokenId: string): Promise<void> => {
  await http.delete(`/token/${tokenId}`);
};
