import { useCallback } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { message } from 'antd';
import { RootState } from '@/store';
import {
  setClips,
  setCurrentClip,
  addClip,
  deleteClip,
  setLoading,
  setError,
} from '@/store/slices/clipSlice';
import * as clipApi from '@/api/clips';
import { UploadClipRequest } from '@/store/types';
import { useNavigate } from 'react-router-dom';

export const useClip = () => {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { clips, currentClip, loading, error } = useSelector(
    (state: RootState) => state.clip
  );

  // 获取剪贴板内容列表
  const fetchClips = useCallback(async (spaceId: string) => {
    try {
      dispatch(setLoading(true));
      const data = await clipApi.listClips(spaceId);
      dispatch(setClips(data.data?.clips || []));
    } catch (error: any) {
      if (error.status === 401) {
        navigate('/login');
      }
      dispatch(setError(error.message));
      message.error(error.message);
    } finally {
      dispatch(setLoading(false));
    }
  }, [dispatch, navigate]);

  // 上传内容
  const uploadClip = useCallback(async (data: UploadClipRequest) => {
    try {
      dispatch(setLoading(true));
      const newClip = await clipApi.uploadClip(data);
      dispatch(addClip(newClip));
      message.success('上传成功');
      return newClip;
    } catch (error: any) {
      dispatch(setError(error.message));
      message.error(error.message);
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }, [dispatch]);

  // 删除内容
  const deleteClipById = useCallback(async (id: string) => {
    try {
      dispatch(setLoading(true));
      await clipApi.deleteClip(id);
      dispatch(deleteClip(id));
      message.success('删除成功');
    } catch (error: any) {
      dispatch(setError(error.message));
      message.error(error.message);
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }, [dispatch]);

  // 下载文件
  const downloadClip = useCallback(async (id: string, filename: string) => {
    try {
      dispatch(setLoading(true));
      const blob = await clipApi.downloadClip(id);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error: any) {
      dispatch(setError(error.message));
      message.error(error.message);
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }, [dispatch]);

  // 设置当前内容
  const selectClip = useCallback((clip: any) => {
    dispatch(setCurrentClip(clip));
  }, [dispatch]);

  return {
    clips,
    currentClip,
    loading,
    error,
    fetchClips,
    uploadClip,
    deleteClipById,
    downloadClip,
    selectClip,
  };
}; 