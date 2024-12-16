import React from 'react';
import { Card, Progress, Tooltip, Typography, Space } from 'antd';
import { InfoCircleOutlined, TeamOutlined, FieldTimeOutlined } from '@ant-design/icons';
import { SpaceWithPermission as SpaceType, Collaborator } from '@/store/types';
import CollaboratorManagement from './CollaboratorManagement';

const { Text } = Typography;

interface SpaceStatsProps {
  space: SpaceType;
  clipCount: number;
  ownerUsername: string;
  loading: boolean;
  collaborators: Collaborator[];
  onSpaceUpdate: () => Promise<void>;
}

const SpaceStats: React.FC<SpaceStatsProps> = ({ space, clipCount, ownerUsername, loading, collaborators = [], onSpaceUpdate }) => {
  const isPublicSpace = space.type === 'public';
  const usagePercent = Math.round((clipCount / space.maxItems) * 100);
  const collaboratorCount = Array.isArray(collaborators) ? collaborators.length : 0;

  return (
    <Card size="small" className="space-stats" loading={loading}>
      <Space direction="vertical" style={{ width: '100%' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Text strong>空间使用情况</Text>
          <Tooltip title="显示当前空间的使用统计信息">
            <InfoCircleOutlined style={{ color: '#1890ff' }} />
          </Tooltip>
        </div>

        <div style={{ marginTop: 16 }}>
          <Text type="secondary">存储用量</Text>
          <Progress 
            percent={usagePercent} 
            status={usagePercent >= 90 ? "exception" : "normal"}
            size="small"
          />
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <Text type="secondary">{clipCount} 个项目</Text>
            <Text type="secondary">上限 {space.maxItems}</Text>
          </div>
        </div>

        <div style={{ marginTop: 16, display: 'flex', justifyContent: 'space-between' }}>
          <Tooltip title={isPublicSpace ? "所有用户可访问" : `${collaboratorCount}个协作者`}>
            <div>
              <TeamOutlined style={{ marginRight: 8 }} />
              <Text>{isPublicSpace ? "公开" : `${collaboratorCount}个协作者`}</Text>
            </div>
          </Tooltip>

          <Tooltip title={`内容保留${space.retentionDays}天`}>
            <div>
              <FieldTimeOutlined style={{ marginRight: 8 }} />
              <Text>{space.retentionDays}天</Text>
            </div>
          </Tooltip>
        </div>

        {!isPublicSpace && (
          <div style={{ marginTop: 16 }}>
            <CollaboratorManagement 
              space={space} 
              collaborators={collaborators}
              onCollaboratorUpdate={onSpaceUpdate}
            />
          </div>
        )}

        {/* 显示空间所有者昵称 */}
        {ownerUsername && (
          <div style={{ marginTop: 16 }}>
            <Text type="secondary">空间所有者: {ownerUsername}</Text>
          </div>
        )}
      </Space>
    </Card>
  );
};

export default SpaceStats; 