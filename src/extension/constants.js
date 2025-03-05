export const API_PATHS = {
  AUTH_LOGIN: '/api/v1/nlip/auth/token-login',
  SPACE_CREATE: '/api/v1/nlip/spaces/create',
  SPACE_LIST: '/api/v1/nlip/spaces/list',
  CLIP_UPLOAD: (spaceId) => `/api/v1/nlip/spaces/${spaceId}/clips/upload`,
  CLIP_LIST: (spaceId) => `/api/v1/nlip/spaces/${spaceId}/clips/list`
};

export const STORAGE_KEYS = {
  SETTINGS: 'nlipSettings',
  TOKEN: 'jwtToken',
  SPACE_ID: 'spaceId'
};

export const NOTIFICATION = {
  TITLE: 'Nlip',
  COPY_SUCCESS: '内容已保存到云端剪贴板',
  PASTE_SUCCESS: '内容已粘贴',
  ERROR: {
    TITLE: '错误',
    UPLOAD_FAIL: '上传失败',
    FETCH_FAIL: '获取内容失败'
  }
}; 

export const DEFAULT_SPACE = 'sync';

export const AUTH_WHITELIST = [
  API_PATHS.AUTH_LOGIN
];

export const API_RESPONSE_CODE = {
  SUCCESS: 200,
  ERROR: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  INTERNAL_SERVER_ERROR: 500
};