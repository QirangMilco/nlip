import React from 'react';
import { Select, Space } from 'antd';
import { FilterOutlined } from '@ant-design/icons';
import styles from './Filter.module.scss';

export interface FilterOption {
  label: string;
  value: string | number;
}

interface FilterProps {
  options: FilterOption[];
  value?: string | number;
  onChange: (value: string | number) => void;
  placeholder?: string;
  className?: string;
}

const Filter: React.FC<FilterProps> = ({
  options,
  value,
  onChange,
  placeholder = '筛选',
  className
}) => {
  return (
    <Space className={`${styles.filter} ${className}`}>
      <FilterOutlined />
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

export default Filter; 