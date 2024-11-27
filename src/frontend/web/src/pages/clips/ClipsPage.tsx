import React, { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useClips } from '@/hooks/useClip';
import { useSpace } from '@/hooks/useSpace';
import { Card, Empty, Spin, Select } from 'antd';
import ClipList from './components/ClipList';
import ClipUpload from './components/ClipUpload';

const ClipsPage: React.FC = () => {
  const navigate = useNavigate();
  const { spaceId } = useParams<{ spaceId: string }>();
  
  // 获取所有空间列表
  const { spaces, loading: loadingSpaces } = useSpace();
  
  // 获取当前空间信息
  const { space, isLoading: isLoadingSpace } = useSpace(spaceId!);
  
  // 获取剪贴板列表
  const {
    clips,
    isLoading: isLoadingClips,
    uploadClip,
    deleteClip,
    downloadClip,
  } = useClips(spaceId!);

  // 如果没有指定空间ID，默认选择第一个公共空间
  useEffect(() => {
    if (!spaceId && spaces.length > 0) {
      const defaultSpace = spaces.find(s => s.type === 'public') || spaces[0];
      navigate(`/clips/${defaultSpace.id}`);
    }
  }, [spaceId, spaces, navigate]);

  if (loadingSpaces || isLoadingSpace || isLoadingClips) {
    return <Spin />;
  }

  if (!space) {
    return <Empty description="空间不存在" />;
  }

  const handleSpaceChange = (newSpaceId: string) => {
    navigate(`/clips/${newSpaceId}`);
  };

  return (
    <Card 
      title={
        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          <span>剪贴板</span>
          <Select
            value={space.id}
            onChange={handleSpaceChange}
            style={{ width: 200 }}
            loading={loadingSpaces}
          >
            {spaces.map(s => (
              <Select.Option key={s.id} value={s.id}>
                {s.name} ({s.type === 'public' ? '公共' : '私有'})
              </Select.Option>
            ))}
          </Select>
        </div>
      }
    >
      <ClipUpload onUpload={uploadClip} />
      <ClipList
        clips={clips}
        onDelete={deleteClip}
        onDownload={downloadClip}
      />
    </Card>
  );
};

export default ClipsPage; 