import React from 'react';
import { Alert } from 'antd';
import { Space } from '@/store/types';

interface SpacePermissionAlertProps {
  space: Space;
  permission?: 'edit' | 'view' | null | undefined;
  isAdmin?: boolean;
}

const SpacePermissionAlert: React.FC<SpacePermissionAlertProps> = ({
  space,
  permission,
  isAdmin,
}) => {
  if (!space) return null;

  const isPublicSpace = space.type === 'public';
  const isOwner = space.ownerId === localStorage.getItem('userId');

  const getAlertType = () => {
    if (isAdmin) return 'success';
    if (isPublicSpace) return 'info';
    if (isOwner) return 'success';
    if (permission === 'edit') return 'success';
    if (permission === 'view') return 'warning';
    return 'error';
  };

  const getMessage = () => {
    if (isAdmin) {
      return '作为管理员，您可以管理此空间的所有内容。';
    }
    if (isPublicSpace) {
      return '您当前在公共空间，所有用户都可以查看内容。登录后可以上传和管理自己的内容。';
    }
    if (isOwner) {
      return '这是您创建的空间，您拥有完全的管理权限。';
    }
    if (permission === 'edit') {
      return '您被授予了编辑权限，可以添加、修改和删除内容。';
    }
    if (permission === 'view') {
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
