import React from 'react';
import { Modal, Form, Input, Select, InputNumber } from 'antd';
import { Space } from '@/store/slices/spaceSlice';

interface SpaceFormProps {
  visible: boolean;
  initialValues?: Space | null;
  onCancel: () => void;
  onSubmit: (values: any) => void;
}

const SpaceForm: React.FC<SpaceFormProps> = ({
  visible,
  initialValues,
  onCancel,
  onSubmit,
}) => {
  const [form] = Form.useForm();
  const isEdit = !!initialValues;

  React.useEffect(() => {
    if (visible && initialValues) {
      form.setFieldsValue(initialValues);
    } else {
      form.resetFields();
    }
  }, [visible, initialValues, form]);

  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      onSubmit(values);
    } catch (error) {
      // 表单验证失败
    }
  };

  return (
    <Modal
      title={isEdit ? '编辑空间' : '创建空间'}
      open={visible}
      onOk={handleOk}
      onCancel={onCancel}
      afterClose={() => form.resetFields()}
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          type: 'private',
          maxItems: 20,
          retentionDays: 7,
        }}
      >
        <Form.Item
          name="name"
          label="空间名称"
          rules={[
            { required: true, message: '请输入空间名称' },
            { min: 2, message: '名称至少2个字符' },
            { max: 50, message: '名称最多50个字符' },
          ]}
        >
          <Input placeholder="请输入空间名称" />
        </Form.Item>

        <Form.Item
          name="type"
          label="空间类型"
          rules={[{ required: true, message: '请选择空间类型' }]}
        >
          <Select
            options={[
              { label: '私有空间', value: 'private' },
              { label: '公共空间', value: 'public' },
            ]}
            disabled={isEdit}
          />
        </Form.Item>

        <Form.Item
          name="maxItems"
          label="最大条目数"
          rules={[
            { required: true, message: '请输入最大条目数' },
            { type: 'number', min: 1, max: 20, message: '条目数必须在1-20之间' },
          ]}
        >
          <InputNumber min={1} max={20} style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item
          name="retentionDays"
          label="保留天数"
          rules={[
            { required: true, message: '请输入保留天数' },
            { type: 'number', min: 1, max: 7, message: '保留天数必须在1-7之间' },
          ]}
        >
          <InputNumber min={1} max={7} style={{ width: '100%' }} />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default SpaceForm; 