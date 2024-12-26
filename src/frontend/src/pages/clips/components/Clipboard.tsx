import React from 'react';
import { Space } from '@/store/types';
import TextArea from 'antd/es/input/TextArea';
import { SaveOutlined, UploadOutlined } from '@ant-design/icons';
import SpacePermissionAlert from '@/pages/spaces/components/SpacePermissionAlert';
import ClipItem from './ClipItem';
import { User } from '@/store/types';

interface ClipboardProps {
  currentSpace: Space;
  currentUser: User | null;
  newContent: string;
  onNewContentChange: (content: string) => void;
  onSaveNew: () => void;
  onFileUpload: (e: React.ChangeEvent<HTMLInputElement>) => void;
  // 从ClipsPage传递所有ClipItem需要的props
  clipItemProps: Omit<React.ComponentProps<typeof ClipItem>, 'clip'>;
  sortedClips: React.ComponentProps<typeof ClipItem>['clip'][];
}

const Clipboard: React.FC<ClipboardProps> = ({
  currentSpace,
  currentUser,
  newContent,
  onNewContentChange,
  onSaveNew,
  onFileUpload,
  clipItemProps,
  sortedClips
}) => {
  return (
    <div className="tw-w-full">
      <h1 className="tw-text-2xl tw-font-medium tw-text-gray-800 tw-mb-6">
        {currentSpace?.name}
      </h1>

      {currentSpace && (
        <SpacePermissionAlert
          space={currentSpace} 
          isAdmin={currentUser?.isAdmin}
          isGuest={!currentUser}
        />
      )}

      {/* 新增剪贴板输入框 */}
      <div className="tw-bg-white tw-rounded-lg tw-shadow-sm tw-p-6 tw-mb-8">
        <TextArea
          value={newContent}
          onChange={e => onNewContentChange(e.target.value)}
          placeholder="输入新的内容..."
          className="tw-mb-4"
          autoSize={{ minRows: 3, maxRows: 6 }}
        />
        <div className="tw-flex tw-justify-end tw-gap-3">
          <input
            type="file"
            id="fileUpload"
            className="tw-hidden"
            onChange={onFileUpload}
            accept="image/*,text/*,application/pdf"
          />
          <button 
            className="tw-px-4 tw-py-2 tw-rounded-md tw-transition-colors tw-bg-gray-50 hover:tw-bg-gray-100 tw-text-gray-600"
            onClick={() => document.getElementById('fileUpload')?.click()}
          >
            <UploadOutlined className="tw-mr-1.5" />
            上传文件
          </button>
          <button
            className="tw-px-4 tw-py-2 tw-rounded-md tw-transition-colors tw-bg-primary tw-text-white hover:tw-bg-primary-600"
            onClick={onSaveNew}
          >
            <SaveOutlined className="tw-mr-1.5" />
            保存
          </button>
        </div>
      </div>

      {/* 剪贴板列表 */}
      <div className="tw-space-y-4">
        {sortedClips.map(clip => (
          <ClipItem
            key={clip.clipId}
            clip={clip}
            {...clipItemProps}
          />
        ))}
      </div>
    </div>
  );
};

export default Clipboard;
