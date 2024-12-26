import React from 'react';
import { PlusOutlined } from '@ant-design/icons';
import { Space, User, SpaceStats, Collaborator } from '@/store/types';
import SpaceList from './SpaceList';
import SpaceStatsCard from './SpaceStatsCard';

interface SpaceMenuProps {
  // 基础数据
  spaces: Space[];
  currentSpace: Space | undefined;
  currentUser: User | null;
  token: string | null;
  spaceId: string | undefined;
  
  // 状态
  loadingSpaces: boolean;
  loadingStats: boolean;
  loadingCollaborators: boolean;
  
  // 数据
  spaceStats: SpaceStats | null;
  collaborators: Collaborator[];
  
  // 权限
  canManageSpace: boolean;
  
  // 事件处理
  onSpaceChange: (spaceId: string) => void;
  onCreateSpace: () => void;
  onOpenSettings: () => void;
  onSpaceUpdate: () => Promise<void>;
}

const SpaceMenu: React.FC<SpaceMenuProps> = ({
  spaces,
  currentSpace,
  currentUser,
  token,
  spaceId,
  loadingSpaces,
  loadingStats,
  loadingCollaborators,
  spaceStats,
  collaborators,
  canManageSpace,
  onSpaceChange,
  onCreateSpace,
  onOpenSettings,
  onSpaceUpdate
}) => {
  return (
    <div className="tw-sticky tw-top-0 tw-left-0 tw-shrink-0 tw-w-56 tw-h-full tw-p-4">
      {/* 空间列表 */}
      {currentUser && token && (
        <div className="tw-mb-8 tw-bg-white tw-rounded-lg tw-shadow-sm tw-p-4">
          <div className="tw-flex tw-items-center tw-justify-between tw-mb-4">
          <h2 className="tw-text-sm tw-font-medium tw-text-gray-700">我的空间</h2>
          <button
            className="tw-text-sm tw-px-2 tw-py-1.5 tw-rounded-md tw-transition-colors 
              tw-bg-gray-50 hover:tw-bg-gray-100 tw-text-gray-600"
            onClick={onCreateSpace}
          >
            <PlusOutlined className="tw-mr-1.5" />
            创建空间
          </button>
        </div>
          <SpaceList
            spaces={spaces}
            currentUser={currentUser}
            value={spaceId}
            loading={loadingSpaces}
            onChange={onSpaceChange}
          />
        </div>
      )}

      {/* 空间统计信息 */}
      {currentSpace && (
        <div className="tw-mb-6">
          <SpaceStatsCard 
            space={currentSpace}
            clipCount={spaceStats?.clipCount || 0}
            ownerUsername={spaceStats?.ownerUsername || ''}
            collaborators={collaborators}
            loading={loadingStats || loadingCollaborators}
            onSpaceUpdate={onSpaceUpdate}
            canManageSpace={canManageSpace}
            onOpenSettings={onOpenSettings}
          />
        </div>
      )}
    </div>
  );
};

export default SpaceMenu;
