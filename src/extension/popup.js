import React, { useState, useEffect } from 'react';
import { createRoot } from 'react-dom/client';
import { TextField, Button, Box, Typography, CircularProgress } from '@mui/material';
import { makeRequest } from './apiClient';
import { API_PATHS, STORAGE_KEYS, DEFAULT_SPACE, API_RESPONSE_CODE } from './constants';
import { getStorage, setStorage } from './storageUtils';

const Popup = () => {
  const [url, setUrl] = useState('');
  const [username, setUsername] = useState('');
  const [token, setToken] = useState('');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [jwtToken, setJwtToken] = useState('');
  const [spaceId, setSpaceId] = useState('');

  useEffect(() => {
    const loadSettings = async () => {
      const result = await getStorage([STORAGE_KEYS.SETTINGS, STORAGE_KEYS.TOKEN, STORAGE_KEYS.SPACE_ID]);
      const settings = result[STORAGE_KEYS.SETTINGS] || {};
      setUrl(settings.url || '');
      setUsername(settings.username || '');
      setToken(settings.token || '');
      setJwtToken(result[STORAGE_KEYS.TOKEN] || '');
      setSpaceId(result[STORAGE_KEYS.SPACE_ID] || '');
    };
    loadSettings();
  }, []);

  const saveSettings = async () => {
    await setStorage({
      [STORAGE_KEYS.SETTINGS]: { url, username, token }
    });
    setMessage('设置已保存');
    setTimeout(() => setMessage(''), 2000);
  };

  const testConnection = async () => {
    setLoading(true);
    try {
      const response = await makeRequest({
        path: API_PATHS.AUTH_LOGIN,
        method: 'POST',
        body: {
          token: token,
          username: username
        }
      });

      if (!response.ok) {
        throw new Error(response.statusText);
      }
      
      const responseData = await response.json();

      if (responseData.code === API_RESPONSE_CODE.SUCCESS) {
        const newToken = responseData.data.jwtToken;
        await setStorage({ [STORAGE_KEYS.TOKEN]: newToken });
        setJwtToken(newToken);
        await checkAndCreateSpace(newToken);
        setMessage(`登录成功！欢迎 ${responseData.data.user.username}`);
      } else {
        setMessage(`错误: ${responseData.message || '请求失败'}`);
      }
    } catch (error) {
      setMessage(`连接失败: ${error.message}`);
    } finally {
      setLoading(false);
    }
  };

  const checkAndCreateSpace = async (token) => {
    try {
      // 1. 检查是否已有空间
      const spaceCheck = await makeRequest({
        path: API_PATHS.SPACE_LIST,
        method: 'GET',
        headers: {
          Authorization: `Bearer ${token}`
        }
      });
      
      if (!spaceCheck.ok) {
        throw new Error(spaceCheck.statusText);
      }
      
      const checkData = await spaceCheck.json();

      const extensionSpace = checkData.data.spaces.find(space => 
        space.name === DEFAULT_SPACE
      );
      
      // 2. 如果不存在则创建新空间
      if (!extensionSpace) {
        const createResponse = await makeRequest({
          path: API_PATHS.SPACE_CREATE,
          method: 'POST',
          headers: {
            Authorization: `Bearer ${token}`
          },
          body: {
            name: DEFAULT_SPACE,
            maxItems: 5,
            retentionDays: 1
          }
        });

        if (!createResponse.ok) {
          throw new Error('空间创建失败');
        }
        
        const createData = await createResponse.json();
        
        // 3. 存储空间ID
        await setStorage({ 
          spaceId: createData.data.space.id,
        });
      } else {
        // 4. 已有空间则直接存储
        await setStorage({
          spaceId: extensionSpace.id
        });
      }
      
      // 5. 更新状态
      const { spaceId } = await getStorage([STORAGE_KEYS.SPACE_ID]);
      setSpaceId(spaceId);
      
    } catch (error) {
      setMessage(`空间处理失败: ${error.message}`);
    }
  };

  return (
    <Box sx={{ width: 300, p: 2 }}>
      <Typography variant="h6" gutterBottom>Nlip服务器配置</Typography>
      
      <TextField
        label="API地址"
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        fullWidth
        margin="dense"
        placeholder="https://api.example.com"
      />
      
      <TextField
        label="用户名"
        value={username}
        onChange={(e) => setUsername(e.target.value)}
        fullWidth
        margin="dense"
      />
      
      <TextField
        label="Token"
        value={token}
        onChange={(e) => setToken(e.target.value)}
        fullWidth
        margin="dense"
        type="password"
      />

      <Box sx={{ display: 'flex', gap: 1, mt: 2 }}>
        <Button 
          variant="contained" 
          color="primary" 
          onClick={saveSettings}
          disabled={loading}
        >
          保存
        </Button>
        
        <Button 
          variant="outlined" 
          onClick={testConnection}
          disabled={loading || !url || !username || !token}
        >
          {loading ? <CircularProgress size={24} /> : '测试连接'}
        </Button>
      </Box>

      {message && (
        <Typography color={message.startsWith('错误') ? 'error' : 'success'} sx={{ mt: 2 }}>
          {message}
        </Typography>
      )}

      <Box sx={{ mt: 2, pt: 2, borderTop: 1, borderColor: 'divider' }}>
        <Typography variant="body2">
          JWT Token: {jwtToken ? `${jwtToken.slice(0, 4)}****` : '未配置'}
        </Typography>
        <Typography variant="body2" sx={{ mt: 1 }}>
          空间ID: {spaceId || '未配置'}
        </Typography>
      </Box>
    </Box>
  );
};

class ErrorBoundary extends React.Component {
  state = { hasError: false }

  static getDerivedStateFromError(error) {
    return { hasError: true }
  }

  render() {
    if (this.state.hasError) {
      return <Typography color="error">界面渲染出错，请检查控制台</Typography>
    }
    return this.props.children
  }
}

// 替换原来的ReactDOM.render
const root = createRoot(document.getElementById('root'));
root.render(
  <ErrorBoundary>
    <Popup />
  </ErrorBoundary>
); 