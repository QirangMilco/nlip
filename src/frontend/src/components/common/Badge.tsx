import React from 'react';
import { Badge as AntBadge } from 'antd';
import { PresetStatusColorType } from 'antd/es/_util/colors';
import styles from './Badge.module.scss';

interface BadgeProps {
  status: PresetStatusColorType;
  text: string;
  dot?: boolean;
  className?: string;
}

const Badge: React.FC<BadgeProps> = ({
  status,
  text,
  dot = false,
  className
}) => {
  return (
    <div className={`${styles.badge} ${className}`}>
      <AntBadge
        status={status}
        text={text}
        dot={dot}
      />
    </div>
  );
};

export default Badge; 