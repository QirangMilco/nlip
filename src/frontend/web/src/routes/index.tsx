import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import LoginPage from '@/pages/auth/LoginPage';
import RegisterPage from '@/pages/auth/RegisterPage';
import ClipsPage from '@/pages/clips/ClipsPage';
import PrivateRoute from '@/components/auth/PrivateRoute';
import ChangePasswordPage from '@/pages/auth/ChangePasswordPage';

const AppRoutes: React.FC = () => {
  return (
    <Routes>
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
      
      {/* 默认重定向到登录页 */}
      <Route path="/" element={<Navigate to="/login" replace />} />
      <Route path="*" element={<Navigate to="/login" replace />} />
    </Routes>
  );
};

export default AppRoutes; 