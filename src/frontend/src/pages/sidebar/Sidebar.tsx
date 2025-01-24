import React from 'react';
import { 
  HomeOutlined, 
  LogoutOutlined,
  LoginOutlined,
  SettingOutlined
} from '@ant-design/icons';
import SidebarItem from './SidebarItem';

interface SidebarProps {
  username?: string;
  token?: string;
  className?: string;
  children?: React.ReactNode;
  handleLogout?: () => void;
  navigate: (to: string) => void;
}

const Sidebar: React.FC<SidebarProps> = ({
  className = '',
  children,
  username,
  token,
  handleLogout,
  navigate
}) => {

  // 基础容器样式
  const containerClasses = `
    tw-group tw-flex tw-flex-col tw-justify-start tw-items-start
    tw-fixed tw-top-0 tw-left-0 tw-select-none tw-bg-zinc-50 
    tw-border-r tw-h-full tw-transition-all tw-hover:shadow-xl
    tw-z-2 tw-w-56 tw-px-4
    ${className}
  `;

  // 如果提供了children,直接渲染
  if (children) {
    return <div className={containerClasses}>{children}</div>;
  }

  // 默认内容
  return (
    <div className={containerClasses}>
      <header className="tw-w-full tw-h-full tw-overflow-auto 
        tw-flex tw-flex-col tw-justify-start tw-items-start 
        tw-py-4 tw-md:pt-6 tw-z-30">
        
        {token && username ? (
          <>
            {/* 用户信息 */}
            <div className="tw-w-full tw-px-1 tw-shrink-0">
              <div className="tw-py-1 tw-my-1 tw-flex tw-flex-row 
                tw-items-center tw-cursor-pointer tw-rounded-2xl 
                tw-border tw-border-transparent tw-px-3">
                <span className="tw-ml-2 tw-text-lg tw-font-medium 
                  tw-text-slate-800 tw-shrink tw-truncate">
                  {username}
                </span>
              </div>
            </div>

            {/* 导航菜单 */}
            <div className="tw-w-full tw-px-1 tw-py-2 
              tw-flex tw-flex-col tw-justify-start tw-items-start 
              tw-shrink-0 tw-space-y-2">
              <SidebarItem
                icon={<HomeOutlined />}
                text="主页"
                onClick={() => navigate('/')}
                active={true}
              />
              <SidebarItem
                icon={<SettingOutlined />}
                text="设置"
                onClick={() => navigate('/settings')}
              />
              <SidebarItem
                icon={<LogoutOutlined />}
                text="退出登录"
                onClick={handleLogout}
              />
            </div>
          </>
        ) : (
          <div className="tw-w-full tw-px-1 tw-space-y-2">
            <div className="tw-py-1 tw-my-1 tw-flex tw-flex-row 
                tw-items-center tw-cursor-pointer tw-rounded-2xl 
                tw-border tw-border-transparent tw-px-3">
                <span className="tw-ml-2 tw-text-lg tw-font-medium 
                  tw-text-slate-800 tw-shrink tw-truncate">
                  游客
                </span>
              </div>
            <SidebarItem
              icon={<HomeOutlined />}
              text="主页"
              onClick={() => navigate('/')}
              active={true}
            />
            <SidebarItem
              icon={<LoginOutlined />}
              text="登录"
              onClick={() => navigate('/login')}
            />
          </div>
        )}
      </header>
    </div>
  );
};

export default Sidebar;
