import React from 'react';
import { Layout, Button, Space } from 'antd';
import { useNavigate } from 'react-router-dom';
import { useSelector, useDispatch } from 'react-redux';
import { clearAuth } from '@/store/slices/authSlice';
import { RootState } from '@/store';

const Header: React.FC = () => {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const { token } = useSelector((state: RootState) => state.auth);

  const handleLogout = () => {
    dispatch(clearAuth());
    navigate('/login');
  };

  const handleLogin = () => {
    navigate('/login');
  };

  const handleChangePassword = () => {
    navigate('/change-password');
  };

  return (
    <Layout.Header>
      <div className="header-content">
        <div className="logo">NLIP</div>
        <Space>
          {token ? (
            <>
              <Button onClick={handleChangePassword}>修改密码</Button>
              <Button onClick={handleLogout}>退出登录</Button>
            </>
          ) : (
            <Button onClick={handleLogin}>登录</Button>
          )}
        </Space>
      </div>
    </Layout.Header>
  );
};

export default Header; 