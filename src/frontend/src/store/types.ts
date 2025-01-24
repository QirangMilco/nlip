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
export interface Collaborator {
  id: string;
  username: string;
  permission: 'edit' | 'view';
}

export interface Space {
  id: string;
  name: string;
  description?: string;
  type: 'private' | 'public';
  ownerId: string;
  maxItems: number;
  retentionDays: number;
  collaborators: Collaborator[];
  createdAt: string;
  updatedAt: string;
}

// 定义权限类型
export type SpacePermission = 'edit' | 'view';

// 扩展 Space 类型以包含权限信息
export interface SpaceWithPermission extends Space {
  permission?: SpacePermission;
  isOwner?: boolean;
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
  ownerUsername: string;
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

export interface VerifyInviteTokenResponse {
  spaceId: string;
  spaceName: string;
  inviterName: string;
  permission: string;
  isCollaborator: boolean;
}

export interface ListCollaboratorsResponse {
  collaborators: Collaborator[];
}

// Token相关类型
export interface Token {
  id: string;
  description: string;
  token: string;
  createdAt: string;
  expiresAt: string;
  lastUsedAt: string;
}

// 创建Token请求参数
export interface CreateTokenRequest {
  description: string;
  expiresAt?: string; // ISO格式日期字符串
}

// 创建Token响应
export interface CreateTokenResponse {
  token: string; // 完整的Token字符串
  tokenInfo: Token;
}

//获取单个Token请求参数
export interface GetTokenRequest {
  tokenId: string;
}

// 获取Token列表响应
export interface ListTokensResponse {
  tokens: Token[];
  maxItems: number;
}

// 获取单个Token响应
export interface GetTokenResponse {
  token: string;
}

// 删除Token请求参数
export interface DeleteTokenRequest {
  tokenId: string;
}