import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import MainLayout from '@/components/layout/MainLayout';
import LoginPage from '@/pages/auth/LoginPage';
import RegisterPage from '@/pages/auth/RegisterPage';
import SpacesPage from '@/pages/spaces/SpacesPage';
import ClipsPage from '@/pages/clips/ClipsPage';

// 临时的欢迎页面组件
const WelcomePage: React.FC = () => (
  <div style={{ padding: '24px' }}>
    <h1>欢迎使用 Nlip</h1>
    <p>轻量级网络剪贴板</p>
  </div>
);

const AppRoutes: React.FC = () => {
  return (
    <Routes>
      {/* 公共路由 */}
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />

      {/* 主布局路由 */}
      <Route path="/" element={<MainLayout />}>
        <Route index element={<Navigate to="/clips" replace />} />
        <Route path="clips/:spaceId?" element={<ClipsPage />} />
        <Route path="spaces" element={<SpacesPage />} />
      </Route>

      {/* 404页面 */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
};

export default AppRoutes; 