import { List as AntList } from 'antd';
import { ListProps as AntListProps } from 'antd/es/list';
import Empty from './Empty';
import styles from './List.module.scss';

interface ListProps<T> extends Omit<AntListProps<T>, 'locale'> {
  emptyText?: string;
  className?: string;
}

function List<T>({
  emptyText = '暂无数据',
  className,
  ...props
}: ListProps<T>) {
  return (
    <AntList<T>
      {...props}
      className={`${styles.list} ${className}`}
      locale={{ emptyText: <Empty description={emptyText} /> }}
    />
  );
}

export default List; 