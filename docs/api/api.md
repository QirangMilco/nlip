# NLIP API Documentation

## Basic Information

- Base URL: `/api/v1/nlip`
- All requests require `Authorization: Bearer {token}` in header (except login and register)
- Response format:
```typescript
{
  code: number;      // HTTP status code
  message: string;   // Response message
  data?: any;        // Response data (optional)
}
```

## Authentication APIs

### Login
- **POST** `/auth/login`
- **Request Body**:
```typescript
{
  username: string;
  password: string;
}
```
- **Response**:
```typescript
{
  token: string;
  user: {
    id: string;
    username: string;
    isAdmin: boolean;
    createdAt: string;
  };
  needChangePwd: boolean;
}
```

### Register
- **POST** `/auth/register`
- **Request Body**:
```typescript
{
  username: string;
  password: string;
}
```
- **Response**:
```typescript
{
  token: string;
  user: {
    id: string;
    username: string;
    isAdmin: boolean;
    createdAt: string;
  }
}
```

## Space Management APIs

### List Spaces
- **GET** `/spaces/list`
- **Authentication Required**: Yes
- **Response**:
```typescript
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
}
```

### Create Space
- **POST** `/spaces/create`
- **Authentication Required**: Yes
- **Request Body**:
```typescript
{
  name: string;
  type: 'public' | 'private';
  maxItems?: number;
  retentionDays?: number;
}
```

## Clipboard Content APIs

### Upload Content
- **POST** `/spaces/:spaceId/clips/upload`
- **Authentication Required**: Yes
- **Content-Type**: `multipart/form-data`
- **Request Parameters**:
```typescript
{
  file?: File;           // Optional, for file upload
  content?: string;      // Optional, for text content
  contentType: string;   // Content type
  spaceId: string;      // Space ID
}
```

### List Contents
- **GET** `/spaces/:spaceId/clips/list`
- **Authentication Required**: Yes

### Update Clip Content
- **PUT** `/spaces/:spaceId/clips/:clipId`
- **Authentication Required**: Yes
- **Request Body**:
```typescript
{
  content: string;
}
```
- **Response**:
```typescript
{
  code: 200;
  data: {
    clip: {
      id: string;
      content: string;
      contentType: string;
      spaceId: string;
      creatorId: string;
      creator?: {
        id: string;
        username: string;
      };
      createdAt: string;
      updatedAt: string;
    }
  };
  message: string;
}
```

### Delete Clip
- **DELETE** `/spaces/:spaceId/clips/:clipId`
- **Authentication Required**: Yes
- **Response**:
```typescript
{
  code: 200;
  message: string;
}
```

## WebSocket API

### Connection
- **URL**: `ws://domain/api/v1/nlip/ws`
- **Authentication Required**: Yes (via URL parameter token)
- **Example**: `ws://domain/api/v1/nlip/ws?token={jwt_token}`

### Message Format
```typescript
interface WSMessage {
  type: 'clip_created' | 'clip_updated' | 'clip_deleted' | 'space_updated';
  data: any;
  timestamp: number;
}
```

## Rate Limits

- Login API: 5 requests/minute
- Other APIs: 60 requests/minute
- File uploads: 10 requests/minute

Exceeding these limits will result in a 429 status code.

## Security Notes

1. File Upload Restrictions:
   - Maximum file size: 10MB
   - Supported file types: image/*, text/*, application/pdf

2. Permissions:
   - Regular users can only access their private spaces and public spaces
   - Administrators can access all spaces
   - Only administrators can create public spaces

## Version Control

Current API version: v1

- API versions are distinguished by URL path: `/api/v1/nlip`
- New versions will be released at `/api/v2/nlip`
- Old versions will be maintained for a period before deprecation

## Changelog

### v1.0.0 (2024-03-01)
- Initial release
- Basic CRUD operations
- User authentication and authorization
- WebSocket real-time notifications 

## Administrator APIs

### Get Server Settings
- **GET** `/admin/settings`
- **Authentication Required**: Yes (Admin only)
- **Response**:
```typescript
{
  code: 200;
  data: {
    file_types: {
      allow_list: string[];
      deny_list: string[];
    };
    upload: {
      max_size: number;
    };
    space: {
      default_max_items: number;
      default_retention_days: number;
    };
    security: {
      token_expiry: string;
    };
  };
  message: string;
}
```

### Update Server Settings
- **PUT** `/admin/settings`
- **Authentication Required**: Yes (Admin only)
- **Request Body**:
```typescript
{
  file_types?: {
    allow_list?: string[];
    deny_list?: string[];
  };
  upload?: {
    max_size?: number;
  };
  space?: {
    default_max_items?: number;
    default_retention_days?: number;
  };
  security?: {
    token_expiry?: string;
  };
}
```
- **Response**:
```typescript
{
  code: 200;
  message: string;
}
```