import React, { useEffect, useMemo, useState, useCallback, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useClips } from '@/hooks/useClip';
import { useSpace } from '@/hooks/useSpace';
import { 
  Card, Empty, Spin, Select, Button, Space as AntSpace, 
  Input, message, Typography, Tooltip, Modal
} from 'antd';
import { 
  LogoutOutlined, SettingOutlined, CopyOutlined,
  EditOutlined, DeleteOutlined, UploadOutlined,
  SaveOutlined, DownloadOutlined,
  LoadingOutlined, ReloadOutlined
} from '@ant-design/icons';
import { useDispatch, useSelector } from 'react-redux';
import { clearAuth } from '@/store/slices/authSlice';
import { RootState } from '@/store';
import dayjs from 'dayjs';
import SpaceSettingsModal from './components/SpaceSettingsModal';
import { UploadClipRequest, Clip, ImagePreviewState } from '@/store/types';
import styles from './ClipsPage.module.scss';
import { updateClip } from '@/api/clips';
import { SPACE_CONSTANTS } from '@/constants/spaces';

const { TextArea } = Input;
const { Text } = Typography;

const ClipsPage: React.FC = () => {
  // 1. 路由相关 hooks
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const params = useParams();
  const spaceId = params.spaceId;

  // 2. 状态管理 hooks
  const currentUser = useSelector((state: RootState) => state.auth.user);
  const { spaces, loading: loadingSpaces } = useSpace();
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
  const currentSpace = useMemo(() => 
    spaces.find(s => s.id === spaceId),
    [spaces, spaceId]
  );

  // 添加用户认证状态检查
  const { token } = useSelector((state: RootState) => state.auth);
  const isPublicSpace = spaceId === SPACE_CONSTANTS.PUBLIC_SPACE_ID;
  const isGuest = isPublicSpace && !token;

  // 修改权限检查逻辑
  const canManageClip = (clip: Clip) => {
    if (isGuest) return false;
    if (!currentUser) return false;
    if (currentUser.isAdmin) return true;
    return clip.creator?.id === currentUser.id;
  };

  // 修改空间管理权限检查
  const canManageSpace = useMemo(() => {
    if (isGuest) return false;
    if (!currentUser || !currentSpace) return false;
    if (currentUser.isAdmin) {
      return currentSpace.ownerId === 'system' || currentSpace.ownerId === currentUser.id;
    }
    return currentSpace.ownerId === currentUser.id;
  }, [currentUser, currentSpace, isGuest]);

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

  // 6. 事件处理函数
  const handleSpaceChange = (newSpaceId: string) => {
    navigate(`/clips/${newSpaceId}`);
  };

  const handleUpload = async (data: UploadClipRequest) => {
    if (!spaceId) return;
    try {
      await uploadClipToSpace(data);
      message.success('上传成功');
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
            url: null
          }
        }));
      }

      // 对于预览操作，使用不同的 API 路径
      const blob = type === 'preview' && isGuest
        ? await downloadClipFromSpace(SPACE_CONSTANTS.PUBLIC_SPACE_ID, clip.clipId)
        : await downloadClipFromSpace(spaceId, clip.clipId);

      const url = window.URL.createObjectURL(blob);

      if (type === 'preview') {
        setImagePreviewStates(prev => ({
          ...prev,
          [clip.clipId]: {
            loading: false,
            error: false,
            url
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
            url: null
          }
        }));
      }
      message.error(error.message || `${type === 'download' ? '下载' : '预览'}失败`);
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
      await navigator.clipboard.writeText(clip.content);
      message.success('复制成功');
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
      await updateClip(spaceId || '', clipId, editContent);
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
      await deleteClipFromSpace(clipId);
      message.success('删除成功');
      await fetchClips();
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

  // 修改剪贴板操作按钮渲染
  const renderClipActions = (clip: Clip) => (
    <div className={styles.clipActions}>
      <Tooltip title="复制文本">
        <Button 
          type="text" 
          icon={<CopyOutlined />}
          onClick={() => handleCopy(clip)}
          disabled={!clip.content}
        />
      </Tooltip>
      
      {clip.filePath && (
        <Tooltip title="下载文件">
          <Button 
            type="text" 
            icon={<DownloadOutlined />}
            onClick={() => handleDownload(clip)}
          />
        </Tooltip>
      )}
      
      {/* 只有非游客且有权限的用户才能看到编辑按钮 */}
      {canManageClip(clip) && !clip.filePath && (
        <Tooltip title="修改">
          <Button 
            type="text" 
            icon={<EditOutlined />}
            onClick={() => handleEdit(clip)}
          />
        </Tooltip>
      )}
      
      {/* 只有非游客且有权限的用户才能看见删除按钮 */}
      {canManageClip(clip) && (
        <Tooltip title="删除">
          <Button 
            type="text" 
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(clip.clipId)}
          />
        </Tooltip>
      )}
    </div>
  );

  // 修改标题栏渲染
  const renderHeader = () => (
    <div style={{ display: 'flex', alignItems: 'center', gap: '16px', justifyContent: 'space-between' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
        <span>剪贴板 - {currentSpace?.name}</span>
        <Select
          value={currentSpace?.id}
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
        {currentUser ? (
          <span className={styles.username}>{currentUser.username}</span>
        ) : (
          <span className={styles.guest}>游客</span>
        )}
        {canManageSpace && (
          <Button 
            icon={<SettingOutlined />}
            onClick={() => setShowSettings(true)}
          >
            空间设置
          </Button>
        )}
        {token ? (
          <Button 
            icon={<LogoutOutlined />} 
            onClick={handleLogout}
            danger
          >
            退出登录
          </Button>
        ) : (
          <Button 
            type="primary"
            onClick={() => navigate('/login')}
          >
            登录
          </Button>
        )}
      </AntSpace>
    </div>
  );

  // 修改渲染剪贴板内容的部分
  const renderClipContent = (clip: Clip) => {
    if (clip.filePath) {
      const fileName = decodeURIComponent(extractFileName(clip.filePath));
      const previewState = imagePreviewStates[clip.clipId] || {
        loading: false,
        error: false,
        url: null
      };

      return (
        <div className={styles.fileContent}>
          <span className={styles.fileName}>{fileName}</span>
          {clip.contentType.startsWith('image/') && (
            <div 
              className={styles.imageContainer}
              onClick={() => setVisibleImage(previewState.url)}
            >
              {previewState.loading && (
                <div className={styles.imageLoading}>
                  <Spin indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />} />
                </div>
              )}
              
              {previewState.error && (
                <div className={styles.imageError}>
                  <span>图片加载失败</span>
                  <Button 
                    type="link" 
                    icon={<ReloadOutlined />}
                    className={styles.retryButton}
                    onClick={() => handleDownload(clip, 'preview')}
                  >
                    重试
                  </Button>
                </div>
              )}
              
              {!previewState.loading && !previewState.error && previewState.url && (
                <img 
                  src={previewState.url}
                  alt={fileName}
                  className={`${styles.imagePreview} ${styles.loaded}`}
                  onError={() => {
                    cleanupPreviewUrl(clip.clipId);
                    setImagePreviewStates(prev => ({
                      ...prev,
                      [clip.clipId]: {
                        loading: false,
                        error: true,
                        url: null
                      }
                    }));
                  }}
                />
              )}
            </div>
          )}
        </div>
      );
    }
    
    // 判断是否为长文本（超过5行或300个字符）
    const isLongText = (clip.content?.length ?? 0) > 300 || 
                      (clip.content?.split('\n').length ?? 0) > 5;
    const isExpanded = expandedClips.has(clip.clipId);

    return (
      <>
        <pre 
          className={`${styles.content} ${!isExpanded && isLongText ? styles.collapsed : ''}`}
        >
          {clip.content}
        </pre>
        {isLongText && (
          <div 
            className={styles.expandButton}
            onClick={() => toggleExpand(clip.clipId)}
          >
            {isExpanded ? '收起' : '展开全文'}
          </div>
        )}
      </>
    );
  };

  // 清理函数
  const cleanupPreviewUrl = useCallback((clipId: string) => {
    const state = imagePreviewStates[clipId];
    if (state?.url) {
      window.URL.revokeObjectURL(state.url);
    }
    setImagePreviewStates(prev => {
      const newState = { ...prev };
      delete newState[clipId];
      return newState;
    });
  }, [imagePreviewStates]);

  // 在组件卸载时清理所有预览URL
  useEffect(() => {
    return () => {
      Object.entries(imagePreviewStates).forEach(([clipId, state]) => {
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

  // 7. 条件渲染
  if (loadingSpaces || isLoadingClips) {
    return (
      <div className={styles.loading}>
        <Spin />
        <div className={styles.loadingText}>
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
    <div className={styles.container}>
      <Card 
        title={renderHeader()} 
        className={styles.pageCard}
      >
        {/* 新增剪贴板输入框 */}
        <div className={styles.newClip}>
          <TextArea
            value={newContent}
            onChange={e => setNewContent(e.target.value)}
            placeholder="输入新的剪贴板内容..."
            autoSize={{ minRows: 3, maxRows: 6 }}
          />
          <div className={styles.inputActions}>
            <input
              type="file"
              id="fileUpload"
              style={{ display: 'none' }}
              onChange={handleFileUpload}
              accept="image/*,text/*,application/pdf"
            />
            <Button 
              icon={<UploadOutlined />}
              onClick={() => document.getElementById('fileUpload')?.click()}
            >
              上传文件
            </Button>
            <Button 
              type="primary"
              icon={<SaveOutlined />}
              onClick={handleSaveNew}
            >
              保存
            </Button>
          </div>
        </div>

        {/* 剪贴板列表 */}
        {isLoadingClips ? (
          <Spin />
        ) : sortedClips.length === 0 ? (
          <Empty description="暂无剪贴板内容" />
        ) : (
          <div className={styles.clipList}>
            {sortedClips.map(clip => (
              <Card 
                key={clip.clipId}
                className={styles.clipItem}
                size="small"
              >
                {/* 头部信息 */}
                <div className={styles.clipHeader}>
                  <div className={styles.clipInfo}>
                    <Text type="secondary">
                      创建于 {dayjs(clip.createdAt).format('YYYY-MM-DD HH:mm')}
                      {clip.updatedAt !== clip.createdAt && 
                        `，更新于 ${dayjs(clip.updatedAt).format('YYYY-MM-DD HH:mm')}`}
                    </Text>
                    <Text type="secondary">
                      创建者: {clip.creator?.username || '游客'}
                    </Text>
                  </div>
                  {renderClipActions(clip)}
                </div>

                {/* 内容部分 */}
                <div className={styles.clipContent}>
                  {editingClipId === clip.clipId ? (
                    <div className={styles.editContent}>
                      <TextArea
                        value={editContent}
                        onChange={e => setEditContent(e.target.value)}
                        autoSize={{ minRows: 3, maxRows: 6 }}
                      />
                      <div className={styles.editActions}>
                        <Button 
                          onClick={() => setEditingClipId(null)}
                        >
                          取消
                        </Button>
                        <Button 
                          type="primary"
                          icon={<SaveOutlined />}
                          onClick={() => handleSaveEdit(clip.clipId)}
                        >
                          保存
                        </Button>
                      </div>
                    </div>
                  ) : (
                    renderClipContent(clip)
                  )}
                </div>
              </Card>
            ))}
          </div>
        )}

        {canManageSpace && showSettings && (
          <SpaceSettingsModal
            visible={showSettings}
            space={currentSpace}
            onClose={() => setShowSettings(false)}
          />
        )}

        {/* 修改图片放大浮窗 */}
        <Modal
          open={!!visibleImage}
          footer={null}
          onCancel={() => {
            setVisibleImage(null);
            setScale(1);
            setPosition({ x: 0, y: 0 });
          }}
          width="90vw"
          style={{ 
            maxWidth: '90vw',
            maxHeight: '90vh',
            padding: 0,
          }}
          centered
          className={styles.imageModal}
          closable={true}
          maskClosable={true}
        >
          <div 
            className={styles.imageModalContent}
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
              style={{ 
                transform: `translate(${position.x}px, ${position.y}px) scale(${scale})`,
                transformOrigin: '0 0',
                pointerEvents: 'none', // 防止图片干扰拖动事件
              }}
            />
          </div>
        </Modal>
      </Card>
    </div>
  );
};

export default ClipsPage; 