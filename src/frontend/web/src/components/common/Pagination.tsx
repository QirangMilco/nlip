import React from 'react';
import { Pagination as AntPagination } from 'antd';
import styles from './Pagination.module.scss';

interface PaginationProps {
  current: number;
  total: number;
  pageSize: number;
  onChange: (page: number, pageSize: number) => void;
  className?: string;
}

const Pagination: React.FC<PaginationProps> = ({
  current,
  total,
  pageSize,
  onChange,
  className
}) => {
  return (
    <div className={`${styles.container} ${className}`}>
      <AntPagination
        current={current}
        total={total}
        pageSize={pageSize}
        onChange={onChange}
        showSizeChanger
        showQuickJumper
        showTotal={(total) => `共 ${total} 条`}
      />
    </div>
  );
};

export default Pagination; 