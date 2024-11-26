import http from './http';
import { UploadClipRequest, Clip, ClipResponse, ListClipsResponse } from '@/store/types';
import axios from 'axios';

// 上传剪贴板内容
export const uploadClip = async (data: UploadClipRequest): Promise<Clip> => {
  const formData = new FormData();
  formData.append('spaceId', data.spaceId);
  formData.append('contentType', data.contentType);
  
  if (data.content) {
    formData.append('content', data.content);
  }
  
  if (data.file) {
    formData.append('file', data.file);
  }

  const response = await http.post<ClipResponse>('/clips/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  
  return response.data.clip;
};

// 获取剪贴板内容列表
export const getClips = async (spaceId: string): Promise<Clip[]> => {
  const response = await http.get<ListClipsResponse>(`/clips/list?spaceId=${spaceId}`);
  return response.data.clips;
};

// 获取单个剪贴板内容
export const getClip = async (id: string): Promise<Clip> => {
  const response = await http.get<ClipResponse>(`/clips/${id}`);
  return response.data.clip;
};

// 删除剪贴板内容
export const deleteClip = async (id: string): Promise<void> => {
  await http.delete(`/clips/${id}`);
};

// 下载文件
export const downloadClip = async (id: string): Promise<Blob> => {
  const response = await http.get(`/clips/${id}?download=true`, {
    responseType: 'blob',
  });
  return response.data;
};

// 获取剪贴板内容列表
export const listClips = async (spaceId: string) => {
    const response = await axios.get(`/api/v1/nlip/clips/${spaceId}`);
    return response.data;
}; 