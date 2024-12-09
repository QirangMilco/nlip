import React, { useState, useEffect } from 'react';
import { Modal, Form, Input, InputNumber, message, Select, Button, Table } from 'antd';
import { Space } from '@/store/types';
import { updateSpace } from '@/api/spaces';
import { getSettings } from '@/api/admin';
import { useSelector } from 'react-redux';
import { RootState } from '@/store';
import { SPACE_CONSTANTS } from '@/constants/spaces';
import { updateCollaboratorPermission, removeCollaborator, inviteCollaborator } from '@/api/spaces';
import { ExclamationCircleOutlined } from '@ant-design/icons';
import { deleteSpace } from '@/api/spaces';

interface SpaceSettingsModalProps {
  visible: boolean;
  space: Space;
  onClose: () => void;
  onSpaceUpdated: (action?: 'delete') => Promise<void>;
}

const SpaceSettingsModal: React.FC<SpaceSettingsModalProps> = ({
  visible,
  space,
  onClose,
  onSpaceUpdated,
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [maxItemsLimit, setMaxItemsLimit] = useState<number>(100);
  const [maxRetentionDaysLimit, setMaxRetentionDaysLimit] = useState<number>(30);
  
  // 获取当前用户信息
  const currentUser = useSelector((state: RootState) => state.auth.user);
  const isAdmin = currentUser?.isAdmin;
  const isOwner = space.ownerId === currentUser?.id;
  const isPublicSpace = space.type === 'public';

  // 修改权限检查逻辑
  const canManageSpace = isAdmin || isOwner;

  useEffect(() => {
    if (visible && space) {
      form.setFieldsValue({
        name: space.name,
        maxItems: space.maxItems,
        retentionDays: space.retentionDays,
        type: space.type,
      });
    }
  }, [visible, space, form]);

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const settings = await getSettings();
        setMaxItemsLimit(settings.max_items_limit);
        setMaxRetentionDaysLimit(settings.max_retention_days_limit);
      } catch (error) {
        message.error('获取设置失败');
      }
    };

    fetchSettings();
  }, []);

  const handleSubmit = async () => {
    if (!canManageSpace) {
      message.error('没有权限修改空间设置');
      return;
    }

    try {
      setLoading(true);
      const values = await form.validateFields();
      await updateSpace(space.id, values);
      message.success('空间设置已更新');
      await onSpaceUpdated();
      onClose();
    } catch (error: any) {
      message.error(error.message || '更新空间设置失败');
    } finally {
      setLoading(false);
    }
  };

  const handlePermissionChange = async (userId: string, newPermission: string) => {
    if (!canManageSpace) return;
    
    try {
      setLoading(true);
      await updateCollaboratorPermission(space.id, userId, newPermission as 'edit' | 'view');
      message.success('权限更新成功');
      await onSpaceUpdated();
    } catch (error: any) {
      message.error(error.message || '更新权限失败');
    } finally {
      setLoading(false);
    }
  };

  const handleRemoveCollaborator = async (userId: string) => {
    if (!canManageSpace) return;

    try {
      setLoading(true);
      await removeCollaborator(space.id, userId);
      message.success('移除协作者成功');
      await onSpaceUpdated();
    } catch (error: any) {
      message.error(error.message || '移除协作者失败');
    } finally {
      setLoading(false);
    }
  };

  // 添加邀请协作者的功能
  const [inviteModalVisible, setInviteModalVisible] = useState(false);
  const [inviteForm] = Form.useForm();

  const handleInviteCollaborator = async () => {
    try {
      const values = await inviteForm.validateFields();
      await inviteCollaborator(space.id, values.userId, values.permission);
      message.success('邀请协作者成功');
      setInviteModalVisible(false);
      inviteForm.resetFields();
      await onSpaceUpdated();
    } catch (error: any) {
      message.error(error.message || '邀请协作者失败');
    }
  };

  // 添加删除空间的处理函数
  const handleDeleteSpace = () => {
    if (!canManageSpace) return;
    if (space.id === SPACE_CONSTANTS.PUBLIC_SPACE_ID) {
      message.error('不能删除默认公共空间');
      return;
    }

    Modal.confirm({
      title: '确认删除空间',
      icon: <ExclamationCircleOutlined />,
      content: '删除空间后，所有相关的剪贴板内容将被永久删除。此操作不可恢复，是否继续？',
      okText: '确认删除',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          setLoading(true);
          await deleteSpace(space.id);
          message.success('空间已删除');
          onClose();
          await onSpaceUpdated('delete');
        } catch (error: any) {
          message.error(error.message || '删除空间失败');
        } finally {
          setLoading(false);
        }
      },
    });
  };

  // 渲染协作者列表
  const renderCollaborators = () => {
    if (!space.collaborators) return null;

    const columns = [
      {
        title: '用户',
        dataIndex: 'userId',
        key: 'userId',
      },
      {
        title: '权限',
        dataIndex: 'permission',
        key: 'permission',
        render: (permission: string) => (
          <Select
            disabled={!canManageSpace}
            value={permission}
            onChange={(value) => handlePermissionChange(permission, value)}
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
          canManageSpace && (
            <Button 
              danger 
              onClick={() => handleRemoveCollaborator(record.userId)}
            >
              移除
            </Button>
          )
        ),
      },
    ];

    const dataSource = space.collaborators.map((collaborator) => ({
      userId: collaborator.id,
      permission: collaborator.permission,
      key: collaborator.id,
    }));

    return (
      <div style={{ marginTop: 24 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
          <h3>协作者管理</h3>
          {canManageSpace && (
            <Button 
              type="primary" 
              onClick={() => setInviteModalVisible(true)}
            >
              邀请协作者
            </Button>
          )}
        </div>
        <Table 
          columns={columns}
          dataSource={dataSource}
          pagination={false}
          size="small"
        />
        
        {/* 邀请协作者的模态框 */}
        <Modal
          title="邀请协作者"
          open={inviteModalVisible}
          onOk={handleInviteCollaborator}
          onCancel={() => {
            setInviteModalVisible(false);
            inviteForm.resetFields();
          }}
        >
          <Form form={inviteForm} layout="vertical">
            <Form.Item
              label="用户ID"
              name="userId"
              rules={[{ required: true, message: '请输入用户ID' }]}
            >
              <Input />
            </Form.Item>
            <Form.Item
              label="权限"
              name="permission"
              rules={[{ required: true, message: '请选择权限' }]}
            >
              <Select>
                <Select.Option value="edit">可编辑</Select.Option>
                <Select.Option value="view">可查看</Select.Option>
              </Select>
            </Form.Item>
          </Form>
        </Modal>
      </div>
    );
  };

  // 在 Modal 底部添加删除按钮
  const modalFooter = [
    <Button key="cancel" onClick={onClose}>
      取消
    </Button>,
    <Button
      key="submit"
      type="primary"
      loading={loading}
      onClick={handleSubmit}
    >
      保存
    </Button>,
    canManageSpace && space.id !== SPACE_CONSTANTS.PUBLIC_SPACE_ID && (
      <Button
        key="delete"
        type="primary"
        danger
        onClick={handleDeleteSpace}
      >
        删除空间
      </Button>
    ),
  ].filter(Boolean);

  return (
    <Modal
      title="空间设置"
      open={visible}
      onCancel={onClose}
      footer={modalFooter}
      width={600}
    >
      <Form
        form={form}
        layout="vertical"
        disabled={!canManageSpace}
      >
        <Form.Item
          label="空间名称"
          name="name"
          rules={[{ required: true, message: '请输入空间名称' }]}
        >
          <Input />
        </Form.Item>

        <Form.Item
          label={`最大条目数 (1-${maxItemsLimit})`}
          name="maxItems"
          rules={[
            { required: true, message: '请输入最大条目数' },
          ]}
        >
          <InputNumber
            min={1}
            max={maxItemsLimit}
            style={{ width: '100%' }}
          />
        </Form.Item>

        <Form.Item
          label={`保留天数 (1-${maxRetentionDaysLimit})`}
          name="retentionDays"
          rules={[
            { required: true, message: '请输入保留天数' },
          ]}
        >
          <InputNumber
            min={1}
            max={maxRetentionDaysLimit}
            style={{ width: '100%' }}
          />
        </Form.Item>

        {isAdmin && !isPublicSpace && (
          <Form.Item
            label="空间类型"
            name="type"
            rules={[{ required: true, message: '请选择空间类型' }]}
          >
            <Select>
              <Select.Option value="public">公共空间</Select.Option>
              <Select.Option value="private">私有空间</Select.Option>
            </Select>
          </Form.Item>
        )}
      </Form>
      {renderCollaborators()}
    </Modal>
  );
};

export default SpaceSettingsModal; 