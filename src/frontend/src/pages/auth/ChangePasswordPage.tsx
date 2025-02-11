import React from 'react';
import { Form, Input, Button, Card, message } from 'antd';
import { useNavigate } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { changePassword } from '@/api/auth';
import { clearAuth } from '@/store/slices/authSlice';

interface ChangePasswordForm {
  oldPassword: string;
  newPassword: string;
  confirmPassword: string;
}

const ChangePasswordPage: React.FC = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const [loading, setLoading] = React.useState(false);

  const handleSubmit = async (values: ChangePasswordForm) => {
    try {
      setLoading(true);
      // const response = await changePassword({
      await changePassword({
        oldPassword: values.oldPassword,
        newPassword: values.newPassword
      });
      
      message.success('密码修改成功，请重新登录');

      // 延迟跳转到登录页面，让用户能看到成功提示
      setTimeout(() => {
        // 清除登录状态
        dispatch(clearAuth());
        navigate('/login', { replace: true });
      }, 1000);
      
    } catch (error: any) {
      message.error(error.message || '修改密码失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="tw-flex tw-justify-center tw-items-center tw-min-h-screen tw-bg-gray-50">
      <Card 
        className="tw-w-full tw-max-w-md tw-shadow-lg"
      >
        <div className="tw-text-xl tw-font-semibold tw-mb-6 tw-text-center">修改密码</div>
        <Form
          form={form}
          onFinish={handleSubmit}
          layout="vertical"
        >
          <Form.Item
            label="当前密码"
            name="oldPassword"
            rules={[
              { required: true, message: '请输入当前密码' },
              { min: 6, message: '旧密码长度不能小于6位' }
            ]}
          >
            <Input.Password />
          </Form.Item>

          <Form.Item
            label="新密码"
            name="newPassword"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 8, message: '密码长度不能小于8位' }
            ]}
          >
            <Input.Password />
          </Form.Item>

          <Form.Item
            label="确认新密码"
            name="confirmPassword"
            dependencies={['newPassword']}
            rules={[
              { required: true, message: '请确认新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('newPassword') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'));
                },
              }),
            ]}
          >
            <Input.Password />
          </Form.Item>

          <Form.Item className="tw-mb-2">
            <Button 
              type="primary" 
              htmlType="submit"
              loading={loading} 
              className="tw-w-full"
            >
              确认修改
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default ChangePasswordPage; 