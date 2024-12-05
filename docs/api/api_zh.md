# NLIP API 文档

## 基础信息

- 基础URL: `/api/v1/nlip`
- 所有请求需要在header中携带 `Authorization: Bearer {token}` (除了登录和注册接口)
- 响应格式:  ```typescript
  {
    code: number;      // HTTP状态码
    message: string;   // 响应消息
    data?: any;       // 响应数据(可选)
  }  ```

## 认证相关 API

### 登录
- **POST** `/auth/login`
- **请求体**:  ```typescript
  {
    username: string;
    password: string;
  }  ```
- **响应**:  ```typescript
  {
    token: string;
    user: {
      id: string;
      username: string;
      isAdmin: boolean;
      createdAt: string;
    };
    needChangePwd: boolean;
  }  ```

### 注册
- **POST** `/auth/register`
- **请求体**:  ```typescript
  {
    username: string;
    password: string;
  }  ```
- **响应**:  ```typescript
  {
    token: string;
    user: {
      id: string;
      username: string;
      isAdmin: boolean;
      createdAt: string;
    }
  }  ```

### 修改密码
- **POST** `/auth/change-password`
- **需要认证**: 是
- **请求体**:  ```typescript
  {
    oldPassword: string;
    newPassword: string;
  }  ```
- **响应**:  ```typescript
  {
    code: 200;
    message: string;
  }  ```

### 获取当前用户信息
- **GET** `/auth/me`
- **需要认证**: 是
- **响应**:  ```typescript
  {
    code: 200;
    data: {
      id: string;
      username: string;
      isAdmin: boolean;
    };
    message: string;
  }  ```

## 空间管理 API

### 获取空间列表
- **GET** `/spaces/list`
- **需要认证**: 是
- **响应**:  ```typescript
  {
    code: 200;
    data: {
      spaces: Array<{
        id: string;
        name: string;
        type: 'public' | 'private';
        ownerId: string;
        maxItems: number;
        retentionDays: number;
        createdAt: string;
      }>;
    };
    message: string;
  }  ```

### 创建空间
- **POST** `/spaces/create`
- **需要认证**: 是
- **请求体**:  ```typescript
  {
    name: string;
    type: 'public' | 'private';
    maxItems?: number;
    retentionDays?: number;
  }  ```
- **响应**:  ```typescript
  {
    code: 201;
    data: {
      space: {
        id: string;
        name: string;
        type: string;
        ownerId: string;
        maxItems: number;
        retentionDays: number;
        createdAt: string;
      };
    };
    message: string;
  }  ```

### 更新空间
- **PUT** `/spaces/:id`
- **需要认证**: 是
- **请求体**:  ```typescript
  {
    name?: string;
    maxItems?: number;
    retentionDays?: number;
  }  ```
- **响应**:  ```typescript
  {
    code: 200;
    data: {
      space: {
        id: string;
        name: string;
        type: string;
        ownerId: string;
        maxItems: number;
        retentionDays: number;
        createdAt: string;
      };
    };
    message: string;
  }  ```

### 删除空间
- **DELETE** `/spaces/:id`
- **需要认证**: 是
- **响应**:  ```typescript
  {
    code: 200;
    message: string;
  }  ```

### 更新空间设置
- **PUT** `/spaces/:id/settings`
- **需要认证**: 是
- **请求体**:  ```typescript
  {
    name?: string;
    maxItems?: number;
    retentionDays?: number;
    visibility?: 'public' | 'private';
  }  ```
- **响应**:  ```typescript
  {
    code: 200;
    data: {
      space: {
        id: string;
        name: string;
        type: string;
        ownerId: string;
        maxItems: number;
        retentionDays: number;
        createdAt: string;
      };
    };
    message: string;
  }  ```

### 协作者管理 API

#### 邀请协作者
- **POST** `/spaces/:id/collaborators/invite`
- **需要认证**: 是
- **请求体**:  ```typescript
  {
    collaboratorId: string;
    permission: 'edit' | 'view';
  }  ```
- **响应**:  ```typescript
  {
    code: 200;
    message: string;
  }  ```

#### 删除协作者
- **DELETE** `/spaces/:id/collaborators/remove`
- **需要认证**: 是
- **请求体**:  ```typescript
  {
    collaboratorId: string;
  }  ```
- **响应**:  ```typescript
  {
    code: 200;
    message: string;
  }  ```

#### 更新协作者权限
- **PUT** `/spaces/:id/collaborators/update-permissions`
- **需要认证**: 是
- **请求体**:  ```typescript
  {
    collaboratorId: string;
    permission: 'edit' | 'view';
  }  ```
- **响应**:  ```typescript
  {
    code: 200;
    message: string;
  }  ```

## 剪贴板内容 API

### 上传内容
- **POST** `/spaces/:spaceId/clips/upload`
- **需要认证**: 是（公共空间的游客上传除外）
- **Content-Type**: `multipart/form-data`
- **请求参数**:
```typescript
{
  file?: File;           // 可选，如果上传文件
  content?: string;      // 可选，如果是文本内容
  contentType: string;   // 内容类型
  spaceId: string;       // 所属空间ID
}
```
- **响应**:
```typescript
{
  code: 201;
  data: {
    clip: {
      id: string;
      clipId: string;
      spaceId: string;
      contentType: string;
      content?: string;
      filePath?: string;
      createdAt: string;
    };
  };
  message: string;
}
```

### 公共空间游客上传
- **POST** `/spaces/public-space/clips/guest-upload`
- **需要认证**: 否
- **Content-Type**: `multipart/form-data`
- **请求参数**:
```typescript
{
  file?: File;           // 可选，如果上传文件
  content?: string;      // 可选，如果是文本内容
  contentType: string;   // 内容类型
  spaceId: string;       // 所属空间ID
  creator: string;       // 必须为 "guest"
}
```
- **响应**:
```typescript
{
  code: 201;
  data: {
    clip: {
      id: string;
      clipId: string;
      spaceId: string;
      contentType: string;
      content?: string;
      filePath?: string;
      createdAt: string;
    };
  };
  message: string;
}
```

### 获取内容列表
- **GET** `/spaces/:spaceId/clips/list`
- **需要认证**: 是
- **响应**:  ```typescript
  {
    code: 200;
    data: {
      clips: Array<{
        id: string;
        clipId: string;
        spaceId: string;
        contentType: string;
        content?: string;
        filePath?: string;
        createdAt: string;
      }>;
    };
    message: string;
  }  ```

### 获取单个内容
- **GET** `/spaces/:spaceId/clips/:id`
- **需要认证**: 是
- **查询参数**:
  - `download`: boolean (可选，如果为true则下载文件)
- **响应**:  ```typescript
  // 如果download=false或未指定:
  {
    code: 200;
    data: {
      clip: {
        id: string;
        clipId: string;
        spaceId: string;
        contentType: string;
        content?: string;
        filePath?: string;
        createdAt: string;
      };
    };
    message: string;
  }
  
  // 如果download=true且是文件类型:
  // 直接返回文件内容，带有适当的Content-Type和Content-Disposition头  ```

### 删除内容
- **DELETE** `/spaces/:spaceId/clips/:id`
- **需要认证**: 是
- **响应**:   ```typescript
  {
    code: 204;
  }  ```

## 错误响应

所有API在发生错误时都会返回以下格式的响应：

```typescript
{
  code: number;      // HTTP错误状态码
  message: string;   // 错误描述
}
```

常见错误状态码：
- 400: 请求参数错误
- 401: 未认证或认证失败
- 403: 权限不足
- 404: 资源不存在
- 500: 服务器内部错误

## 注意事项

1. 文件上传限制：
   - 最大文件大小: 10MB
   - 支持的文件类型: image/*, text/*, application/pdf

2. 权限说明：
   - 普通用户只能访问自己创建的私有空间和公共空间
   - 管理员可以访问所有空间
   - 只有管理员可以创建公共空间

3. 空间限制：
   - maxItems: 空间可存储的最大条目数
   - retentionDays: 内容保留天数，超过后自动删除

4. 安全建议：
   - 所有请求都应使用HTTPS
   - Token应妥善保管,不要泄露
   - 敏感数据建议加密后再上传

5. 性能优化：
   - 建议使用合适的缓存策略
   - 大文件上传建议使用分片上传
   - 获取列表数据支持分页

## API版本控制

当前API版本: v1

- API版本通过URL路径区分: `/api/v1/nlip`
- 新版本会在 `/api/v2/nlip` 发布
- 旧版本会持续维护一段时间,但最终会被废弃

## WebSocket API

### 连接
- **URL**: `ws://domain/api/v1/nlip/ws`
- **需要认证**: 是 (通过URL参数传递token)
- **示例**: `ws://domain/api/v1/nlip/ws?token={jwt_token}`

### 消息格式

```typescript
interface WSMessage {
  type: 'clip_created' | 'clip_updated' | 'clip_deleted' | 'space_updated';
  data: any;
  timestamp: number;
}
```

### 事件类型

1. clip_created: 新建剪贴板内容
2. clip_updated: 更新剪贴板内容
3. clip_deleted: 删除剪贴板内容
4. space_updated: 空间信息更新

## 速率限制

- 登录接口: 5次/分钟
- 其他API: 60次/分钟
- 文件上传: 10次/分钟

超过限制将返回429状态码。

## 调试

开发环境下可以:

1. 添加 `debug=true` 查询参数获取详细错误信息
2. 使用 `POST /api/v1/nlip/_debug/log-level` 调整日志级别
3. 通过 `GET /api/v1/nlip/_debug/metrics` 查看性能指标

## 管理员 API

### 获取服务器设置
- **GET** `/admin/settings`
- **需要认证**: 是（仅管理员）
- **响应**:
```typescript
{
  code: 200;
  data: {
    file_types: {
      allow_list: string[];    // 允许的文件类型列表
      deny_list: string[];     // 禁止的文件类型列表
    };
    upload: {
      max_size: number;        // 最大上传文件大小(字节)
    };
    space: {
      default_max_items: number;       // 空间默认最大条目数
      default_retention_days: number;   // 空间默认保留天数
    };
    security: {
      token_expiry: string;    // 令牌过期时间
    };
  };
  message: string;
}
```

### 更新服务器设置
- **PUT** `/admin/settings`
- **需要认证**: 是（仅管理员）
- **请求体**:
```typescript
{
  file_types?: {
    allow_list?: string[];    // 允许的文件类型列表
    deny_list?: string[];     // 禁止的文件类型列表
  };
  upload?: {
    max_size?: number;        // 最大上传文件大小(字节)
  };
  space?: {
    default_max_items?: number;       // 空间默认最大条目数
    default_retention_days?: number;   // 空间默认保留天数
  };
  security?: {
    token_expiry?: string;    // 令牌过期时间
  };
}
```
- **响应**:
```typescript
{
  code: 200;
  message: string;
}
```

注意：
1. 所有设置项都是可选的，只更新提供的设置
2. 文件类型只需提供扩展名，不含点号(如: "jpg" 而不是 ".jpg")
3. 令牌过期时间格式为时间字符串(如: "24h", "7d")
4. 更新设置后会立即生效并保存到配置文件
5. 公共空间的上传可以通过 `/guest-upload` 接口由游客完成。
6. 确保游客上传时 `creator` 字段设置为 "guest"。
