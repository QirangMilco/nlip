import React, { useCallback } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { RootState } from '@/store';
import { Spin } from 'antd';
import { useAuth } from '@/hooks/useAuth';

interface PrivateRouteProps {
  children: React.ReactNode;
}

const PrivateRoute: React.FC<PrivateRouteProps> = ({ children }) => {
  const location = useLocation();
  const { token, user, needChangePwd } = useSelector((state: RootState) => state.auth);
  const { isInitialCheckDone } = useAuth();

  const isPublicRoute = useCallback((path: string) => {
    return ['/login', '/register', '/change-password'].some(route => path.includes(route));
  }, []);

  if (!isInitialCheckDone) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh' 
      }}>
        <Spin size="large" tip="验证登录状态..." />
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