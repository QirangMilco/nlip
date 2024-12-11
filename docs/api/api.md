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

### Change Password
- **POST** `/auth/change-password`
- **Authentication Required**: Yes
- **Request Body**:
```typescript
{
  oldPassword: string;
  newPassword: string;
}
```
- **Response**:
```typescript
{
  code: 200;
  message: string;
}
```

### Get Current User Info
- **GET** `/auth/me`
- **Authentication Required**: Yes
- **Response**:
```typescript
{
  code: 200;
  data: {
    id: string;
    username: string;
    isAdmin: boolean;
  };
  message: string;
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
- **Response**:
```typescript
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
}
```

### Update Space
- **PUT** `/spaces/:id`
- **Authentication Required**: Yes
- **Request Body**:
```typescript
{
  name?: string;
  maxItems?: number;
  retentionDays?: number;
}
```
- **Response**:
```typescript
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
}
```

### Delete Space
- **DELETE** `/spaces/:id`
- **Authentication Required**: Yes
- **Response**:
```typescript
{
  code: 200;
  message: string;
}
```

### Update Space Settings
- **PUT** `/spaces/:id/settings`
- **Authentication Required**: Yes
- **Request Body**:
```typescript
{
  name?: string;
  maxItems?: number;
  retentionDays?: number;
  visibility?: 'public' | 'private';
}
```
- **Response**:
```typescript
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
}
```

### Collaborator Management APIs

#### Invite Collaborator
- **POST** `/spaces/:id/collaborators/invite`
- **Authentication Required**: Yes
- **Request Body**:
```typescript
{
  email: string;           // Invitee's email
  permission: 'edit' | 'view';  // Permission type
}
```
- **Response**:
```typescript
{
  code: 200;
  message: string;
  data: {
    inviteLink: string;    // Invitation link
  }
}
```

#### Verify Invite Token
- **POST** `/spaces/collaborators/verify-invite`
- **Authentication Required**: Yes
- **Request Body**:
```typescript
{
  token: string;    // Invitation token
}
```
- **Response**:
```typescript
{
  code: 200;
  message: string;
  data: {
    spaceId: string;           // Space ID
    spaceName: string;         // Space name
    inviterName: string;       // Inviter's name
    permission: string;        // Granted permission
    isCollaborator: boolean;   // Whether already a collaborator
    currentPermission: string; // Current permission (if already a collaborator)
  }
}
```

#### Accept Invite
- **POST** `/spaces/collaborators/accept-invite`
- **Authentication Required**: Yes
- **Request Body**:
```typescript
{
  token: string;    // Invitation token
}
```
- **Response**:
```typescript
{
  code: 200;
  message: string;
}
```

#### Remove Collaborator
- **DELETE** `/spaces/:id/collaborators/remove`
- **Authentication Required**: Yes
- **Request Body**:
```typescript
{
  collaboratorId: string;
}
```
- **Response**:
```typescript
{
  code: 200;
  message: string;
}
```

#### Update Collaborator Permissions
- **PUT** `/spaces/:id/collaborators/update-permissions`
- **Authentication Required**: Yes
- **Request Body**:
```typescript
{
  collaboratorId: string;
  permission: 'edit' | 'view';
}
```
- **Response**:
```typescript
{
  code: 200;
  message: string;
}
```

## Clipboard Content APIs

### Upload Content
- **POST** `/spaces/:spaceId/clips/upload`
- **Authentication Required**: Yes (except for guest uploads in public space)
- **Content-Type**: `multipart/form-data`
- **Request Parameters**:
```typescript
{
  file?: File;           // Optional, for file upload
  content?: string;      // Optional, for text content
  contentType: string;   // Content type
  spaceId: string;       // Space ID
}
```
- **Response**:
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

### Guest Upload in Public Space
- **POST** `/spaces/public-space/clips/guest-upload`
- **Authentication Required**: No
- **Content-Type**: `multipart/form-data`
- **Request Parameters**:
```typescript
{
  file?: File;           // Optional, for file upload
  content?: string;      // Optional, for text content
  contentType: string;   // Content type
  spaceId: string;       // Space ID
  creator: string;       // Must be "guest"
}
```
- **Response**:
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

### List Contents
- **GET** `/spaces/:spaceId/clips/list`
- **Authentication Required**: Yes
- **Response**:
```typescript
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
}
```

### Get Single Content
- **GET** `/spaces/:spaceId/clips/:id`
- **Authentication Required**: Yes
- **Query Parameters**:
  - `download`: boolean (optional, if true, download the file)
- **Response**:
```typescript
// If download=false or not specified:
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

// If download=true and it's a file type:
// Directly return file content with appropriate Content-Type and Content-Disposition headers
```

### Delete Content
- **DELETE** `/spaces/:spaceId/clips/:id`
- **Authentication Required**: Yes
- **Response**:
```typescript
{
  code: 204;
}
```

## Error Responses

All APIs return the following format in case of errors:

```typescript
{
  code: number;      // HTTP error status code
  message: string;   // Error description
}
```

Common error status codes:
- 400: Bad request parameters
- 401: Unauthorized or authentication failed
- 403: Insufficient permissions
- 404: Resource not found
- 500: Internal server error

## Notes

1. File Upload Restrictions:
   - Maximum file size: 10MB
   - Supported file types: image/*, text/*, application/pdf

2. Permissions:
   - Regular users can only access their private spaces and public spaces
   - Administrators can access all spaces
   - Only administrators can create public spaces

3. Space Limits:
   - maxItems: Maximum number of items a space can store
   - retentionDays: Number of days content is retained before automatic deletion

4. Security Recommendations:
   - All requests should use HTTPS
   - Tokens should be kept secure and not exposed
   - Sensitive data should be encrypted before uploading

5. Performance Optimization:
   - Use appropriate caching strategies
   - For large file uploads, consider using chunked uploads
   - List data retrieval supports pagination

6. Collaborator Features:
   - Space owners can invite other users as collaborators
   - Collaborator permissions: edit or view
   - Invitation links expire after 24 hours
   - If email is enabled, system will automatically send invitation emails

7. Space Types:
   - public: Public spaces, accessible by all, supports guest uploads
   - private: Private spaces, accessible only by owner and collaborators

## API Version Control

Current API version: v1

- API versions are distinguished by URL path: `/api/v1/nlip`
- New versions will be released at `/api/v2/nlip`
- Old versions will be maintained for a period before deprecation

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

### Event Types

1. clip_created: New clipboard content created
2. clip_updated: Clipboard content updated
3. clip_deleted: Clipboard content deleted
4. space_updated: Space information updated

## Rate Limits

- Login API: 5 requests/minute
- Other APIs: 60 requests/minute
- File uploads: 10 requests/minute

Exceeding these limits will result in a 429 status code.

## Debugging

In development environment:

1. Add `debug=true` query parameter to get detailed error information
2. Use `POST /api/v1/nlip/_debug/log-level` to adjust log level
3. Use `GET /api/v1/nlip/_debug/metrics` to view performance metrics

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
      allow_list: string[];    // Allowed file types
      deny_list: string[];     // Denied file types
    };
    upload: {
      max_size: number;        // Maximum upload file size (bytes)
    };
    space: {
      default_max_items: number;         // Default maximum items per space
      default_retention_days: number;     // Default retention days
      max_items_limit: number;           // Maximum items limit
      max_retention_days_limit: number;   // Maximum retention days limit
    };
    security: {
      token_expiry: string;    // Token expiry time
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
    allow_list?: string[];    // Allowed file types
    deny_list?: string[];     // Denied file types
  };
  space_defaults?: {
    max_items?: number;              // Default maximum items
    retention_days?: number;         // Default retention days
    max_items_limit?: number;        // Maximum items limit
    max_retention_days_limit?: number; // Maximum retention days limit
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

Notes:
1. All settings are optional, only provided settings will be updated
2. File types should be provided as extensions without dot (e.g., "jpg" not ".jpg")
3. Token expiry time should be in time string format (e.g., "24h", "7d")
4. Settings take effect immediately and are saved to configuration file
5. Public space uploads can be done by guests using the `/guest-upload` endpoint
6. Ensure the `creator` field is set to "guest" for guest uploads