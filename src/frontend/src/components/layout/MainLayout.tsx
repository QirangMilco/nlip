import React, { memo } from 'react';
import { Layout, Avatar, Dropdown, MenuProps, Spin } from 'antd';
import { Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '@/hooks/useAuth';
import {
  UserOutlined,
  LogoutOutlined,
  AppstoreOutlined,
} from '@ant-design/icons';
import styles from './MainLayout.module.scss';

const { Header, Content } = Layout;

const MainLayout: React.FC = memo(() => {
  const { user, logout, isInitialCheckDone } = useAuth();
  const navigate = useNavigate();
  const [isTimeout, setIsTimeout] = React.useState(false);

  React.useEffect(() => {
    if (!isInitialCheckDone) {
      const timer = setTimeout(() => {
        setIsTimeout(true);
      }, 1000);
      return () => clearTimeout(timer);
    }
  }, [isInitialCheckDone]);

  const userMenu: MenuProps['items'] = [
    {
      key: 'spaces',
      icon: <AppstoreOutlined />,
      label: '空间管理',
      onClick: () => navigate('/spaces')
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: logout
    }
  ];

  if (!isInitialCheckDone) {
    return (
      <Layout className={styles.layout}>
        <Header className={styles.header}>
          <div className={styles.logo}>Nlip</div>
        </Header>
        <Content className={styles.content}>
          <div className={styles.loadingWrapper}>
            <Spin size="large" />
            <div className={styles.loadingText}>
              {isTimeout ? '加载时间较长，请耐心等待...' : '正在加载...'}
            </div>
          </div>
        </Content>
      </Layout>
    );
  }

  return (
    <Layout className={styles.layout}>
      <Header className={styles.header}>
        <div className={styles.logo} onClick={() => navigate('/')}>
          Nlip
        </div>
        <div className={styles.userInfo}>
          <Dropdown menu={{ items: userMenu }} placement="bottomRight">
            <span className={styles.userDropdown}>
              <Avatar icon={<UserOutlined />} />
              <span className={styles.username}>{user?.username}</span>
            </span>
          </Dropdown>
        </div>
      </Header>
      <Content className={styles.content}>
        <Outlet />
      </Content>
    </Layout>
  );
});

MainLayout.displayName = 'MainLayout';

export default MainLayout; 