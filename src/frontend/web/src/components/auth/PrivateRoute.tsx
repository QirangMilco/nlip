import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { RootState } from '@/store';
import { Spin } from 'antd';
import { useAuth } from '@/hooks/useAuth';
import { isPublicRoute } from '@/constants/routes';
import styles from '@/components/common/Loading.module.scss';

interface PrivateRouteProps {
  children: React.ReactNode;
}

const PrivateRoute: React.FC<PrivateRouteProps> = ({ children }) => {
  const location = useLocation();
  const { token, user, needChangePwd } = useSelector((state: RootState) => state.auth);
  const { isInitialCheckDone } = useAuth();

  if (!isInitialCheckDone) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh' 
      }}>
        <Spin size="large" />
        <div className={styles.loadingText}>
          验证登录状态...
        </div>
      </div>
    );
  }

  if (!token || !user) {
    if (isPublicRoute(location.pathname)) {
      return <>{children}</>;
    }
    return <Navigate to={`/login?redirect=${encodeURIComponent(location.pathname)}`} replace />;
  }

  if (needChangePwd && !location.pathname.includes('/change-password')) {
    return <Navigate to="/change-password" replace />;
  }

  return <>{children}</>;
};

export default PrivateRoute; 