import React, { useState } from 'react';
import { Table, Button, Select, Popconfirm, message, Input, Tag } from 'antd';
import { DeleteOutlined, UserOutlined, SearchOutlined, UserDeleteOutlined } from '@ant-design/icons';
import { Collaborator, SpaceWithPermission as SpaceType } from '@/store/types';
import { updateCollaboratorPermission, removeCollaborator } from '@/api/spaces';
import InviteCollaboratorModal from './InviteCollaboratorModal';

interface CollaboratorManagementProps {
  space: SpaceType;
  collaborators: Collaborator[];
  onCollaboratorUpdate: () => void;
}

const CollaboratorManagement: React.FC<CollaboratorManagementProps> = ({ 
  space, 
  collaborators = [], 
  onCollaboratorUpdate 
}) => {
  // const [loading, setLoading] = useState(false);
  const [inviteModalVisible, setInviteModalVisible] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [selectedRows, setSelectedRows] = useState<string[]>([]);

  const handlePermissionChange = async (collaboratorId: string | undefined, newPermission: 'edit' | 'view') => {
    if (!collaboratorId) {
      message.error('协作者ID无效');
      return;
    }
    try {
      // setLoading(true);
      await updateCollaboratorPermission(space.id, collaboratorId, newPermission);
      message.success('权限更新成功');
      onCollaboratorUpdate();
    } catch (error: any) {
      message.error(error.message || '权限更新失败');
    } finally {
      // setLoading(false);
    }
  };

  const handleRemoveCollaborator = async (collaboratorId: string | undefined) => {
    if (!collaboratorId) {
      message.error('协作者ID无效');
      return;
    }
    try {
      // setLoading(true);
      await removeCollaborator(space.id, collaboratorId);
      message.success('移除协作者成功');
      onCollaboratorUpdate();
    } catch (error: any) {
      message.error(error.message || '移除协作者失败');
    } finally {
      // setLoading(false);
    }
  };

  const handleBatchRemove = async () => {
    try {
      // setLoading(true);
      await Promise.all(selectedRows.map(id => removeCollaborator(space.id, id)));
      message.success('批量移除协作者成功');
      setSelectedRows([]);
      onCollaboratorUpdate();
    } catch (error: any) {
      message.error(error.message || '批量移除失败');
    } finally {
      // setLoading(false);
    }
  };

  const ExtraContent = () => (
    <div className="tw-flex tw-flex-col sm:tw-flex-row tw-gap-2 sm:tw-gap-4 tw-items-start sm:tw-items-center">
      <Input
        placeholder="搜索协作者"
        prefix={<SearchOutlined />}
        onChange={e => setSearchText(e.target.value)}
        className="tw-w-full sm:tw-w-48 md:tw-w-64"
      />
      {space.isOwner && (
        <>
          {selectedRows.length > 0 && (
            <Popconfirm
              title={`确定要移除选中的 ${selectedRows.length} 个协作者吗？`}
              onConfirm={handleBatchRemove}
            >
              <Button
                type="primary"
                danger
                icon={<UserDeleteOutlined />}
                className="tw-w-full sm:tw-w-auto"
              >
                批量移除
              </Button>
            </Popconfirm>
          )}
          <Button
            type="primary"
            icon={<UserOutlined />}
            onClick={() => setInviteModalVisible(true)}
            className="tw-w-full sm:tw-w-auto"
          >
            邀请协作者
          </Button>
        </>
      )}
    </div>
  );

  const columns = [
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
      filterable: true,
      ellipsis: true,
      width: '40%',
      render: (username: string, record: Collaborator) => (
        <div className="tw-flex tw-items-center tw-gap-2 tw-truncate">
          <span className="tw-truncate">{username}</span>
          {record.id === space.ownerId && (
            <Tag color="gold" className="tw-shrink-0">拥有者</Tag>
          )}
        </div>
      ),
    },
    {
      title: '权限',
      dataIndex: 'permission',
      key: 'permission',
      render: (permission: string, record: Collaborator) => {
        if (space.isOwner && record.id !== space.ownerId) {
          return (
            <Select
              value={permission}
              onChange={(value) => handlePermissionChange(record.id, value as 'edit' | 'view')}
              style={{ width: 120 }}
            >
              <Select.Option value="edit">可编辑</Select.Option>
              <Select.Option value="view">可查看</Select.Option>
            </Select>
          );
        } else {
          return <span>{permission === 'edit' ? '可编辑' : '可查看'}</span>;
        }
      },
    },
    // 只有空间拥有者才显示操作列
    ...(space.isOwner ? [{
      title: '操作',
      key: 'action',
      render: (_: any, record: Collaborator) => {
        return record.id !== space.ownerId && (
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
        );
      },
    }] : []),
  ];

  const filteredCollaborators = (collaborators || []).map(c => ({
    ...c,
    key: c.id
  })).filter(c => c.username.toLowerCase().includes(searchText.toLowerCase()));

  return (
    <>
      <div className="tw-mb-4">
        <ExtraContent />
      </div>

      <Table
        columns={columns}
        dataSource={filteredCollaborators}
        rowKey="id"
        pagination={{
          pageSize: 10,
          showSizeChanger: true,
          showTotal: total => `共 ${total} 条记录`,
        }}
        scroll={{ x: 'max-content' }}
        rowSelection={space.isOwner ? {
          selectedRowKeys: selectedRows,
          onChange: (keys) => setSelectedRows(keys as string[]),
          getCheckboxProps: (record) => ({
            disabled: record.id === space.ownerId
          })
        } : undefined}
      />

      {space.isOwner && (
        <InviteCollaboratorModal
          spaceId={space.id}
          visible={inviteModalVisible}
          onClose={() => {
            setInviteModalVisible(false);
            onCollaboratorUpdate();
          }}
        />
      )}
    </>
  );
};

export default CollaboratorManagement;
