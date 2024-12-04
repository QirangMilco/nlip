import React from 'react';
import { Upload, Button } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import { UploadClipRequest } from '@/store/types';

interface ClipUploadProps {
  onUpload: (data: UploadClipRequest) => Promise<void>;
  spaceId: string;
}

const ClipUpload: React.FC<ClipUploadProps> = ({ onUpload, spaceId }) => {
  const handleUpload = async (file: File) => {
    const data: UploadClipRequest = {
      file,
      contentType: file.type,
      spaceId,
    };
    await onUpload(data);
  };

  return (
    <Upload
      beforeUpload={handleUpload}
      showUploadList={false}
    >
      <Button icon={<UploadOutlined />}>上传文件</Button>
    </Upload>
  );
};

export default ClipUpload; 