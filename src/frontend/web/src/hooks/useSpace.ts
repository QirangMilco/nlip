import { useState, useCallback, useEffect } from 'react';
import { Space } from '@/store/slices/spaceSlice';
import { CreateSpaceRequest, UpdateSpaceRequest } from '@/store/types';
import * as spaceApi from '@/api/spaces';

// 重载 useSpace hook 以支持两种使用方式
export function useSpace(): {
  spaces: Space[];
  loading: boolean;
  error: Error | null;
  fetchSpaces: () => Promise<void>;
  createSpace: (data: CreateSpaceRequest) => Promise<Space>;
  updateSpaceById: (id: string, data: UpdateSpaceRequest) => Promise<Space>;
  deleteSpaceById: (id: string) => Promise<void>;
};
export function useSpace(spaceId: string): {
  space: Space | null;
  isLoading: boolean;
  error: Error | null;
  fetchSpace: () => Promise<void>;
};
export function useSpace(spaceId?: string) {
  const [spaces, setSpaces] = useState<Space[]>([]);
  const [space, setSpace] = useState<Space | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  // 获取单个空间
  const fetchSpace = useCallback(async () => {
    if (!spaceId) return;
    try {
      setLoading(true);
      const data = await spaceApi.getSpace(spaceId);
      setSpace(data);
      setError(null);
    } catch (err) {
      setError(err as Error);
      setSpace(null);
    } finally {
      setLoading(false);
    }
  }, [spaceId]);

  // 获取空间列表
  const fetchSpaces = useCallback(async () => {
    try {
      setLoading(true);
      const data = await spaceApi.listSpaces();
      setSpaces(data);
      setError(null);
    } catch (err) {
      setError(err as Error);
    } finally {
      setLoading(false);
    }
  }, []);

  // 根据是否传入 spaceId 决定获取单个空间还是空间列表
  useEffect(() => {
    if (spaceId) {
      fetchSpace();
    } else {
      fetchSpaces();
    }
  }, [spaceId, fetchSpace, fetchSpaces]);

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

  if (spaceId) {
    return {
      space,
      isLoading: loading,
      error,
      fetchSpace
    };
  }

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