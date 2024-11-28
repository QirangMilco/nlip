import React from 'react';
import { Modal, Form, Input, InputNumber, message } from 'antd';
import { Space } from '@/store/types';
import { updateSpace } from '@/api/spaces';

interface SpaceSettingsModalProps {
  visible: boolean;
  space: Space;
  onClose: () => void;
}

const SpaceSettingsModal: React.FC<SpaceSettingsModalProps> = ({
  visible,
  space,
  onClose,
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = React.useState(false);

  React.useEffect(() => {
    if (visible && space) {
      form.setFieldsValue({
        name: space.name,
        maxItems: space.maxItems,
        retentionDays: space.retentionDays,
      });
    }
  }, [visible, space, form]);

  const handleSubmit = async () => {
    try {
      setLoading(true);
      const values = await form.validateFields();
      await updateSpace(space.id, values);
      message.success('空间设置已更新');
      onClose();
    } catch (error: any) {
      message.error(error.message || '更新空间设置失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title="空间设置"
      open={visible}
      onCancel={onClose}
      onOk={handleSubmit}
      confirmLoading={loading}
    >
      <Form
        form={form}
        layout="vertical"
      >
        <Form.Item
          label="空间名称"
          name="name"
          rules={[{ required: true, message: '请输入空间名称' }]}
        >
          <Input />
        </Form.Item>

        <Form.Item
          label="最大条目数"
          name="maxItems"
          rules={[{ required: true, message: '请输入最大条目数' }]}
        >
          <InputNumber min={1} max={1000} style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item
          label="保留天数"
          name="retentionDays"
          rules={[{ required: true, message: '请输入保留天数' }]}
        >
          <InputNumber min={1} max={365} style={{ width: '100%' }} />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default SpaceSettingsModal; 