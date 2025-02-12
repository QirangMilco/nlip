import React, { useEffect, useMemo, useState, useCallback, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useClips } from '@/hooks/useClip';
import { useSpace } from '@/hooks/useSpace';
import { 
  Empty, Spin,
  message, Modal
} from 'antd';
import { useDispatch, useSelector } from 'react-redux';
import { clearAuth } from '@/store/slices/authSlice';
import { RootState } from '@/store';
import dayjs from 'dayjs';
import SpaceSettingsModal from '@/pages/spaces/components/SpaceSettingsModal';
import { UploadClipRequest, Clip, ImagePreviewState, SpaceStats, Collaborator } from '@/store/types';
import { updateClip } from '@/api/clips';
import CreateSpaceModal from '@/pages/spaces/components/CreateSpaceModal';
import { useSpaceNavigation } from '@/hooks/useSpaceNavigation';
import { getSpaceStats, getSpaceCollaborators } from '@/api/spaces';
import Sidebar from '../sidebar/Sidebar';
import Clipboard from './components/Clipboard';
import SpaceMenu from '@/pages/spaces/components/SpaceMenu';
import { copyToClipboard } from '@/utils/clipboard';

const ClipsPage: React.FC = () => {
  // 1. 路由相关 hooks
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const params = useParams();
  const spaceId = params.spaceId;

  // 2. 状态管理 hooks
  const currentUser = useSelector((state: RootState) => state.auth.user);
  const { spaces, loading: loadingSpaces, fetchSpaces } = useSpace();
  const {
    clips,
    isLoading: isLoadingClips,
    uploadClip: uploadClipToSpace,
    deleteClip: deleteClipFromSpace,
    downloadClip: downloadClipFromSpace,
    fetchClips
  } = useClips(spaceId || '');

  // 3. 本地状态 hooks
  const [showSettings, setShowSettings] = useState(false);
  const [newContent, setNewContent] = useState('');
  const [editingClipId, setEditingClipId] = useState<string | null>(null);
  const [expandedClips, setExpandedClips] = useState<Set<string>>(new Set());
  const [editContent, setEditContent] = useState<string>('');

  // 添加图片预览状态管理
  const [imagePreviewStates, setImagePreviewStates] = useState<Record<string, ImagePreviewState>>({});

  // 添加用于显示大图的状态
  const [visibleImage, setVisibleImage] = useState<string | null>(null);

  // 修改状态定义
  const [scale, setScale] = useState(1);
  const [position, setPosition] = useState({ x: 0, y: 0 });

  // 添加拖动相关状态
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const lastPosition = useRef(position);

  // 4. 计算属性 hooks
  const currentSpace = useMemo(() => {
    return spaces.find(s => s.id === spaceId)
    },
    [spaces, spaceId]
  );

  // 添加用户认证状态检查
  const { token } = useSelector((state: RootState) => state.auth);

  // 更新空间管理权限检查
  const canManageSpace = useMemo(() => {
    if (!currentSpace || !currentUser) return false;
    if (currentSpace.type === 'public') return currentUser.isAdmin;
    return currentSpace.ownerId === currentUser.id;
  }, [currentSpace, currentUser]);

  const sortedClips = useMemo(() => {
    if (!clips) return [];
    return [...clips].sort((a, b) => 
      dayjs(b.updatedAt).valueOf() - dayjs(a.updatedAt).valueOf()
    );
  }, [clips]);

  // 5. 副作用 hooks
  useEffect(() => {
    if (!loadingSpaces && spaces.length > 0) {
      if (!spaceId || !spaces.some(s => s.id === spaceId)) {
        // 优先选择用户的私有空间
        const privateSpace = spaces.find(s => s.type === 'private');
        // 如果没有私有空间，则选择第一个公共空间
        const defaultSpace = privateSpace || spaces.find(s => s.type === 'public') || spaces[0];
        navigate(`/clips/${defaultSpace.id}`, { replace: true });
      }
    }
  }, [spaceId, spaces, loadingSpaces, navigate]);

  useEffect(() => {
    if (spaceId) {
      fetchClips();
    }
  }, [spaceId, fetchClips]);

  const [collaborators, setCollaborators] = useState<Collaborator[]>([]);
  const [loadingCollaborators, setLoadingCollaborators] = useState(false);

  const fetchCollaborators = useCallback(async () => {
    if (!spaceId) return;
    if (loadingSpaces) return;
    if (currentSpace?.type === 'public') return;
    try {
      setLoadingCollaborators(true);
      const collabs = await getSpaceCollaborators(spaceId);
      setCollaborators(collabs);
    } catch (error) {
      console.error('获取协作者信息失败:', error);
    } finally {
      setLoadingCollaborators(false);
    }
  }, [spaceId, spaces, loadingSpaces]);

  useEffect(() => {
    if (spaceId) {
      fetchSpaceStats();
      fetchCollaborators();
    }
  }, [spaceId, fetchCollaborators, currentSpace?.type]);

  // 6. 事件处理函数
  const { handleSpaceChange: navigateToSpace } = useSpaceNavigation();

  const handleSpaceChange = useCallback(async (newSpaceId: string) => {
    try {
      // 清空当前状态
      setNewContent('');
      setEditingClipId(null);
      setExpandedClips(new Set());
      setImagePreviewStates({});
      
      // 获取目标空间的类型
      const targetSpace = spaces.find(s => s.id === newSpaceId);
      const targetSpaceType = targetSpace?.type || 'private';
      
      navigateToSpace(newSpaceId, targetSpaceType, spaces);
      
    } catch (error) {
      message.error('切换空间失败');
    }
  }, [spaces, navigateToSpace]);

  const handleUpload = async (data: UploadClipRequest) => {
    if (!spaceId) return;
    try {
      await uploadClipToSpace(data);
      message.success('上传成功');
      await fetchSpaceStats();
    } catch (error: any) {
      message.error(error.message || '上传失败');
    }
  };

  // 修改下载和预览处理函数
  const handleDownload = async (clip: Clip, type: 'download' | 'preview' = 'download') => {
    if (!spaceId || !clip.filePath) {
      message.error('无效的文件');
      return null;
    }

    try {
      if (type === 'preview') {
        // 更新加载状态
        setImagePreviewStates(prev => ({
          ...prev,
          [clip.clipId]: {
            loading: true,
            error: false,
            url: null,
            scale: 1,
            position: { x: 0, y: 0 }
          }
        }));
      }

      // 对于预览操作，使用不同的 API 路径
      const blob = await downloadClipFromSpace(spaceId, clip.clipId);

      const url = window.URL.createObjectURL(blob);

      if (type === 'preview') {
        setImagePreviewStates(prev => ({
          ...prev,
          [clip.clipId]: {
            loading: false,
            error: false,
            url,
            scale: 1,
            position: { x: 0, y: 0 }
          }
        }));
        return url;
      }

      // 下载逻辑
      const fileName = extractFileName(clip.filePath);
      const a = document.createElement('a');
      a.href = url;
      a.download = fileName;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
      message.success('下载成功');
      return null;
    } catch (error: any) {
      console.error(`${type === 'download' ? '下载' : '预览'}失败:`, error);
      if (type === 'preview') {
        setImagePreviewStates(prev => ({
          ...prev,
          [clip.clipId]: {
            loading: false,
            error: true,
            url: null,
            scale: 1,
            position: { x: 0, y: 0 }
          }
        }));
      }
      message.error(error.message || `${type === 'download' ? '下载' : '预'}失败`);
      return null;
    }
  };

  const handleLogout = () => {
    dispatch(clearAuth());
    navigate('/clips/public-space', { replace: true });
  };

  const handleCopy = async (clip: Clip) => {
    if (!clip.content) {
      message.error('没有可复制的内容');
      return;
    }
    try {
      await copyToClipboard(
        clip.content,
        () => message.success('复制成功')
      );
    } catch (err) {
      message.error('复制失败');
    }
  };

  const handleSaveNew = async () => {
    if (!newContent.trim()) {
      message.warning('内容不能为空');
      return;
    }
    try {
      const data: UploadClipRequest = {
        content: newContent,
        contentType: 'text/plain',
        spaceId: spaceId || ''
      };
      await uploadClipToSpace(data);
      setNewContent('');
      message.success('保存成功');
      await fetchClips();
      await fetchSpaceStats();
    } catch (error: any) {
      message.error(error.message || '保存失败');
    }
  };

  const handleEdit = (clip: Clip) => {
    setEditingClipId(clip.clipId);
    setEditContent(clip.content || '');
  };

  const handleSaveEdit = async (clipId: string) => {
    try {
      await updateClip(spaceId || '', clipId, editContent, currentSpace?.type || 'private');
      setEditingClipId(null);
      message.success('更新成功');
      await fetchClips();
    } catch (error: any) {
      message.error(error.message || '更新失败');
    }
  };

  const toggleExpand = (clipId: string) => {
    const newExpanded = new Set(expandedClips);
    if (expandedClips.has(clipId)) {
      newExpanded.delete(clipId);
    } else {
      newExpanded.add(clipId);
    }
    setExpandedClips(newExpanded);
  };

  const handleDelete = async (clipId: string) => {
    try {
      await deleteClipFromSpace(clipId, currentSpace?.type || 'private');
      message.success('删除成功');
      await fetchClips();
      await fetchSpaceStats();
    } catch (error: any) {
      message.error(error.message || '删除失败');
    }
  };

  // 修改文件上传处理
  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file || !spaceId) return;

    // 生成随机文件名并保留原始扩展名
    const fileExt = file.name.includes('.') 
      ? `.${file.name.split('.').pop()}`
      : '';
    const randomName = `clip-${Math.random().toString(36).substring(2, 10)}${fileExt}`;

    // 创建新的 File 对象，使用随机生成的文件名
    const renamedFile = new File([file], randomName, { type: file.type });

    const data: UploadClipRequest = {
      file: renamedFile,
      contentType: file.type,
      spaceId: spaceId
    };
    handleUpload(data);
  };

  // 添加文件名提取工具函数
  function extractFileName(filePath: string): string {
    // 使用 split 方法将路径按路径分隔符分割成数组
    const parts = filePath.split(/[\\/]/);
    
    // 取数组的最后一个元素，即文件名
    const fileName = parts.pop();
    
    // 返回文件名
    return fileName || '';
  }

  function canEditClip(clip: Clip) {
    if (!currentUser || !currentSpace) return false;
    if (currentSpace.type === 'public') {
      if (currentUser.isAdmin) {
        return true;
      }
      if (!clip.creator)
        return false;
      return clip.creator.id === currentUser.id;
    }
    return currentSpace.ownerId === currentUser.id || currentSpace.permission === 'edit';
  }

  // 在组件卸载时清理所有预览URL
  useEffect(() => {
    return () => {
      Object.entries(imagePreviewStates).forEach(([_, state]) => {
        if (state.url) {
          window.URL.revokeObjectURL(state.url);
        }
      });
    };
  }, [imagePreviewStates]);

  // 在首次渲染时加载图片预览
  useEffect(() => {
    sortedClips.forEach(clip => {
      if (clip.contentType.startsWith('image/') && !imagePreviewStates[clip.clipId]) {
        handleDownload(clip, 'preview');
      }
    });
  }, [sortedClips]);

  // 添加滚轮缩放处理函数
  const imageRef = useRef<HTMLImageElement>(null);

  const handleWheel = (e: React.WheelEvent<HTMLDivElement>) => {
    e.preventDefault();
    
    const img = imageRef.current;
    if (!img) return;

    // 计算新的缩放比例
    const delta = e.deltaY * -0.001;
    const newScale = Math.min(Math.max(scale + delta * scale, 0.1), 5);
    
    // 获取图片的边界信息
    const imgRect = img.getBoundingClientRect();
    
    // 计算鼠标相对于图片的位置（考虑当前缩放和位置）
    const mouseX = e.clientX - imgRect.left;
    const mouseY = e.clientY - imgRect.top;
    
    // 计算鼠标在图片上的相对位置（0-1范围）
    const relativeX = mouseX / imgRect.width;
    const relativeY = mouseY / imgRect.height;
    
    // 计算新位置，保持鼠标指向的图片点不变
    const newPosition = {
      x: position.x - (newScale - scale) * img.width * relativeX,
      y: position.y - (newScale - scale) * img.height * relativeY
    };
    
    setScale(newScale);
    setPosition(newPosition);
  };

  // 添加鼠标事件处理函数
  const handleMouseDown = (e: React.MouseEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragging(true);
    setDragStart({ x: e.clientX, y: e.clientY });
    lastPosition.current = position;
  };

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!isDragging) return;
    
    const dx = e.clientX - dragStart.x;
    const dy = e.clientY - dragStart.y;
    
    setPosition({
      x: lastPosition.current.x + dx,
      y: lastPosition.current.y + dy
    });
  };

  const handleMouseUp = () => {
    setIsDragging(false);
  };

  const [spaceStats, setSpaceStats] = useState<SpaceStats | null>(null);
  const [loadingStats, setLoadingStats] = useState(false);

  const fetchSpaceStats = useCallback(async () => {
    if (!spaceId) return;
    try {
      setLoadingStats(true);
      const stats = await getSpaceStats(spaceId);
      setSpaceStats(stats);
    } catch (error) {
      console.error('获取空间统计信息失败:', error);
    } finally {
      setLoadingStats(false);
    }
  }, [spaceId]);

  const [showCreateSpace, setShowCreateSpace] = useState(false);

  // 7. 条件渲染
  if (loadingSpaces || isLoadingClips) {
    return (
      <div className="tw-flex tw-flex-col tw-items-center tw-justify-center tw-min-h-screen">
        <Spin />
        <div className="tw-mt-4 tw-text-gray-600">
          {loadingSpaces ? '加载空间中...' : '加载剪贴板中...'}
        </div>
      </div>
    );
  }

  if (spaces.length === 0) {
    return <Empty description="暂无可用空间" />;
  }

  if (!currentSpace) {
    return <Empty description="空间不存在" />;
  }

  // 8. 主要渲染
  return (
    <div className="tw-flex tw-flex-row tw-w-full tw-transition-all tw-mx-auto 
      tw-min-h-screen tw-justify-center tw-items-start tw-pl-56 tw-bg-gray-50">
      {/* 左侧导航栏 - 改为最小宽度 */}
      <Sidebar 
        navigate={navigate}
        username={currentUser?.username}
        token={token || ''}
        handleLogout={handleLogout}
      />

      {/* 主要区域 */}
      <main className="tw-w-full tw-h-auto tw-flex-grow tw-shrink tw-flex 
        tw-flex-col tw-justify-start tw-items-center">
          <div className="tw-container tw-w-full tw-max-w-5xl tw-min-h-full 
            tw-flex tw-flex-col tw-justify-start tw-items-center tw-pt-3 tw-pb-8"
          >
            <div className="tw-w-full tw-flex tw-flex-row tw-justify-start 
              tw-items-start tw-px-4 tw-gap-4">
              <Clipboard
                currentSpace={currentSpace}
                currentUser={currentUser}
                newContent={newContent}
                onNewContentChange={(content) => setNewContent(content)}
                onSaveNew={handleSaveNew}
                onFileUpload={handleFileUpload}
                sortedClips={sortedClips}
                clipItemProps={{
                  editingClipId,
                  editContent,
                  expandedClips,
                  imagePreviewStates,
                  canEditClip,
                  onEdit: handleEdit,
                  onSaveEdit: handleSaveEdit,
                  onCancelEdit: () => setEditingClipId(null),
                  onDelete: handleDelete,
                  onDownload: handleDownload,
                  onCopy: handleCopy,
                  onToggleExpand: toggleExpand,
                  onEditContentChange: (content) => setEditContent(content),
                  onImageClick: (url) => setVisibleImage(url),
                  extractFileName
                }}
              />

              <SpaceMenu 
                spaces={spaces}
                currentSpace={currentSpace}
                currentUser={currentUser}
                token={token || ''}
                spaceId={spaceId}
                loadingSpaces={loadingSpaces}
                loadingStats={loadingStats}
                loadingCollaborators={loadingCollaborators}
                spaceStats={spaceStats}
                collaborators={collaborators}
                canManageSpace={canManageSpace}
                onSpaceChange={handleSpaceChange}
                onCreateSpace={() => setShowCreateSpace(true)}
                onOpenSettings={() => setShowSettings(true)}
                onSpaceUpdate={async () => {
                  await Promise.all([
                    fetchSpaces(),
                    fetchSpaceStats(),
                    fetchCollaborators()
                  ]);
                }}
              />
            </div>
          </div>
      </main>

      {/* 图片预览模态框 */}
      <Modal
        open={!!visibleImage}
        footer={null}
        onCancel={() => {
          setVisibleImage(null);
          setScale(1);
          setPosition({ x: 0, y: 0 });
        }}
        width="90vw"
        style={{ maxWidth: '90vw', maxHeight: '90vh', padding: 0 }}
        centered
        closable={true}
        maskClosable={true}
      >
        <div 
          className="tw-relative tw-w-full tw-h-[80vh] tw-overflow-hidden tw-bg-gray-900"
          onWheel={handleWheel}
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          style={{ cursor: isDragging ? 'grabbing' : 'grab' }}
        >
          <img 
            ref={imageRef}
            src={visibleImage || ''} 
            alt="Preview" 
            className="tw-absolute tw-transition-transform"
            style={{ 
              transform: `translate(${position.x}px, ${position.y}px) scale(${scale})`,
              transformOrigin: '0 0',
              pointerEvents: 'none'
            }}
          />
        </div>
      </Modal>

      <CreateSpaceModal
        visible={showCreateSpace}
        onClose={() => setShowCreateSpace(false)}
        onSuccess={fetchSpaces}
      />

      {/* 其他模态框保持不变 */}
      {canManageSpace && showSettings && (
        <SpaceSettingsModal
          visible={showSettings}
          space={currentSpace}
          onClose={() => setShowSettings(false)}
          onSpaceUpdated={async (action?: 'delete') => {
            try {
              if (action === 'delete') {
                setShowSettings(false);
                await fetchSpaces();
                
                if (spaces.length > 0) {
                  const privateSpace = spaces.find(s => s.type === 'private');
                  const defaultSpace = privateSpace || spaces[0];
                  navigate(`/clips/${defaultSpace.id}`, { replace: true });
                } else {
                  navigate('/clips', { replace: true });
                }
              } else {
                await Promise.all([
                  fetchSpaces(),
                  fetchSpaceStats(),
                  fetchClips()
                ]);
                setShowSettings(false);
              }
            } catch (error) {
              console.error('更新空间数据失败:', error);
              message.error('更新空间数据失败，请刷新页面重试');
            }
          }}
        />
      )}
    </div>
  );
};

export default ClipsPage;