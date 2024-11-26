import React from 'react';
import { Card, Button, Table, Space, Upload, Modal, message, Select } from 'antd';
import { UploadOutlined, DeleteOutlined, DownloadOutlined, PlusOutlined } from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { useClip } from '@/hooks/useClip';
import { useSpace } from '@/hooks/useSpace';
import { UploadFile } from 'antd/es/upload/interface';
import dayjs from 'dayjs';
import styles from './ClipsPage.module.scss';

const ClipsPage: React.FC = () => {
  const { spaceId } = useParams<{ spaceId: string }>();
  const { clips, loading, fetchClips, uploadClip, deleteClipById, downloadClip } = useClip();
  const { spaces, currentSpace, fetchSpaces, selectSpace } = useSpace();
  const navigate = useNavigate();

  const [uploadModalVisible, setUploadModalVisible] = React.useState(false);
  const [textContent, setTextContent] = React.useState('');

  React.useEffect(() => {
    fetchSpaces().then(() => {
      if (spaces.length > 0) {
        const targetSpace = spaceId ? spaces.find(s => s.id === spaceId) : spaces[0];
        if (targetSpace) {
          selectSpace(targetSpace);
          if (!spaceId) {
            navigate(`/clips/${targetSpace.id}`);
          }
        }
      }
    });
  }, [spaceId, spaces, fetchSpaces, selectSpace, navigate]);

  React.useEffect(() => {
    if (spaceId) {
      fetchClips(spaceId);
    }
  }, [spaceId, fetchClips]);

  const handleSpaceChange = (newSpaceId: string) => {
    const space = spaces.find(s => s.id === newSpaceId);
    if (space) {
      selectSpace(space);
      navigate(`/clips/${newSpaceId}`);
    }
  };

  const handleUploadText = async () => {
    if (!textContent.trim()) {
      message.error('请输入内容');
      return;
    }

    try {
      await uploadClip({
        spaceId: spaceId!,
        contentType: 'text/plain',
        content: textContent,
      });
      setUploadModalVisible(false);
      setTextContent('');
      message.success('上传成功');
      fetchClips(spaceId!);
    } catch (error: any) {
      message.error(error.message);
    }
  };

  const handleUploadFile = async (file: UploadFile) => {
    try {
      await uploadClip({
        spaceId: spaceId!,
        contentType: file.type || 'application/octet-stream',
        file: file as any,
      });
      message.success('上传成功');
      fetchClips(spaceId!);
      return false; // 阻止自动上传
    } catch (error: any) {
      message.error(error.message);
      return false;
    }
  };

  const handleDelete = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这条内容吗？',
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        try {
          await deleteClipById(id);
          message.success('删除成功');
          fetchClips(spaceId!);
        } catch (error: any) {
          message.error(error.message);
        }
      },
    });
  };

  const handleDownload = async (id: string, filename: string) => {
    try {
      await downloadClip(id, filename);
    } catch (error: any) {
      message.error(error.message);
    }
  };

  const columns = [
    {
      title: '类型',
      dataIndex: 'contentType',
      key: 'contentType',
      render: (text: string) => {
        if (text.startsWith('text/')) return '文本';
        if (text.startsWith('image/')) return '图片';
        return '文件';
      },
    },
    {
      title: '内容',
      dataIndex: 'content',
      key: 'content',
      render: (text: string, record: any) => {
        if (record.filePath) {
          return <span>{record.filePath.split('/').pop()}</span>;
        }
        return <span>{text}</span>;
      },
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: any) => (
        <Space>
          {record.filePath && (
            <Button
              type="text"
              icon={<DownloadOutlined />}
              onClick={() => handleDownload(record.id, record.filePath.split('/').pop())}
            >
              下载
            </Button>
          )}
          <Button
            type="text"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(record.id)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  const spaceOptions = Array.isArray(spaces) ? spaces : [];

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <Select
          value={currentSpace?.id}
          onChange={handleSpaceChange}
          style={{ width: 200 }}
          loading={loading}
        >
          {spaceOptions.map(space => (
            <Select.Option key={space.id} value={space.id}>
              {space.name}
            </Select.Option>
          ))}
        </Select>
        <Button 
          type="primary" 
          icon={<PlusOutlined />}
          onClick={() => navigate('/spaces')}
        >
          管理空间
        </Button>
      </div>
      <Card
        title={`${currentSpace?.name || '剪贴板内容'}`}
        extra={
          <Space>
            <Button
              type="primary"
              icon={<UploadOutlined />}
              onClick={() => setUploadModalVisible(true)}
            >
              上传文本
            </Button>
            <Upload
              showUploadList={false}
              beforeUpload={handleUploadFile}
              accept=".txt,.pdf,.png,.jpg,.jpeg,.gif"
            >
              <Button icon={<UploadOutlined />}>上传文件</Button>
            </Upload>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={clips}
          rowKey="id"
          loading={loading}
        />
      </Card>

      <Modal
        title="上传文本内容"
        open={uploadModalVisible}
        onOk={handleUploadText}
        onCancel={() => {
          setUploadModalVisible(false);
          setTextContent('');
        }}
      >
        <textarea
          className={styles.textArea}
          value={textContent}
          onChange={(e) => setTextContent(e.target.value)}
          placeholder="请输入要上传的文本内容"
          rows={6}
        />
      </Modal>
    </div>
  );
};

export default ClipsPage; 