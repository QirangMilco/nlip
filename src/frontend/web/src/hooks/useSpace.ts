import { useState, useCallback, useEffect } from 'react';
import { Space } from '@/store/slices/spaceSlice';
import { CreateSpaceRequest, UpdateSpaceRequest } from '@/store/types';
import * as spaceApi from '@/api/spaces';

// 移除重载定义，只保留一个返回类型
export function useSpace() {
  const [spaces, setSpaces] = useState<Space[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  // 获取空间列表
  const fetchSpaces = useCallback(async () => {
    try {
      setLoading(true);
      console.log('fetchSpaces time', new Date().toLocaleString());
      const data = await spaceApi.listSpaces();
      setSpaces(data);
      setError(null);
    } catch (err) {
      setError(err as Error);
    } finally {
      setLoading(false);
      console.log('fetchSpaces done time', new Date().toLocaleString());
    }
  }, []);

  // 初始化时获取空间列表
  useEffect(() => {
    fetchSpaces();
  }, [fetchSpaces]);

  const createSpace = useCallback(async (data: CreateSpaceRequest) => {
    try {
      setLoading(true);
      const newSpace = await spaceApi.createSpace(data);
      setSpaces(prev => [...prev, newSpace]);
      setError(null);
      return newSpace;
    } catch (err) {
      setError(err as Error);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const updateSpaceById = useCallback(async (id: string, data: UpdateSpaceRequest) => {
    try {
      setLoading(true);
      const updatedSpace = await spaceApi.updateSpace(id, data);
      setSpaces(prev => prev.map(space => 
        space.id === id ? updatedSpace : space
      ));
      setError(null);
      return updatedSpace;
    } catch (err) {
      setError(err as Error);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const deleteSpaceById = useCallback(async (id: string) => {
    try {
      setLoading(true);
      await spaceApi.deleteSpace(id);
      setSpaces(prev => prev.filter(space => space.id !== id));
      setError(null);
    } catch (err) {
      setError(err as Error);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    spaces,
    loading,
    error,
    fetchSpaces,
    createSpace,
    updateSpaceById,
    deleteSpaceById
  };
} 