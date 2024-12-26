import React from 'react';

interface SidebarItemProps {
  icon?: React.ReactNode;
  text: string;
  onClick?: () => void;
  active?: boolean;
}

const SidebarItem: React.FC<SidebarItemProps> = ({
  icon,
  text,
  onClick,
  active = false,
}) => {
  // 基础样式类
  const baseClasses = `
    tw-w-full tw-px-4 tw-py-2 tw-text-lg tw-rounded-xl
    tw-flex tw-flex-row tw-items-center
    tw-text-gray-800 hover:tw-bg-gray-100 tw-transition-all
    ${active 
      ? 'tw-bg-gray-100' 
      : ''
    }
  `;

  return (
    <button 
      className={baseClasses}
      onClick={onClick}
    >
      {icon && (
        <span className="tw-w-6 tw-h-auto tw-opacity-70 tw-shrink-0">
          {icon}
        </span>
      )}
      <span className="tw-ml-3 tw-truncate">{text}</span>
    </button>
  );
};

export default SidebarItem;
