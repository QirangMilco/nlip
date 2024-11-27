import React from 'react';
import { List, Space } from 'antd';
import { Clip } from '@/store/types';
import IconButton from '@/components/common/IconButton';
import { DeleteOutlined, DownloadOutlined } from '@ant-design/icons';

interface ClipListProps {
  clips: Clip[];
  onDelete: (clipId: string) => void;
  onDownload: (clipId: string) => void;
}

const ClipList: React.FC<ClipListProps> = ({ clips, onDelete, onDownload }) => {
  return (
    <List
      dataSource={clips}
      renderItem={(clip) => (
        <List.Item
          actions={[
            <IconButton
              key="download"
              icon={<DownloadOutlined />}
              tooltip="下载"
              onClick={() => onDownload(clip.id)}
            />,
            <IconButton
              key="delete"
              icon={<DeleteOutlined />}
              tooltip="删除"
              onClick={() => onDelete(clip.id)}
            />
          ]}
        >
          <List.Item.Meta
            title={clip.title}
            description={clip.contentType}
          />
        </List.Item>
      )}
    />
  );
};

export default ClipList; 