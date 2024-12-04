import React, { useState, useEffect } from 'react';
import { Modal, Form, Input, InputNumber, message } from 'antd';
import { Space } from '@/store/types';
import { updateSpace } from '@/api/spaces';
import { getSettings } from '@/api/admin';

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
  const [loading, setLoading] = useState(false);
  const [maxItemsLimit, setMaxItemsLimit] = useState<number>(100);
  const [maxRetentionDaysLimit, setMaxRetentionDaysLimit] = useState<number>(30);

  useEffect(() => {
    if (visible && space) {
      form.setFieldsValue({
        name: space.name,
        maxItems: space.maxItems,
        retentionDays: space.retentionDays,
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
      </Form>
    </Modal>
  );
};

export default SpaceSettingsModal; 