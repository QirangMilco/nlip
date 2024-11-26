import React from 'react';
import { Space } from 'antd';
import styles from './Toolbar.module.scss';

interface ToolbarProps {
  left?: React.ReactNode;
  right?: React.ReactNode;
  className?: string;
}

const Toolbar: React.FC<ToolbarProps> = ({
  left,
  right,
  className
}) => {
  return (
    <div className={`${styles.container} ${className}`}>
      <Space className={styles.left}>
        {left}
      </Space>
      <Space className={styles.right}>
        {right}
      </Space>
    </div>
  );
};

export default Toolbar; 