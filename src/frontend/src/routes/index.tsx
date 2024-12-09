import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import LoginPage from '@/pages/auth/LoginPage';
import RegisterPage from '@/pages/auth/RegisterPage';
import ClipsPage from '@/pages/clips/ClipsPage';
import PrivateRoute from '@/components/auth/PrivateRoute';
import ChangePasswordPage from '@/pages/auth/ChangePasswordPage';
import InviteConfirmation from '@/pages/auth/InviteConfirmation';

const AppRoutes: React.FC = () => {
  return (
    <Routes>
      {/* 默认重定向到公共空间 */}
      <Route path="/" element={<Navigate to="/clips/public-space" replace />} />
      
      {/* 公共路由 */}
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      
      {/* 需要认证的路由 */}
      <Route
        path="/clips"
        element={
          <PrivateRoute>
            <ClipsPage />
          </PrivateRoute>
        }
      />
      <Route
        path="/clips/:spaceId"
        element={
          <PrivateRoute>
            <ClipsPage />
          </PrivateRoute>
        }
      />
      
      {/* 修改密码路由 */}
      <Route
        path="/change-password"
        element={
          <PrivateRoute>
            <ChangePasswordPage />
          </PrivateRoute>
        }
      />

      {/* 邀请确认页面不需要登录就能访问 */}
      <Route
        path="/spaces/invite/:token"
        element={
          <PrivateRoute>
            <InviteConfirmation />
          </PrivateRoute>
        }
      />
      
      {/* 未匹配路由重定向到公共空间 */}
      <Route path="*" element={<Navigate to="/spaces/public-space" replace />} />
    </Routes>
  );
};

export default AppRoutes; 