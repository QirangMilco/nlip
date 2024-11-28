import React, { useEffect, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useClips } from '@/hooks/useClip';
import { useSpace } from '@/hooks/useSpace';
import { Card, Empty, Spin, Select, Button, Space as AntSpace, message } from 'antd';
import { LogoutOutlined, SettingOutlined } from '@ant-design/icons';
import { useDispatch, useSelector } from 'react-redux';
import { clearAuth } from '@/store/slices/authSlice';
import { RootState } from '@/store';
import ClipList from './components/ClipList';
import ClipUpload from './components/ClipUpload';
import SpaceSettingsModal from './components/SpaceSettingsModal';
import { UploadClipRequest } from '@/store/types';

const ClipsPage: React.FC = () => {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const params = useParams();
  const spaceId = params.spaceId;
  
  // 获取当前用户信息
  const currentUser = useSelector((state: RootState) => state.auth.user);
  
  // 获取所有空间列表
  const { spaces, loading: loadingSpaces } = useSpace();
  
  // 获取当前选中的空间
  const currentSpace = useMemo(() => 
    spaces.find(s => s.id === spaceId),
    [spaces, spaceId]
  );

  // 检查当前用户是否可以管理该空间
  const canManageSpace = useMemo(() => {
    if (!currentUser || !currentSpace) return false;

    // 如果是管理员
    if (currentUser.isAdmin) {
      return currentSpace.ownerId === 'system' || currentSpace.ownerId === currentUser.id;
    }
    
    // 非管理员只能管理自己的空间
    return currentSpace.ownerId === currentUser.id;
  }, [currentUser, currentSpace]);

  const [showSettings, setShowSettings] = React.useState(false);

  // 只有在有效的 spaceId 时才获取剪贴板列表
  const {
    clips,
    isLoading: isLoadingClips,
    uploadClip,
    deleteClip,
    downloadClip,
    fetchClips
  } = useClips(spaceId || '');

  // 选择默认空间并跳转
  useEffect(() => {
    if (!loadingSpaces && spaces.length > 0) {
      if (!spaceId || !spaces.some(s => s.id === spaceId)) {
        // 优先选择用户的私有空间
        const privateSpace = spaces.find(s => s.type === 'private');
        // 如果没有私有空间，则选择第一个公共空间
        const defaultSpace = privateSpace || spaces.find(s => s.type === 'public') || spaces[0];
        navigate(`/clips/${defaultSpace.id}`, { replace: true });
        return;
      }
    }
  }, [spaceId, spaces, loadingSpaces, navigate]);

  // 当 spaceId 变化时重新获取剪贴板列表
  useEffect(() => {
    if (spaceId) {
      fetchClips();
    }
  }, [spaceId, fetchClips]);

  if (loadingSpaces) {
    return <Spin />;
  }

  // 如果没有空间，显示空状态
  if (spaces.length === 0) {
    return <Empty description="暂无可用空间" />;
  }

  if (!currentSpace) {
    return <Empty description="空间不存在" />;
  }

  const handleSpaceChange = (newSpaceId: string) => {
    navigate(`/clips/${newSpaceId}`);
  };

  // 处理上传回调
  const handleUpload = async (data: UploadClipRequest) => {
    if (!spaceId) return;
    
    try {
        await uploadClip(data);
    } catch (error: any) {
        message.error(error.message || '上传失败');
    }
  };

  const handleDownload = async (clipId: string) => {
    if (!spaceId) return;
    try {
      const blob = await downloadClip(spaceId, clipId);
      // 创建下载链接
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `clip-${clipId}`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (error: any) {
      message.error(error.message || '下载失败');
    }
  };

  const handleLogout = () => {
    dispatch(clearAuth());
    navigate('/login', { replace: true });
  };

  return (
    <Card 
      title={
        <div style={{ display: 'flex', alignItems: 'center', gap: '16px', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
            <span>剪贴板 - {currentSpace.name}</span>
            <Select
              value={currentSpace.id}
              onChange={handleSpaceChange}
              style={{ width: 200 }}
              loading={loadingSpaces}
            >
              {spaces.map(s => (
                <Select.Option key={s.id} value={s.id}>
                  {s.name}
                  {s.ownerId === currentUser?.id && ' (我的)'}
                  {s.ownerId === 'system' && ' (公共)'}
                </Select.Option>
              ))}
            </Select>
          </div>
          <AntSpace>
            {canManageSpace && (
              <Button 
                icon={<SettingOutlined />}
                onClick={() => setShowSettings(true)}
              >
                空间设置
              </Button>
            )}
            <Button 
              icon={<LogoutOutlined />} 
              onClick={handleLogout}
              danger
            >
              退出登录
            </Button>
          </AntSpace>
        </div>
      }
    >
      {isLoadingClips ? (
        <Spin />
      ) : (
        <>
          <ClipUpload onUpload={handleUpload} />
          <ClipList
            clips={clips}
            onDelete={deleteClip}
            onDownload={handleDownload}
          />
        </>
      )}

      {canManageSpace && showSettings && (
        <SpaceSettingsModal
          visible={showSettings}
          space={currentSpace}
          onClose={() => setShowSettings(false)}
        />
      )}
    </Card>
  );
};

export default ClipsPage; 