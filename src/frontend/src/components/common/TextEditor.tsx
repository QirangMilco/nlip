import React from 'react';
import { Input } from 'antd';
import styles from './TextEditor.module.scss';

const { TextArea } = Input;

interface TextEditorProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  maxLength?: number;
  rows?: number;
  disabled?: boolean;
  className?: string;
}

const TextEditor: React.FC<TextEditorProps> = ({
  value,
  onChange,
  placeholder = '请输入内容...',
  maxLength,
  rows = 6,
  disabled = false,
  className
}) => {
  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    onChange(e.target.value);
  };

  return (
    <div className={`${styles.container} ${className}`}>
      <TextArea
        value={value}
        onChange={handleChange}
        placeholder={placeholder}
        maxLength={maxLength}
        rows={rows}
        disabled={disabled}
        showCount={!!maxLength}
        className={styles.editor}
      />
    </div>
  );
};

export default TextEditor; 