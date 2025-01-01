import React, { useState } from 'react';
import { Card, Progress, Tooltip, Typography, Space, Modal } from 'antd';
import { InfoCircleOutlined, TeamOutlined, FieldTimeOutlined, SettingOutlined, UserOutlined, LockOutlined, GlobalOutlined, ControlOutlined } from '@ant-design/icons';
import { SpaceWithPermission as SpaceType, Collaborator } from '@/store/types';
import CollaboratorManagement from './CollaboratorManagement';

const { Text } = Typography;

interface SpaceStatsProps {
  space: SpaceType;
  clipCount: number;
  ownerUsername: string;
  loading: boolean;
  collaborators: Collaborator[];
  canManageSpace: boolean;
  onOpenSettings: () => void;
  onSpaceUpdate: () => Promise<void>;
}

const SpaceStats: React.FC<SpaceStatsProps> = ({ 
  space, 
  clipCount, 
  ownerUsername, 
  loading, 
  collaborators = [], 
  canManageSpace, 
  onOpenSettings, 
  onSpaceUpdate 
}) => {
  const isPublicSpace = space.type === 'public';
  const usagePercent = Math.round((clipCount / space.maxItems) * 100);
  const collaboratorCount = Array.isArray(collaborators) ? collaborators.length : 0;
  const [showCollaborators, setShowCollaborators] = useState(false);

  return (
    <Card size="small" className="tw-card space-stats tw-p-2" loading={loading}>
      <Space direction="vertical" size={4} className="tw-w-full">
        <div className="tw-flex tw-justify-start tw-items-center">
          <Text strong className="tw-mr-3">空间情况</Text>
          <Tooltip title="显示当前空间的统计信息">
            <InfoCircleOutlined className="tw-text-primary-500" />
          </Tooltip>
        </div>

        {canManageSpace && (
          <button 
            className="tw-w-full tw-text-sm tw-px-2 tw-py-1.5 tw-rounded-md tw-transition-colors 
              tw-bg-gray-50 hover:tw-bg-gray-100 tw-text-gray-600"
            onClick={onOpenSettings}
          >
            <SettingOutlined className="tw-mr-1.5" />
            空间设置
          </button>
        )}

        <div className="tw-mt-1">
          <Text type="secondary">存储用量</Text>
          <Progress 
            percent={usagePercent} 
            status={usagePercent >= 90 ? "exception" : "normal"}
            size="small"
          />
          <div className="tw-flex tw-justify-between">
            <Text type="secondary">{clipCount} 个项目</Text>
            <Text type="secondary">上限 {space.maxItems}</Text>
          </div>
        </div>

        <div className="tw-mt-2 tw-flex tw-justify-between">
          <Tooltip title={isPublicSpace ? "所有用户可访问" : "仅特定用户可访问"}>
            <div className="tw-flex tw-items-center">
              {isPublicSpace ? <GlobalOutlined className="tw-mr-2" /> : <LockOutlined className="tw-mr-2" />}
              <Text className="tw-text-primary-500">
                {isPublicSpace ? "公共" : "私有"}
              </Text>
            </div>
          </Tooltip>

          <Tooltip title={`内容保留${space.retentionDays}天`}>
            <div className="tw-flex tw-items-center">
              <FieldTimeOutlined className="tw-mr-2" />
              <Text>{space.retentionDays}天</Text>
            </div>
          </Tooltip>
        </div>

        {!isPublicSpace && (
          <div className="tw-mt-2">
            <button 
              className="tw-w-full tw-text-sm tw-px-2 tw-py-1.5 tw-rounded-md tw-transition-colors 
                tw-bg-gray-50 hover:tw-bg-gray-100 tw-text-gray-600"
              onClick={() => setShowCollaborators(true)}
            >
              <ControlOutlined className="tw-mr-1.5" />
              管理协作者
            </button>

            <Modal
              title="协作者管理"
              open={showCollaborators}
              onCancel={() => setShowCollaborators(false)}
              footer={null}
              width={800}
              bodyStyle={{ padding: '1rem' }}
            >
              <CollaboratorManagement
                space={space}
                collaborators={collaborators}
                onCollaboratorUpdate={async () => {
                  await onSpaceUpdate();
                  setShowCollaborators(false);
                }}
              />
            </Modal>
          </div>
        )}

        {ownerUsername && (
          <div className="tw-mt-2">
            <div className="tw-mb-1 tw-flex tw-items-center">
              <TeamOutlined className="tw-mr-2" />
              <Text type="secondary">
                {collaboratorCount > 0 ? `${collaboratorCount}个协作者` : '暂无协作者'}
              </Text>
            </div>
            <div className="tw-flex tw-items-center">
              <UserOutlined className="tw-mr-2" />
              <Text type="secondary">空间所有者: {ownerUsername}</Text>
            </div>
          </div>
        )}
      </Space>
    </Card>
  );
};

export default SpaceStats; 