import React from 'react';
import { Select, Space } from 'antd';
import { SortAscendingOutlined } from '@ant-design/icons';
import styles from './Sort.module.scss';

export interface SortOption {
  label: string;
  value: string;
}

interface SortProps {
  options: SortOption[];
  value?: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
}

const Sort: React.FC<SortProps> = ({
  options,
  value,
  onChange,
  placeholder = '排序',
  className
}) => {
  return (
    <Space className={`${styles.sort} ${className}`}>
      <SortAscendingOutlined />
      <Select
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        options={options}
        allowClear
        className={styles.select}
      />
    </Space>
  );
};

export default Sort; 