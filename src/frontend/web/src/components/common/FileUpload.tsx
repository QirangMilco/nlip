import React from 'react';
import { Upload, Button, message } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import type { UploadProps } from 'antd';
import styles from './FileUpload.module.scss';

interface FileUploadProps {
  onUpload: (file: File) => Promise<void>;
  accept?: string;
  maxSize?: number;
  buttonText?: string;
  disabled?: boolean;
}

const FileUpload: React.FC<FileUploadProps> = ({
  onUpload,
  accept = '*',
  maxSize = 10 * 1024 * 1024, // 默认10MB
  buttonText = '上传文件',
  disabled = false,
}) => {
  const uploadProps: UploadProps = {
    beforeUpload: async (file) => {
      // 检查文件大小
      if (file.size > maxSize) {
        message.error('文件大小超过限制');
        return false;
      }

      try {
        await onUpload(file);
        return false; // 阻止默认上传行为
      } catch (error: any) {
        message.error(error.message || '上传失败');
        return false;
      }
    },
    showUploadList: false,
  };

  return (
    <div className={styles.container}>
      <Upload {...uploadProps} accept={accept}>
        <Button 
          icon={<UploadOutlined />} 
          disabled={disabled}
        >
          {buttonText}
        </Button>
      </Upload>
    </div>
  );
};

export default FileUpload; 