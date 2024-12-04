import React from 'react';
import { Spin } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';
import styles from './Loading.module.scss';

interface LoadingProps {
  size?: 'small' | 'default' | 'large';
  fullScreen?: boolean;
}

const Loading: React.FC<LoadingProps> = ({ 
  size = 'default',
  fullScreen = false 
}) => {
  const antIcon = <LoadingOutlined style={{ fontSize: size === 'large' ? 40 : 24 }} spin />;

  const content = (
    <div className={styles.spinWrapper}>
      <Spin indicator={antIcon} size={size} />
      {fullScreen && <div className={styles.loadingText}>加载中...</div>}
    </div>
  );

  if (fullScreen) {
    return (
      <div className={styles.fullScreen}>
        {content}
      </div>
    );
  }

  return <div className={styles.container}>{content}</div>;
};

export default Loading; 