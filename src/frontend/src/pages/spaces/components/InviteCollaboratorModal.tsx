import React, { useState } from 'react';
import { Modal, Form, Input, Select, message, Typography } from 'antd';
import { inviteCollaborator } from '@/api/spaces';
import { CopyOutlined } from '@ant-design/icons';
import { copyToClipboard } from '@/utils/clipboard';

const { Text } = Typography;

interface InviteCollaboratorModalProps {
  spaceId: string;
  visible: boolean;
  onClose: () => void;
}

const InviteCollaboratorModal: React.FC<InviteCollaboratorModalProps> = ({
  spaceId,
  visible,
  onClose,
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [inviteLink, setInviteLink] = useState('');

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);
      
      const { inviteLink } = await inviteCollaborator(
        spaceId,
        values.email,
        values.permission
      );
      
      setInviteLink(inviteLink);
      message.success('邀请已发送');
    } catch (err: any) {
      message.error(err.message || '发送邀请失败');
    } finally {
      setLoading(false);
    }
  };

  const copyInviteLink = async () => {
    try {
      await copyToClipboard(
        inviteLink,
        () => message.success('邀请链接已复制')
      );
    } catch (err) {
      message.error('复制链接失败');
    }
  };

  return (
    <Modal
      title="邀请协作者"
      open={visible}
      onCancel={onClose}
      onOk={handleSubmit}
      confirmLoading={loading}
    >
      <Form form={form} layout="vertical">
        <Form.Item
          name="email"
          label="邮箱地址"
          rules={[
            { required: true, message: '请输入邮箱地址' },
            { type: 'email', message: '请输入有效的邮箱地址' }
          ]}
        >
          <Input placeholder="请输入被邀请人的邮箱地址" />
        </Form.Item>

        <Form.Item
          name="permission"
          label="权限级别"
          initialValue="view"
          rules={[{ required: true }]}
        >
          <Select>
            <Select.Option value="edit">可编辑</Select.Option>
            <Select.Option value="view">可查看</Select.Option>
          </Select>
        </Form.Item>

        {inviteLink && (
          <Form.Item label="邀请链接">
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <Text ellipsis style={{ flex: 1 }}>{inviteLink}</Text>
              <CopyOutlined 
                onClick={copyInviteLink}
                style={{ cursor: 'pointer', color: '#1890ff' }}
              />
            </div>
          </Form.Item>
        )}
      </Form>
    </Modal>
  );
};

export default InviteCollaboratorModal;
