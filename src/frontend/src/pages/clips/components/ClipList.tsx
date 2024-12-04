import React from 'react';
import { List, Space, Typography } from 'antd';
import { Clip } from '@/store/types';
import IconButton from '@/components/common/IconButton';
import { DeleteOutlined, DownloadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';

const { Text } = Typography;

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
            title={
              <Space>
                <Text>{clip.contentType}</Text>
                <Text type="secondary">
                  {dayjs(clip.createdAt).format('YYYY-MM-DD HH:mm')}
                </Text>
              </Space>
            }
            description={clip.content}
          />
        </List.Item>
      )}
    />
  );
};

export default ClipList; 