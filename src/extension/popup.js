import React, { useState, useEffect } from 'react';
import { createRoot } from 'react-dom/client';
import { TextField, Button, Box, Typography, CircularProgress } from '@mui/material';

const Popup = () => {
  const [url, setUrl] = useState('');
  const [username, setUsername] = useState('');
  const [token, setToken] = useState('');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [jwtToken, setJwtToken] = useState('');
  const [spaceId, setSpaceId] = useState('');

  useEffect(() => {
    chrome.storage.local.get(['nlipSettings', 'jwtToken', 'spaceId'], (result) => {
      const settings = result.nlipSettings || {};
      setUrl(settings.url || '');
      setUsername(settings.username || '');
      setToken(settings.token || '');
      setJwtToken(result.jwtToken || '');
      setSpaceId(result.spaceId || '');
    });
  }, []);

  const saveSettings = () => {
    chrome.storage.local.set({
      nlipSettings: { url, username, token }
    });
    setMessage('设置已保存');
    setTimeout(() => setMessage(''), 2000);
  };

  const testConnection = async () => {
    setLoading(true);
    try {
      const response = await chrome.runtime.sendMessage({
        type: 'apiRequest',
        url: `${url}/api/v1/nlip/auth/token-login`,
        options: {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            token: token,
            username: username
          })
        }
      });

      if (response.success) {
        const newToken = response.data.data.jwtToken;
        chrome.storage.local.set({ jwtToken: newToken });
        setJwtToken(newToken);
        checkAndCreateSpace(newToken);
        setMessage(`登录成功！欢迎 ${response.data.data.user.username}`);
      } else {
        setMessage(`错误: ${response.data?.message || '请求失败'}`);
      }
    } catch (error) {
      setMessage('连接失败，请检查URL和网络');
    } finally {
      setLoading(false);
    }
  };

  const checkAndCreateSpace = async (token) => {
    try {
      const spaceResponse = await chrome.runtime.sendMessage({
        type: 'apiRequest',
        url: `${url}/api/v1/nlip/spaces/list`,
        options: {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        }
      });

      if (spaceResponse.success) {
        const extensionSpace = spaceResponse.data.data.spaces.find(space => 
          space.name === 'extension'
        );

        if (!extensionSpace) {
          const createResponse = await chrome.runtime.sendMessage({
            type: 'apiRequest',
            url: `${url}/api/v1/nlip/spaces/create`,
            options: {
              method: 'POST',
              headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
              },
              body: JSON.stringify({
                name: 'extension',
                maxItems: 5,
                retentionDays: 1
              })
            }
          });

          if (createResponse.success) {
            chrome.storage.local.set({ spaceId: createResponse.data.data.id });
            setSpaceId(createResponse.data.data.space.id);
          }
        } else {
          chrome.storage.local.set({ spaceId: extensionSpace.id });
          setSpaceId(extensionSpace.id);
        }
      }
    } catch (error) {
      console.error('空间管理错误:', error);
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