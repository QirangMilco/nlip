import React from 'react';
import { Tag as AntTag } from 'antd';
import styles from './Tag.module.scss';

type TagType = 'default' | 'success' | 'warning' | 'error';

interface TagProps {
  type?: TagType;
  text: string;
  className?: string;
}

const Tag: React.FC<TagProps> = ({
  type = 'default',
  text,
  className
}) => {
  const getColor = () => {
    switch (type) {
      case 'success':
        return '#52c41a';
      case 'warning':
        return '#faad14';
      case 'error':
        return '#ff4d4f';
      default:
        return '#1890ff';
    }
  };

  return (
    <AntTag
      color={getColor()}
      className={`${styles.tag} ${className}`}
    >
      {text}
    </AntTag>
  );
};

export default Tag; 