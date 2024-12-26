import React from 'react';
import { Spin } from 'antd';
import { 
  LoadingOutlined, ReloadOutlined,
  DownloadOutlined, CopyOutlined,
  EditOutlined, DeleteOutlined,
  SaveOutlined
} from '@ant-design/icons';
import { Clip, ImagePreviewState } from '@/store/types';
import dayjs from 'dayjs';
import TextArea from 'antd/es/input/TextArea';

interface ClipItemProps {
  clip: Clip;
  editingClipId: string | null;
  editContent: string;
  expandedClips: Set<string>;
  imagePreviewStates: Record<string, ImagePreviewState>;
  canEditClip: (clip: Clip) => boolean;
  onEdit: (clip: Clip) => void;
  onSaveEdit: (clipId: string) => void;
  onCancelEdit: () => void;
  onDelete: (clipId: string) => void;
  onDownload: (clip: Clip, type?: 'download' | 'preview') => void;
  onCopy: (clip: Clip) => void;
  onToggleExpand: (clipId: string) => void;
  onEditContentChange: (content: string) => void;
  onImageClick: (url: string | null) => void;
  extractFileName: (filePath: string) => string;
}

const ClipItem: React.FC<ClipItemProps> = ({
  clip,
  editingClipId,
  editContent,
  expandedClips,
  imagePreviewStates,
  canEditClip,
  onEdit,
  onSaveEdit,
  onCancelEdit,
  onDelete,
  onDownload,
  onCopy,
  onToggleExpand,
  onEditContentChange,
  onImageClick,
  extractFileName
}) => {
  return (
    <div className="tw-bg-white tw-rounded-lg tw-shadow-sm tw-p-6">
      {/* 头部信息 */}
      <div className="tw-flex tw-items-center tw-justify-between tw-mb-4">
        <div className="tw-text-sm tw-text-gray-500">
          <span>创建于 {dayjs(clip.createdAt).format('YYYY-MM-DD HH:mm')}</span>
          {clip.updatedAt !== clip.createdAt && (
            <span className="tw-ml-3">
              更新于 {dayjs(clip.updatedAt).format('YYYY-MM-DD HH:mm')}
            </span>
          )}
          <span className="tw-ml-3">
            创建者: {clip.creator?.username || '游客'}
          </span>
        </div>
        
        <div className="tw-flex tw-gap-2">
          {clip.filePath ? (
            <button className="tw-p-2 tw-text-gray-500 hover:tw-text-primary hover:tw-bg-primary-50 tw-rounded-md tw-transition-colors"
              onClick={() => onDownload(clip)}
            >
              <DownloadOutlined />
            </button>
          ) : (
            <button className="tw-p-2 tw-text-gray-500 hover:tw-text-primary hover:tw-bg-primary-50 tw-rounded-md tw-transition-colors"
              onClick={() => onCopy(clip)}
            >
              <CopyOutlined />
            </button>
          )}
          
          {canEditClip(clip) && (
            <>
              {!clip.filePath && (
                <button className="tw-p-2 tw-text-gray-500 hover:tw-text-primary hover:tw-bg-primary-50 tw-rounded-md tw-transition-colors"
                  onClick={() => onEdit(clip)}
                >
                  <EditOutlined />
                </button>
              )}
              <button className="tw-p-2 tw-text-gray-500 hover:tw-text-red-500 hover:tw-bg-red-50 tw-rounded-md tw-transition-colors"
                onClick={() => onDelete(clip.clipId)}
              >
                <DeleteOutlined />
              </button>
            </>
          )}
        </div>
      </div>

      {/* 内容区域 */}
      <div className="tw-mt-3">
        {editingClipId === clip.clipId ? (
          <div className="tw-space-y-4">
            <TextArea
              value={editContent}
              onChange={e => onEditContentChange(e.target.value)}
              autoSize={{ minRows: 3, maxRows: 6 }}
            />
            <div className="tw-flex tw-justify-end tw-gap-3">
              <button 
                className="tw-px-4 tw-py-2 tw-rounded-md tw-transition-colors tw-bg-gray-50 hover:tw-bg-gray-100 tw-text-gray-600"
                onClick={onCancelEdit}
              >
                取消
              </button>
              <button 
                className="tw-px-4 tw-py-2 tw-rounded-md tw-transition-colors tw-bg-primary tw-text-white hover:tw-bg-primary-600"
                onClick={() => onSaveEdit(clip.clipId)}
              >
                <SaveOutlined className="tw-mr-1.5" />
                保存
              </button>
            </div>
          </div>
        ) : (
          <div>
            {clip.filePath ? (
              <div className="tw-space-y-2">
                <div className="tw-text-sm tw-text-gray-900">
                  {decodeURIComponent(extractFileName(clip.filePath))}
                </div>
                {clip.contentType.startsWith('image/') && (
                  <div 
                    className="tw-relative tw-group tw-cursor-pointer"
                    onClick={() => onImageClick(imagePreviewStates[clip.clipId]?.url)}
                  >
                    {imagePreviewStates[clip.clipId]?.loading ? (
                      <div className="tw-flex tw-items-center tw-justify-center tw-h-40 tw-bg-gray-100 tw-rounded">
                        <Spin indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />} />
                      </div>
                    ) : imagePreviewStates[clip.clipId]?.error ? (
                      <div className="tw-flex tw-items-center tw-justify-center tw-h-40 tw-bg-gray-100 tw-rounded">
                        <div className="tw-text-center">
                          <div className="tw-text-gray-500">图片加载失败</div>
                          <button 
                            className="tw-mt-2 tw-text-blue-500 hover:tw-text-blue-600"
                            onClick={() => onDownload(clip, 'preview')}
                          >
                            <ReloadOutlined className="tw-mr-1" />
                            重试
                          </button>
                        </div>
                      </div>
                    ) : (
                      <div className="tw-relative tw-overflow-hidden tw-rounded">
                        <img 
                          src={imagePreviewStates[clip.clipId]?.url || ''}
                          alt={decodeURIComponent(extractFileName(clip.filePath))}
                          className="tw-w-full tw-h-auto tw-transition-transform group-hover:tw-scale-105"
                        />
                        <div className="tw-absolute tw-inset-0 tw-bg-black tw-bg-opacity-0 group-hover:tw-bg-opacity-10 tw-transition-opacity" />
                      </div>
                    )}
                  </div>
                )}
              </div>
            ) : (
              <div>
                <pre className={`tw-whitespace-pre-wrap tw-font-sans tw-text-gray-900 ${
                  !expandedClips.has(clip.clipId) && 
                  ((clip.content?.length ?? 0) > 300 || (clip.content?.split('\n').length ?? 0) > 5)
                    ? 'tw-max-h-40 tw-overflow-hidden'
                    : ''
                }`}>
                  {clip.content}
                </pre>
                {((clip.content?.length ?? 0) > 300 || (clip.content?.split('\n').length ?? 0) > 5) && (
                  <button
                    className="tw-mt-2 tw-text-sm tw-text-blue-500 hover:tw-text-blue-600"
                    onClick={() => onToggleExpand(clip.clipId)}
                  >
                    {expandedClips.has(clip.clipId) ? '收起' : '展开全文'}
                  </button>
                )}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default ClipItem;
