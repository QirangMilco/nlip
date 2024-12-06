// 用户相关类型
export interface User {
  id: string;
  username: string;
  isAdmin: boolean;
  createdAt: string;
  updatedAt: string;
}

// 认证相关类型
export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user: User | null;
  needChangePwd: boolean;
}

// 修改密码接口
export interface ChangePasswordRequest {
  oldPassword: string;
  newPassword: string;
}

export interface ChangePasswordResponse {
  code: number;
  message: string;
}

// 空间相关类型
export interface Space {
  id: string;
  name: string;
  type: 'public' | 'private';
  ownerId: string;
  maxItems: number;
  retentionDays: number;
  invitedUsers?: Record<string, 'edit' | 'view'>;
  createdAt: string;
  updatedAt: string;
}

// 剪贴板相关类型
export interface Clip {
  id: string;
  clipId: string;
  spaceId: string;
  contentType: string;
  content?: string;
  filePath?: string;
  creator?: {
    id: string;
    username: string;
  };
  createdAt: string;
  updatedAt: string;
}

export interface UploadClipRequest {
  content?: string;
  file?: File;
  contentType: string;
  spaceId: string;
}

export interface ClipResponse {
  clip: Clip;
}

export interface ListClipsResponse {
  clips: Clip[];
}

export interface CreateSpaceRequest {
  name: string;
  type: 'private' | 'public';
  maxItems: number;
  retentionDays: number;
}

export interface UpdateSpaceRequest {
  name?: string;
  maxItems?: number;
  retentionDays?: number;
}

export interface SpaceStats {
  clipCount: number;
  totalSize: number;
  lastUpdated: string;
}

export interface UpdateClipRequest {
  content: string;
} 

export interface ImagePreviewState {
  loading: boolean;
  error: boolean;
  url: string | null;
  scale: number;
  position: {
    x: number;
    y: number;
  };
}

export interface SpaceSettings {
  default_max_items: number;
  default_retention_days: number;
  max_items_limit: number;
  max_retention_days_limit: number;
}

// 服务器设置类型定义
export interface ServerSettings {
    file_types: {
        allow_list: string[];
        deny_list: string[];
    };
    upload: {
        max_size: number;
    };
    space: SpaceSettings;
    security: {
        token_expiry: string;
    };
}