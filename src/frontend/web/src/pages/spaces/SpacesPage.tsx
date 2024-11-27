import React from 'react';
import { Card, Button, Table, Space, Modal, message } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { useSpace } from '@/hooks/useSpace';
import { useNavigate } from 'react-router-dom';
import SpaceForm from './components/SpaceForm';
import { Space as SpaceType } from '@/store/slices/spaceSlice';
import dayjs from 'dayjs';

const SpacesPage: React.FC = () => {
  const navigate = useNavigate();
  const { spaces, loading, error, fetchSpaces, createSpace, updateSpaceById, deleteSpaceById } = useSpace();
  const [isModalVisible, setIsModalVisible] = React.useState(false);
  const [editingSpace, setEditingSpace] = React.useState<SpaceType | null>(null);

  React.useEffect(() => {
    fetchSpaces();
  }, [fetchSpaces]);

  const handleCreate = async (values: any) => {
    try {
      await createSpace(values);
      setIsModalVisible(false);
      message.success('创建空间成功');
    } catch (error: any) {
      message.error(error.message);
    }
  };

  const handleUpdate = async (values: any) => {
    if (!editingSpace) return;
    try {
      await updateSpaceById(editingSpace.id, values);
      setIsModalVisible(false);
      setEditingSpace(null);
      message.success('更新空间成功');
    } catch (error: any) {
      message.error(error.message);
    }
  };

  const handleDelete = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '删除空间将同时删除其中的所有内容，确定要删除吗？',
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        try {
          await deleteSpaceById(id);
          message.success('删除空间成功');
        } catch (error: any) {
          message.error(error.message);
        }
      },
    });
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: SpaceType) => (
        <a onClick={() => navigate(`/clips/${record.id}`)}>{text}</a>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (text: string) => (text === 'public' ? '公共' : '私有'),
    },
    {
      title: '最大条目数',
      dataIndex: 'maxItems',
      key: 'maxItems',
    },
    {
      title: '保留天数',
      dataIndex: 'retentionDays',
      key: 'retentionDays',
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
      render: (_: any, record: SpaceType) => (
        <Space>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => {
              setEditingSpace(record);
              setIsModalVisible(true);
            }}
          >
            编辑
          </Button>
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

  return (
    <div className="page-container">
      <Card
        title="空间管理"
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              setEditingSpace(null);
              setIsModalVisible(true);
            }}
          >
            创建空间
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={spaces}
          rowKey="id"
          loading={loading}
        />
      </Card>

      <SpaceForm
        visible={isModalVisible}
        initialValues={editingSpace}
        onCancel={() => {
          setIsModalVisible(false);
          setEditingSpace(null);
        }}
        onSubmit={editingSpace ? handleUpdate : handleCreate}
      />
    </div>
  );
};

export default SpacesPage; 