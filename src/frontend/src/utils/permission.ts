import { Space, User } from '@/store/types';

export const checkSpaceAccess = (space: Space, user?: User | null): boolean => {
  if (!space) return false;

  // 公共空间所有人都能访问
  if (space.type === 'public') {
    return true;
  }

  // 未登录用户只能访问公共空间
  if (!user) {
    return false;
  }

  // 管理员可以访问所有空间
  if (user.isAdmin) {
    return true;
  }

  // 空间所有者可以访问
  if (space.ownerId === user.id) {
    return true;
  }

  // 被邀请的用户可以访问
  return !!space.collaborators?.find((collaborator) => collaborator.id === user.id);
};

export const getSpacePermission = (space: Space, user?: User | null): 'edit' | 'view' | null => {
  if (!space || !user) return null;

  // 管理员和所有者有完全权限
  if (user?.isAdmin || space.ownerId === user?.id) {
    return 'edit';
  }

  // 公共空间登录用户可编辑，未登录用户可查看
  if (space.type === 'public') {
    return user ? 'edit' : 'view';
  }

  // 返回邀请用户的权限
  return space.collaborators?.find((collaborator) => collaborator.id === user.id)?.permission || null;
};
