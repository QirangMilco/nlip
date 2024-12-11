import { useState, useCallback, useEffect } from 'react';
import { SpacePermission, SpaceWithPermission } from '@/store/types';
import { CreateSpaceRequest, UpdateSpaceRequest } from '@/store/types';
import * as spaceApi from '@/api/spaces';
import { store } from '@/store';


// 移除重载定义，只保留一个返回类型
export function useSpace() {
  const [spaces, setSpaces] = useState<SpaceWithPermission[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  // 获取空间列表并处理权限
  const fetchSpaces = useCallback(async () => {
    try {
      setLoading(true);
      const data = await spaceApi.listSpaces();
      
      // 处理权限信息，确保类型正确
      const spacesWithPermissions = data.map(space => ({
        ...space,
        isOwner: space.ownerId === store.getState().auth.user?.id,
        permission: space.collaborators?.find(
          collab => collab.id === store.getState().auth.user?.id
        )?.permission as SpacePermission | undefined
      }));

      setSpaces(spacesWithPermissions);
      setError(null);
    } catch (err) {
      setError(err as Error);
    } finally {
      setLoading(false);
    }
  }, []);

  // 检查用户是否有权限执行特定操作
  const checkPermission = useCallback((spaceId: string, spaceType: string, requiredPermission: SpacePermission = 'edit') => {
    const space = spaces.find(s => s.id === spaceId);
    const isAdmin = store.getState().auth.user?.isAdmin;
    
    if (!space) return false;
    
    // 公共空间特殊处理
    if (spaceType === 'public') {
      return true; // 所有人都可以访问公共空间
    }

    // 空间所有者和管理员有完全权限
    if (space.isOwner || isAdmin) {
      return true;
    }

    // 检查被邀请用户的权限
    if (requiredPermission === 'view') {
      return !!space.permission; // 任何权限都可以查看
    } else {
      return space.permission === 'edit'; // 只有编辑权限才能修改
    }
  }, [spaces]);

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
    deleteSpaceById,
    checkPermission
  };
} 