import React, { useEffect, useMemo, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useClips } from '@/hooks/useClip';
import { useSpace } from '@/hooks/useSpace';
import { 
  Card, Empty, Spin, Select, Button, Space as AntSpace, 
  Input, message, Typography, Tooltip 
} from 'antd';
import { 
  LogoutOutlined, SettingOutlined, CopyOutlined,
  EditOutlined, DeleteOutlined, UploadOutlined,
  SaveOutlined, DownloadOutlined, FileOutlined
} from '@ant-design/icons';
import { useDispatch, useSelector } from 'react-redux';
import { clearAuth } from '@/store/slices/authSlice';
import { RootState } from '@/store';
import dayjs from 'dayjs';
import SpaceSettingsModal from './components/SpaceSettingsModal';
import { UploadClipRequest, Clip } from '@/store/types';
import styles from './ClipsPage.module.scss';
import { updateClip } from '@/api/clips';

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

  // 4. 计算属性 hooks
  const currentSpace = useMemo(() => 
    spaces.find(s => s.id === spaceId),
    [spaces, spaceId]
  );

  const canManageSpace = useMemo(() => {
    if (!currentUser || !currentSpace) return false;
    if (currentUser.isAdmin) {
      return currentSpace.ownerId === 'system' || currentSpace.ownerId === currentUser.id;
    }
    return currentSpace.ownerId === currentUser.id;
  }, [currentUser, currentSpace]);

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

  const handleDownload = async (clip: Clip) => {
    if (!spaceId || !clip.filePath) {
      message.error('无效的文件');
      return;
    }

    try {
      const blob = await downloadClipFromSpace(spaceId, clip.clipId);
      
      // 生成随机文件名并保留原始扩展名
      const originalFileName = clip.filePath.split('/').pop() || '';
      const fileExt = originalFileName.includes('.') 
        ? `.${originalFileName.split('.').pop()}`
        : '';
      const randomName = `clip-${Math.random().toString(36).substring(2, 10)}${fileExt}`;
      
      // 创建下载链接
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = randomName;
      
      // 触发下载
      document.body.appendChild(a);
      a.click();
      
      // 清理
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      
      message.success('下载成功');
    } catch (error: any) {
      console.error('下载失败:', error);
      message.error(error.message || '下载失败');
    }
  };

  const handleLogout = () => {
    dispatch(clearAuth());
    navigate('/login', { replace: true });
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
    const data: UploadClipRequest = {
      file,
      contentType: file.type,
      spaceId: spaceId
    };
    handleUpload(data);
  };

  // 7. 条件渲染
  if (loadingSpaces) {
    return <Spin />;
  }

  if (spaces.length === 0) {
    return <Empty description="暂无可用空间" />;
  }

  if (!currentSpace) {
    return <Empty description="空间不存在" />;
  }

  // 8. 主要渲染
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
      {/* 新增输入框部分 */}
      <div className={styles.newClip}>
        <TextArea
          value={newContent}
          onChange={e => setNewContent(e.target.value)}
          placeholder="输入新的剪贴板内容..."
          autoSize={{ minRows: 3, maxRows: 6 }}
        />
        <div className={styles.actions}>
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
                  {currentSpace?.type === 'public' && (
                    <Text type="secondary">
                      创建者: {clip.creator?.username || '未知用户'}
                    </Text>
                  )}
                </div>
                <div className={styles.clipActions}>
                  <Tooltip title="复制文本">
                    <Button 
                      type="text" 
                      icon={<CopyOutlined />}
                      onClick={() => handleCopy(clip)}
                      // 如果没有文本内容则禁用复制按钮
                      disabled={!clip.content}
                    />
                  </Tooltip>
                  
                  {/* 如果有文件路径，显示下载按钮 */}
                  {clip.filePath && (
                    <Tooltip title="下载文件">
                      <Button 
                        type="text" 
                        icon={<DownloadOutlined />}
                        onClick={() => handleDownload(clip)}
                      />
                    </Tooltip>
                  )}
                  
                  <Tooltip title="修改">
                    <Button 
                      type="text" 
                      icon={<EditOutlined />}
                      onClick={() => handleEdit(clip)}
                      // 如果是文件类型则禁用修改按钮
                      disabled={!!clip.filePath}
                    />
                  </Tooltip>
                  
                  {(currentSpace?.type === 'private' || 
                    clip.creator?.id === currentUser?.id) && (
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
              </div>

              {/* 内容部分 */}
              <div className={styles.clipContent}>
                {editingClipId === clip.clipId ? (
                  <div className={styles.editContent}>
                    <TextArea
                      value={editContent}
                      onChange={e => setEditContent(e.target.value)}
                      autoSize
                    />
                    <Button 
                      type="primary"
                      size="small"
                      onClick={() => handleSaveEdit(clip.clipId)}
                    >
                      保存
                    </Button>
                  </div>
                ) : (
                  <>
                    {clip.filePath ? (
                      // 文件类型显示
                      <div className={styles.fileContent}>
                        <FileOutlined />
                        <span>{clip.filePath.split('/').pop()}</span>
                        {clip.contentType.startsWith('image/') && (
                          <img 
                            src={`/api/v1/nlip/spaces/${spaceId}/clips/${clip.clipId}/file`}
                            alt="预览图片"
                            className={styles.imagePreview}
                          />
                        )}
                      </div>
                    ) : (
                      // 文本类型显示
                      <>
                        <pre className={`${styles.content} ${!expandedClips.has(clip.clipId) && styles.collapsed}`}>
                          {clip.content}
                        </pre>
                        {(clip.content?.length ?? 0) > 300 && (
                          <Button 
                            type="link" 
                            size="small"
                            onClick={() => toggleExpand(clip.clipId)}
                          >
                            {expandedClips.has(clip.clipId) ? '收起' : '展开'}
                          </Button>
                        )}
                      </>
                    )}
                  </>
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
    </Card>
  );
};

export default ClipsPage; 