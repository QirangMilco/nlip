import React, { useState, useEffect } from 'react';
import { Modal, Form, Input, InputNumber, message, Select } from 'antd';
import { createSpace } from '@/api/spaces';
import { getSettings } from '@/api/admin';
import { useSelector } from 'react-redux';
import { RootState } from '@/store';

const CreateSpaceModal: React.FC<{
  visible: boolean;
  onClose: () => void;
  onSuccess: () => void;
}> = ({ visible, onClose, onSuccess }) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [maxItemsLimit, setMaxItemsLimit] = useState<number>(100);
  const [maxRetentionDaysLimit, setMaxRetentionDaysLimit] = useState<number>(30);

  // 获取当前用户信息
  const isAdmin = useSelector((state: RootState) => state.auth.user?.isAdmin);

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const settings = await getSettings();
        setMaxItemsLimit(settings.max_items_limit);
        setMaxRetentionDaysLimit(settings.max_retention_days_limit);
        form.setFieldsValue({
          maxItems: settings.default_max_items || settings.max_items_limit,
          retentionDays: settings.default_retention_days || settings.max_retention_days_limit,
          type: 'private' // 默认设置为私有空间
        });
      } catch (error) {
        message.error('获取设置失败');
      }
    };

    if (visible) {
      fetchSettings();
    }
  }, [visible, form]);

  const handleSubmit = async () => {
    try {
      setLoading(true);
      const values = await form.validateFields();
      await createSpace(values);
      message.success('空间创建成功');
      form.resetFields();
      await onSuccess();
      onClose();
    } catch (error: any) {
      message.error(error.message || '创建空间失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title="创建新空间"
      open={visible}
      onOk={handleSubmit}
      onCancel={onClose}
      confirmLoading={loading}
    >
      <Form form={form} layout="vertical">
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
          rules={[{ required: true, message: '请输入最大条目数' }]}
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
          rules={[{ required: true, message: '请输入保留天数' }]}
        >
          <InputNumber
            min={1}
            max={maxRetentionDaysLimit}
            style={{ width: '100%' }}
          />
        </Form.Item>

        {isAdmin && (
          <Form.Item
            label="空间类型"
            name="type"
            rules={[{ required: true, message: '请选择空间类型' }]}
            initialValue="private"
          >
            <Select>
              <Select.Option value="public">公共空间</Select.Option>
              <Select.Option value="private">私有空间</Select.Option>
            </Select>
          </Form.Item>
        )}
      </Form>
    </Modal>
  );
};

export default CreateSpaceModal;
