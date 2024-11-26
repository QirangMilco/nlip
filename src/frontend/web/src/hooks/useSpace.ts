import { useCallback, useRef, useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { message } from 'antd';
import { RootState } from '@/store';
import {
  setSpaces,
  setCurrentSpace,
  addSpace,
  updateSpace,
  deleteSpace,
  setLoading,
  setError,
} from '@/store/slices/spaceSlice';
import * as spaceApi from '@/api/spaces';
import { CreateSpaceRequest, UpdateSpaceRequest } from '@/store/types';
import { useNavigate } from 'react-router-dom';
import { throttle } from 'lodash';
import { useAuth } from './useAuth';

export const useSpace = () => {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { user, isInitialCheckDone } = useAuth();
  const { spaces, currentSpace, loading, error } = useSelector(
    (state: RootState) => state.space
  );
  const fetchingRef = useRef(false);

  // 获取空间列表
  const fetchSpaces = useCallback(async () => {
    // 如果未完成初始认证检查或正在获取数据，则跳过
    if (!isInitialCheckDone || !user || fetchingRef.current) {
      return;
    }

    try {
      fetchingRef.current = true;
      dispatch(setLoading(true));
      const spaces = await spaceApi.listSpaces();
      dispatch(setSpaces(spaces || []));
    } catch (error: any) {
      if (error.status === 401) {
        navigate('/login');
      }
      dispatch(setError(error.message));
      message.error(error.message);
    } finally {
      dispatch(setLoading(false));
      fetchingRef.current = false;
    }
  }, [dispatch, navigate, isInitialCheckDone, user]);

  // 在组件挂载和认证状态变化时获取数据
  useEffect(() => {
    if (isInitialCheckDone && user) {
      fetchSpaces();
    }
  }, [fetchSpaces, isInitialCheckDone, user]);

  // 创建空间
  const createSpace = useCallback(async (data: CreateSpaceRequest) => {
    try {
      dispatch(setLoading(true));
      const newSpace = await spaceApi.createSpace(data);
      dispatch(addSpace(newSpace));
      message.success('创建空间成功');
      return newSpace;
    } catch (error: any) {
      dispatch(setError(error.message));
      message.error(error.message);
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }, [dispatch]);

  // 更新空间 - 使用 throttle 代替 debounce
  const throttledUpdateSpace = useCallback(
    throttle(async (id: string, data: UpdateSpaceRequest) => {
      try {
        dispatch(setLoading(true));
        const updatedSpace = await spaceApi.updateSpace(id, data);
        dispatch(updateSpace(updatedSpace));
        message.success('更新空间成功');
        return updatedSpace;
      } catch (error: any) {
        dispatch(setError(error.message));
        message.error(error.message);
        throw error;
      } finally {
        dispatch(setLoading(false));
      }
    }, 1000), // 1秒内最多执行一次
    [dispatch]
  );

  // 删除空间
  const deleteSpaceById = useCallback(async (id: string) => {
    try {
      dispatch(setLoading(true));
      await spaceApi.deleteSpace(id);
      dispatch(deleteSpace(id));
      message.success('删除空间成功');
    } catch (error: any) {
      dispatch(setError(error.message));
      message.error(error.message);
      throw error;
    } finally {
      dispatch(setLoading(false));
    }
  }, [dispatch]);

  // 设置当前空间 - 使用 throttle
  const throttledSelectSpace = useCallback(
    throttle((space: any) => {
      dispatch(setCurrentSpace(space));
    }, 500), // 500ms 内最多执行一次
    [dispatch]
  );

  return {
    spaces,
    currentSpace,
    loading,
    error,
    fetchSpaces,
    createSpace,
    updateSpaceById: throttledUpdateSpace,
    deleteSpaceById,
    selectSpace: throttledSelectSpace,
  };
};

export default useSpace; 