chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.type === 'apiRequest') {
    handleApiRequest(request, sendResponse);
    return true; // 保持消息通道开放用于异步响应
  }
  return false;
});

let isRefreshing = false;
let pendingRequests = [];

async function handleApiRequest(request, sendResponse) {
  try {
    // 添加Authorization头
    const headers = {
      ...request.options.headers,
      Authorization: `Bearer ${await getStoredToken()}`
    };

    let response = await fetch(request.url, {
      ...request.options,
      headers
    });

    // 处理401错误
    if (response.status === 401 && !request._retry) {
      const newToken = await refreshToken();
      if (newToken) {
        // 重试原始请求
        const retryResponse = await fetch(request.url, {
          ...request.options,
          headers: {
            ...headers,
            Authorization: `Bearer ${newToken}`
          }
        });
        response = retryResponse;
      }
    }

    const data = await response.json();
    sendResponse({
      success: response.ok,
      status: response.status,
      data
    });
  } catch (error) {
    sendResponse({
      success: false,
      error: error.message
    });
  }
}

async function getStoredToken() {
  return new Promise(resolve => {
    chrome.storage.local.get(['jwtToken'], result => {
      resolve(result.jwtToken || '');
    });
  });
}

async function refreshToken() {
  if (isRefreshing) return new Promise(resolve => pendingRequests.push(resolve));
  
  isRefreshing = true;
  try {
    const { nlipSettings } = await new Promise(resolve => {
      chrome.storage.local.get(['nlipSettings'], resolve);
    });

    const response = await fetch(`${nlipSettings.url}/api/v1/nlip/auth/token-login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        token: nlipSettings.token,
        username: nlipSettings.username
      })
    });

    if (response.ok) {
      const data = await response.json();
      const newToken = data.data.token;
      await new Promise(resolve => chrome.storage.local.set({ jwtToken: newToken }, resolve));
      return newToken;
    }
  } catch (error) {
    console.error('Token刷新失败:', error);
  } finally {
    isRefreshing = false;
    pendingRequests.forEach(resolve => resolve());
    pendingRequests = [];
  }
}

async function ensureSpace() {
  let { spaceId } = await chrome.storage.local.get(['spaceId']);
  if (!spaceId) {
    const { nlipSettings } = await chrome.storage.local.get(['nlipSettings']);
    const response = await fetch(`${nlipSettings.url}/api/v1/nlip/spaces/create`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${await getStoredToken()}`
      },
      body: JSON.stringify({
        name: 'extension',
        maxItems: 5,
        retentionDays: 1
      })
    });
    
    if (response.ok) {
      const data = await response.json();
      spaceId = data.data.space.id;
      await chrome.storage.local.set({ spaceId });
    }
  }
  return spaceId;
}

// 修复右键菜单上下文设置
chrome.runtime.onInstalled.addListener(() => {
  chrome.contextMenus.create({
    id: 'nlip-menu',
    title: 'Nlip',
    contexts: ['selection', 'editable']
  }, () => {
    chrome.contextMenus.create({
      id: 'copy-to-nlip',
      parentId: 'nlip-menu',
      title: '复制到Nlip（Alt+C）',
      contexts: ['selection', 'editable']
    });

    chrome.contextMenus.create({
      id: 'paste-from-nlip',
      parentId: 'nlip-menu',
      title: '从Nlip粘贴（Alt+V）',
      contexts: ['editable']
    });
  });
});

// 修改复制功能实现方式
chrome.contextMenus.onClicked.addListener(async (info, tab) => {
  if (info.menuItemId === 'copy-to-nlip') {
    const { nlipSettings } = await chrome.storage.local.get(['nlipSettings']);
    let spaceId = await ensureSpace();
    
    try {
      const response = await fetch(`${nlipSettings.url}/api/v1/nlip/spaces/${spaceId}/clips/upload`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${await getStoredToken()}`
        },
        body: JSON.stringify({
          content: info.selectionText,
          contentType: 'text/plain',
          spaceId: spaceId
        })
      });
      
      if (!response.ok) throw new Error('上传失败');
      chrome.notifications.create({
        type: 'basic',
        iconUrl: 'icon-128.png',
        title: 'Nlip',
        message: '内容已保存到剪贴板'
      });
    } catch (error) {
      chrome.notifications.create({
        type: 'basic',
        iconUrl: 'icon-128.png',
        title: '错误',
        message: error.message
      });
    }
  }
  
  if (info.menuItemId === 'paste-from-nlip') {
    const { nlipSettings } = await chrome.storage.local.get(['nlipSettings']);
    let spaceId = await ensureSpace();
    
    const response = await fetch(`${nlipSettings.url}/api/v1/nlip/spaces/${spaceId}/clips/list`, {
      headers: {
        Authorization: `Bearer ${await getStoredToken()}`
      }
    });
    
    if (response.ok) {
      const data = await response.json();
      const latestClip = data.data.clips?.[0];
      if (latestClip) {
        await chrome.scripting.executeScript({
          target: { tabId: tab.id },
          func: (content) => {
            const activeElement = document.activeElement;
            if (activeElement.value !== undefined) {
              activeElement.value = content;
            } else if (activeElement.textContent !== undefined) {
              activeElement.textContent = content;
            }
          },
          args: [latestClip.content]
        });
        chrome.notifications.create({
          type: 'basic',
          iconUrl: 'icon-128.png',
          title: 'Nlip',
          message: '内容已粘贴'
        });
      } else {
        chrome.notifications.create({
          type: 'basic',
          iconUrl: 'icon-128.png',
          title: '错误',
          message: '获取剪贴板内容失败'
        });
      }
    }
  }
});

// 处理快捷键命令
chrome.commands.onCommand.addListener(async (command) => {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  const { nlipSettings } = await chrome.storage.local.get(['nlipSettings']);
  let spaceId = await ensureSpace();

  try {
    if (command === 'copy-to-nlip') {
      const [selection] = await chrome.scripting.executeScript({
        target: { tabId: tab.id },
        func: () => window.getSelection().toString()
      });

      if (selection?.result) {
        const response = await fetch(`${nlipSettings.url}/api/v1/nlip/spaces/${spaceId}/clips/upload`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${await getStoredToken()}`
          },
          body: JSON.stringify({
            content: selection.result,
            contentType: 'text/plain',
            spaceId: spaceId
          })
        });

        if (!response.ok) throw new Error('上传失败');
        
        chrome.notifications.create({
          type: 'basic',
          iconUrl: 'icon-128.png',
          title: 'Nlip',
          message: '内容已保存到剪贴板'
        });
      }
    }

    if (command === 'paste-from-nlip') {
      const response = await fetch(`${nlipSettings.url}/api/v1/nlip/spaces/${spaceId}/clips/list`, {
        headers: { Authorization: `Bearer ${await getStoredToken()}` }
      });
      
      if (!response.ok) throw new Error('获取内容失败');
      
      const data = await response.json();
      const latestClip = data.data.clips?.[0];
      if (latestClip) {
        await chrome.scripting.executeScript({
          target: { tabId: tab.id },
          func: (content) => document.execCommand('insertText', false, content),
          args: [latestClip.content]
        });
        
        chrome.notifications.create({
          type: 'basic',
          iconUrl: 'icon-128.png',
          title: 'Nlip',
          message: '内容已粘贴'
        });
      }
    }
  } catch (error) {
    chrome.notifications.create({
      type: 'basic',
      iconUrl: 'icon-128.png',
      title: '错误',
      message: error.message
    });
  }
}); 