// API响应类型
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
}

// API错误类型
export class ApiError extends Error {
  constructor(
    public code: number,
    message: string,
    public data?: any
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

// 文件上传进度
export interface UploadProgress {
  loaded: number;
  total: number;
  percent: number;
} 