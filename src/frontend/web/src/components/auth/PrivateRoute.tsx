import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { RootState } from '@/store';

interface PrivateRouteProps {
  children: React.ReactNode;
}

const PrivateRoute: React.FC<PrivateRouteProps> = ({ children }) => {
  const location = useLocation();
  const { token, user, needChangePwd } = useSelector((state: RootState) => state.auth);

  if (!token || !user) {
    // 将当前路径作为 redirect 参数传递给登录页
    return <Navigate to={`/login?redirect=${encodeURIComponent(location.pathname)}`} replace />;
  }

  if (needChangePwd && !location.pathname.includes('/change-password')) {
    return <Navigate to="/change-password" replace />;
  }

  return <>{children}</>;
};

export default PrivateRoute; 