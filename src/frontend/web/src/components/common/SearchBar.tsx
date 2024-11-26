import React from 'react';
import { Input } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import { debounce } from 'lodash';
import styles from './SearchBar.module.scss';

const { Search } = Input;

interface SearchBarProps {
  onSearch: (value: string) => void;
  placeholder?: string;
  delay?: number;
  className?: string;
}

const SearchBar: React.FC<SearchBarProps> = ({
  onSearch,
  placeholder = '搜索...',
  delay = 300,
  className,
}) => {
  // 使用debounce防止频繁搜索
  const debouncedSearch = React.useMemo(
    () => debounce(onSearch, delay),
    [onSearch, delay]
  );

  React.useEffect(() => {
    return () => {
      debouncedSearch.cancel();
    };
  }, [debouncedSearch]);

  return (
    <div className={`${styles.container} ${className}`}>
      <Search
        placeholder={placeholder}
        onChange={(e) => debouncedSearch(e.target.value)}
        prefix={<SearchOutlined />}
        allowClear
      />
    </div>
  );
};

export default SearchBar; 