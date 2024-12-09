import { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { message } from 'antd';
import { Space } from '@/store/types';
import { store } from '@/store';

export function useSpaceNavigation() {
  const navigate = useNavigate();

  const handleSpaceChange = useCallback((spaceId: string, spaceType: string, spaces: Space[]) => {
    const targetSpace = spaces.find(s => s.id === spaceId);
    const userId = store.getState().auth.user?.id;
    const isAdmin = store.getState().auth.user?.isAdmin;

    if (!targetSpace) {
      message.error('空间不存在');
      return;
    }

    // 检查访问权限
    const isPublicSpace = spaceType === 'public';
    const isOwner = targetSpace.ownerId === userId;
    const hasPermission = targetSpace.invitedUsers?.[userId || ''];

    if (!isPublicSpace && !isOwner && !hasPermission && !isAdmin) {
      message.error('您没有权限访问此空间');
      return;
    }

    // 导航到新空间
    navigate(`/clips/${spaceId}`);
  }, [navigate]);

  return {
    handleSpaceChange
  };
}
