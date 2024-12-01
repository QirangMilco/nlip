import http from './http';
import { Clip } from '@/store/types';
import { store } from '@/store';
import { SPACE_CONSTANTS } from '@/constants/spaces';

// 获取空间下的所有剪贴板内容
export const getClips = async (spaceId: string): Promise<Clip[]> => {
  const response = await http.get(`/spaces/${spaceId}/clips/list`);
  return response.data.clips;
};

// 上传新的剪贴板内容
export const uploadClip = async (data: FormData): Promise<Clip> => {
  const spaceId = data.get('spaceId');
  
  // 判断是否为公共空间
  if (spaceId === SPACE_CONSTANTS.PUBLIC_SPACE_ID) {
    // 根据用户是否登录选择不同的上传路径
    const isAuthenticated = !!store.getState().auth.token;
    const uploadPath = isAuthenticated 
      ? '/spaces/public-space/clips/upload'
      : '/spaces/public-space/clips/guest-upload';
      
    if (!isAuthenticated) {
      data.append('creator', SPACE_CONSTANTS.GUEST_USER);
    }
    
    const response = await http.post(uploadPath, data, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data.clip;
  }
  
  // 普通空间上传
  const response = await http.post(`/spaces/${spaceId}/clips/upload`, data, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data.clip;
};

// 删除剪贴板内容
export const deleteClip = async (spaceId: string, clipId: string): Promise<void> => {
  // 检查是否有权限删除
  if (spaceId === SPACE_CONSTANTS.PUBLIC_SPACE_ID && !store.getState().auth.token) {
    throw new Error('游客无权删除剪贴板内容');
  }
  await http.delete(`/spaces/${spaceId}/clips/${clipId}`);
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
export const updateClip = async (spaceId: string, clipId: string, content: string): Promise<Clip> => {
  // 检查是否有权限更新
  if (spaceId === SPACE_CONSTANTS.PUBLIC_SPACE_ID && !store.getState().auth.token) {
    throw new Error('游客无权修改剪贴板内容');
  }
  const response = await http.put(`/spaces/${spaceId}/clips/${clipId}`, { content });
  return response.data.clip;
};