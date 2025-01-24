import http from './http';
import { Clip } from '@/store/types';
import { store } from '@/store';

// 获取空间下的所有剪贴板内容
export const getClips = async (spaceId: string): Promise<Clip[]> => {
  const response = await http.get(`/spaces/${spaceId}/clips/list`);
  return response.data.clips;
};

// 上传新的剪贴板内容
export const uploadClip = async (data: FormData): Promise<Clip> => {
  const spaceId = data.get('spaceId');
  const response = await http.post(`/spaces/${spaceId}/clips/upload`, data, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data.clip;
};

// 删除剪贴板内容
export const deleteClip = async (spaceId: string, clipId: string, spaceType: string): Promise<string> => {
  // 从 Redux store 获取认证状态
  if (spaceType === 'public' && !store.getState().auth.token) {
    throw new Error('游客无权删除剪贴板内容');
  }
  const response = await http.delete(`/spaces/${spaceId}/clips/${clipId}`);
  return response.data;
};

// 下载剪贴板内容
export const downloadClip = async (spaceId: string, clipId: string): Promise<Blob> => {
  const response = await http.get(`/spaces/${spaceId}/clips/${clipId}`, {
    params: { download: true },
    responseType: 'blob',
    headers: {
      'Accept': '*/*'
    }
  });
  
  return response.data;
};

// 更新剪贴板内容
export const updateClip = async (spaceId: string, clipId: string, content: string, spaceType: string): Promise<Clip> => {
  // 检查是否有权限更新
  if (spaceType === 'public' && !store.getState().auth.token) {
    throw new Error('游客无权修改剪贴板内容');
  }
  const response = await http.put(`/spaces/${spaceId}/clips/${clipId}`, { content });
  return response.data.clip;
};