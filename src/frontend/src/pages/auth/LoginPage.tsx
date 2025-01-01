import React from 'react';
import { Form, Input, Button, Card, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { Link} from 'react-router-dom';
import { useAuth } from '@/hooks/useAuth';


const LoginPage: React.FC = () => {
  const { login } = useAuth();

  const onFinish = async (values: { username: string; password: string }) => {
    try {
      await login(values);
    } catch (error: any) {
      message.error(error.message || '登录失败');
    }
  };

  return (
    <div className="tw-flex tw-justify-center tw-items-center tw-min-h-screen tw-bg-gray-50">
      <Card className="tw-w-full tw-max-w-md tw-shadow-lg">
        <div className="tw-text-xl tw-font-semibold tw-mb-6 tw-text-center">登录</div>
        <Form
          onFinish={onFinish}
          autoComplete="off"
          layout="vertical"
        >
          <Form.Item
            name="username"
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, message: '用户名至少3个字符' },
              { max: 50, message: '用户名最多50个字符' }
            ]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="用户名"
              size="large"
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 6, message: '密码至少6个字符' },
              { max: 50, message: '密码最多50个字符' }
            ]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="密码"
              size="large"
            />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              size="large"
              block
            >
              登录
            </Button>
          </Form.Item>

          <div className="tw-mt-4 tw-text-center">
            没有账号？<Link to="/register" className="tw-text-blue-500 hover:tw-text-blue-700">立即注册</Link>
          </div>
        </Form>
      </Card>
    </div>
  );
};

export default LoginPage; 