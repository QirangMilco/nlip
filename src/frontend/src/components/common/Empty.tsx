import React from 'react';
import { Empty as AntEmpty } from 'antd';
import { EmptyProps } from 'antd/es/empty';
import styles from './Empty.module.scss';

interface CustomEmptyProps extends EmptyProps {
  fullPage?: boolean;
}

const Empty: React.FC<CustomEmptyProps> = ({ 
  description = '暂无数据',
  fullPage = false,
  ...props 
}) => {
  const content = (
    <AntEmpty
      description={description}
      {...props}
    />
  );

  if (fullPage) {
    return (
      <div className={styles.fullPage}>
        {content}
      </div>
    );
  }

  return <div className={styles.container}>{content}</div>;
};

export default Empty; 