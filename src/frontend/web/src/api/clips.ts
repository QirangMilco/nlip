import http from './http';
import { UploadClipRequest, Clip, ClipResponse, ListClipsResponse } from '@/store/types';

// 上传剪贴板内容
export const uploadClip = async (spaceId: string, data: UploadClipRequest): Promise<Clip> => {
  const formData = new FormData();
  formData.append('contentType', data.contentType);
  
  if (data.content) {
    formData.append('content', data.content);
  }
  
  if (data.file) {
    formData.append('file', data.file);
  }

  const response = await http.post<ClipResponse>(`/spaces/${spaceId}/clips/upload`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  
  return response.data.clip;
};

// 获取剪贴板内容列表
export const getClips = async (spaceId: string): Promise<Clip[]> => {
  const response = await http.get<ListClipsResponse>(`/spaces/${spaceId}/clips/list`);
  return response.data.clips;
};

// 获取单个剪贴板内容
export const getClip = async (spaceId: string, clipId: string): Promise<Clip> => {
  const response = await http.get<ClipResponse>(`/spaces/${spaceId}/clips/${clipId}`);
  return response.data.clip;
};

// 删除剪贴板内容
export const deleteClip = async (spaceId: string, clipId: string): Promise<void> => {
  await http.delete(`/spaces/${spaceId}/clips/${clipId}`);
};

// 下载文件
export const downloadClip = async (spaceId: string, clipId: string): Promise<Blob> => {
  const response = await http.get(`/spaces/${spaceId}/clips/${clipId}?download=true`, {
    responseType: 'blob',
  });
  return response.data;
};