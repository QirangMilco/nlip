import React from 'react';
import { Alert } from 'antd';
import { SpaceWithPermission } from '@/store/types';
import { store } from '@/store';

interface SpacePermissionAlertProps {
  space: SpaceWithPermission;
  isAdmin?: boolean;
  isGuest?: boolean;
}

const SpacePermissionAlert: React.FC<SpacePermissionAlertProps> = ({
  space,
  isAdmin,
  isGuest,
}) => {
  if (!space) return null;

  const isPublicSpace = space.type === 'public';
  const isOwner = space.ownerId === store.getState().auth.user?.id;

  const getAlertType = () => {
    if (isAdmin) return 'success';
    if (isGuest) return 'warning';
    if (isPublicSpace) return 'info';
    if (isOwner) return 'success';
    if (space.permission === 'edit') return 'info';
    if (space.permission === 'view') return 'success';
    return 'error';
  };

  const getMessage = () => {
    if (isAdmin) {
      return '作为管理员，您可以管理此空间的所有内容。';
    }
    if (isGuest) {
      return '您当前在公共空间，所有用户都可以查看内容。登录后可以上传和管理自己的内容。';
    }
    if (isPublicSpace) {
      return '您当前在公共空间，所有用户都可以查看内容。';
    }
    if (isOwner) {
      return '这是您创建的空间，您拥有完全的管理权限。';
    }
    if (space.permission === 'edit') {
      return '您被授予了编辑权限，可以添加、修改和删除内容。';
    }
    if (space.permission === 'view') {
      return '您被授予了查看权限，可以查看和复制内容。';
    }
    return '您没有访问此空间的权限。';
  };

  return (
    <Alert
      message={getMessage()}
      type={getAlertType()}
      showIcon
      style={{ marginBottom: 16 }}
    />
  );
};

export default SpacePermissionAlert;
