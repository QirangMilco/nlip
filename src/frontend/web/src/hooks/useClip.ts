import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { message } from 'antd';
import * as clipApi from '@/api/clips';
import { Clip, UploadClipRequest } from '@/store/types';

export const useClips = (spaceId: string) => {
  const queryClient = useQueryClient();

  // 获取剪贴板列表
  const { data: clips = [], isLoading, error } = useQuery<Clip[]>({
    queryKey: ['clips', spaceId],
    queryFn: () => clipApi.getClips(spaceId),
    enabled: !!spaceId,
  });

  // 上传剪贴板内容
  const { mutateAsync: uploadClip, isPending: isUploading } = useMutation({
    mutationFn: (data: UploadClipRequest) => clipApi.uploadClip(spaceId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clips', spaceId] });
      message.success('上传成功');
    },
    onError: (error: Error) => {
      message.error(error.message || '上传失败');
    },
  });

  // 删除剪贴板内容
  const { mutateAsync: deleteClip, isPending: isDeleting } = useMutation({
    mutationFn: (clipId: string) => clipApi.deleteClip(spaceId, clipId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clips', spaceId] });
      message.success('删除成功');
    },
    onError: (error: Error) => {
      message.error(error.message || '删除失败');
    },
  });

  // 下载剪贴板内容
  const { mutateAsync: downloadClip, isPending: isDownloading } = useMutation({
    mutationFn: (clipId: string) => clipApi.downloadClip(spaceId, clipId),
    onError: (error: Error) => {
      message.error(error.message || '下载失败');
    },
  });

  return {
    clips,
    isLoading,
    error,
    uploadClip,
    isUploading,
    deleteClip,
    isDeleting,
    downloadClip,
    isDownloading,
  };
};

// 获取单个剪贴板内容的hook
export const useClip = (spaceId: string, clipId: string) => {
  const { data: clip, isLoading, error } = useQuery({
    queryKey: ['clip', spaceId, clipId],
    queryFn: () => clipApi.getClip(spaceId, clipId),
    enabled: !!spaceId && !!clipId,
  });

  return {
    clip,
    isLoading,
    error,
  };
}; 