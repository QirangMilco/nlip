import React, { useState } from 'react';
import { Card, Table, Button, Select, Popconfirm, message } from 'antd';
import { DeleteOutlined, UserOutlined } from '@ant-design/icons';
import { Space as SpaceType } from '@/store/types';
import { updateCollaboratorPermission, removeCollaborator } from '@/api/spaces';
import InviteCollaboratorModal from './InviteCollaboratorModal';

interface CollaboratorManagementProps {
  space: SpaceType;
  onCollaboratorUpdate: () => void;
}

const CollaboratorManagement: React.FC<CollaboratorManagementProps> = ({ 
  space, 
  onCollaboratorUpdate 
}) => {
  const [loading, setLoading] = useState(false);
  const [inviteModalVisible, setInviteModalVisible] = useState(false);

  const handlePermissionChange = async (collaboratorId: string, newPermission: 'edit' | 'view') => {
    try {
      setLoading(true);
      await updateCollaboratorPermission(space.id, collaboratorId, newPermission);
      message.success('权限更新成功');
      onCollaboratorUpdate();
    } catch (error: any) {
      message.error(error.message || '权限更新失败');
    } finally {
      setLoading(false);
    }
  };

  const handleRemoveCollaborator = async (collaboratorId: string) => {
    try {
      setLoading(true);
      await removeCollaborator(space.id, collaboratorId);
      message.success('移除协作者成功');
      onCollaboratorUpdate();
    } catch (error: any) {
      message.error(error.message || '移除协作者失败');
    } finally {
      setLoading(false);
    }
  };

  const columns = [
    {
      title: '用户',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: '权限',
      dataIndex: 'permission',
      key: 'permission',
      render: (permission: string, record: any) => (
        <Select
          value={permission}
          onChange={(value) => handlePermissionChange(record.id, value as 'edit' | 'view')}
          disabled={record.id === space.ownerId}
        >
          <Select.Option value="edit">可编辑</Select.Option>
          <Select.Option value="view">可查看</Select.Option>
        </Select>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: any) => (
        record.id !== space.ownerId && (
          <Popconfirm
            title="确定要移除该协作者吗？"
            onConfirm={() => handleRemoveCollaborator(record.id)}
          >
            <Button 
              type="text" 
              danger 
              icon={<DeleteOutlined />}
            >
              移除
            </Button>
          </Popconfirm>
        )
      ),
    },
  ];

  return (
    <>
      <Card 
        title="协作者管理" 
        loading={loading}
        extra={
          <Button
            type="primary"
            icon={<UserOutlined />}
            onClick={() => setInviteModalVisible(true)}
          >
            邀请协作者
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={space.collaborators}
          rowKey="id"
          pagination={false}
        />
      </Card>

      <InviteCollaboratorModal
        spaceId={space.id}
        visible={inviteModalVisible}
        onClose={() => {
          setInviteModalVisible(false);
          onCollaboratorUpdate();
        }}
      />
    </>
  );
};

export default CollaboratorManagement;
