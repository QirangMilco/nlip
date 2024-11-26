import React from 'react';
import { Tooltip as AntTooltip } from 'antd';
import { TooltipPlacement } from 'antd/es/tooltip';
import styles from './Tooltip.module.scss';

interface TooltipProps {
  title: React.ReactNode;
  children: React.ReactNode;
  placement?: TooltipPlacement;
  className?: string;
}

const Tooltip: React.FC<TooltipProps> = ({
  title,
  children,
  placement = 'top',
  className
}) => {
  return (
    <AntTooltip
      title={title}
      placement={placement}
      overlayClassName={`${styles.tooltip} ${className}`}
    >
      {children}
    </AntTooltip>
  );
};

export default Tooltip; 