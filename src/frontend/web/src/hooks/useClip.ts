import { useState, useCallback } from 'react';
import { Clip, UploadClipRequest} from '@/store/types';
import * as clipApi from '@/api/clips';
import { useQuery } from '@tanstack/react-query';

export const useClips = (spaceId: string) => {
  const [clips, setClips] = useState<Clip[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const fetchClips = useCallback(async () => {
    try {
      setIsLoading(true);
      const clips = await clipApi.getClips(spaceId);
      setClips(clips);
      setError(null);
    } catch (err) {
      setError(err as Error);
    } finally {
      setIsLoading(false);
    }
  }, [spaceId]);

  const uploadClip = useCallback(async (data: UploadClipRequest) => {
    try {
      // 创建 FormData 对象
      const formData = new FormData();
      if (data.file) {
        formData.append('file', data.file);
      }
      if (data.content) {
        formData.append('content', data.content);
      }
      formData.append('contentType', data.contentType);
      formData.append('spaceId', data.spaceId);

      const response = await clipApi.uploadClip(formData);
      if (!response || !response.clipId) {
        throw new Error('上传响应数据格式错误');
      }
      
      setClips(prev => Array.isArray(prev) ? [response, ...prev] : [response]);
      return response;
    } catch (err) {
      console.error('上传失败:', err);
      throw err;
    }
  }, []);

  const deleteClip = useCallback(async (clipId: string) => {
    try {
      await clipApi.deleteClip(spaceId, clipId);
      setClips(prev => Array.isArray(prev) ? prev.filter(clip => clip.id !== clipId) : []);
    } catch (err) {
      throw err;
    }
  }, [spaceId]);

  const downloadClip = useCallback(async (spaceId: string, clipId: string): Promise<Blob> => {
    try {
      return await clipApi.downloadClip(spaceId, clipId);
    } catch (err) {
      console.error('下载文件失败:', err);
      throw err;
    }
  }, []);

  return {
    clips,
    isLoading,
    error,
    uploadClip,
    deleteClip,
    downloadClip,
    fetchClips
  };
};

// 获取单个剪贴板内容的hook
export const useClip = (spaceId: string, clipId: string) => {
  const { data: clip, isLoading, error } = useQuery({
    queryKey: ['clip', spaceId, clipId],
    queryFn: () => clipApi.getClips(spaceId).then(clips => 
      clips.find(c => c.id === clipId)
    ),
    enabled: !!spaceId && !!clipId,
  });

  return {
    clip,
    isLoading,
    error,
  };
}; 