import React from 'react';
import { Form, Input, Button, Card, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { Link, useNavigate, useSearchParams, useLocation } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { setAuth } from '@/store/slices/authSlice';
import { login } from '@/api/auth';
import styles from './AuthPage.module.scss';


const LoginPage: React.FC = () => {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const location = useLocation();
  // const [searchParams] = useSearchParams();
  // const redirect = searchParams.get('redirect') || '/clips';

  const onFinish = async (values: { username: string; password: string }) => {
    try {
      const response = await login(values);
      dispatch(setAuth({
        token: response.token,
        user: response.user,
        needChangePwd: response.needChangePwd
      }));
      
      // 检查是否需要修改密码
      if (response.needChangePwd) {
        navigate('/change-password');
      } else {
        // 获取重定向地址
        const params = new URLSearchParams(location.search);
        const redirect = params.get('redirect') || '/clips';
        navigate(redirect);
      }
    } catch (error: any) {
      message.error(error.message || '登录失败');
    }
  };

  return (
    <div className={styles.container}>
      <Card title="登录" className={styles.card}>
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

          <div className={styles.footer}>
            没有账号？<Link to="/register">立即注册</Link>
          </div>
        </Form>
      </Card>
    </div>
  );
};

export default LoginPage; 