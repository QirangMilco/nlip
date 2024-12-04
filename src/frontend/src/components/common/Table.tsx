import { Table as AntTable } from 'antd';
import type { TableProps } from 'antd';
import styles from './Table.module.scss';

interface CustomTableProps<T> extends TableProps<T> {
  loading?: boolean;
  error?: string;
  onRetry?: () => void;
}

function Table<T extends object = any>({
  loading,
  error,
  onRetry,
  className,
  ...props
}: CustomTableProps<T>) {
  return (
    <div className={`${styles.container} ${className}`}>
      <AntTable<T>
        {...props}
        loading={loading}
        locale={{
          emptyText: error ? (
            <div className={styles.error}>
              <span>{error}</span>
              {onRetry && (
                <a onClick={onRetry} className={styles.retry}>
                  重试
                </a>
              )}
            </div>
          ) : (
            '暂无数据'
          ),
        }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条`,
          ...props.pagination,
        }}
      />
    </div>
  );
}

export default Table; 