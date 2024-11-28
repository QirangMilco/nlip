import http from './http';
import { Clip } from '@/store/types';

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
export const deleteClip = async (spaceId: string, clipId: string): Promise<void> => {
  await http.delete(`/spaces/${spaceId}/clips/${clipId}`);
};

// 下载剪贴板内容
export const downloadClip = async (spaceId: string, clipId: string): Promise<Blob> => {
  const response = await http.get(`/spaces/${spaceId}/clips/${clipId}`, {
    params: { download: true },
    responseType: 'blob'
  });
  return response.data;
};