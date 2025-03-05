import { makeRequest } from './apiClient';
import { getStorage, setStorage } from './storageUtils';
import { API_PATHS, STORAGE_KEYS, NOTIFICATION, DEFAULT_SPACE } from './constants';

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
    const response = await makeRequest(request);

    // 处理401错误
    if (response.status === 401 && !request._retry) {
      const newToken = await refreshToken();
      if (newToken) {
        // 创建新请求对象避免污染原始请求
        const retryRequest = {
          ...request,
          _retry: true,
          options: {
            ...request.options,
            headers: {
              ...request.options.headers,
              Authorization: `Bearer ${newToken}`
            }
          }
        };
        
        const retryResponse = await fetch(retryRequest.url, retryRequest.options);
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
  const { [STORAGE_KEYS.TOKEN]: token } = await getStorage(STORAGE_KEYS.TOKEN);
  return token || '';
}

async function refreshToken() {
  if (isRefreshing) {
    return new Promise((resolve, reject) => {
      pendingRequests.push({ resolve, reject });
    });
  }
  
  isRefreshing = true;
  try {
    const { [STORAGE_KEYS.SETTINGS]: nlipSettings } = await getStorage(STORAGE_KEYS.SETTINGS);

    const response = await makeRequest({
      path: API_PATHS.AUTH_LOGIN,
      method: 'POST',
      body: {
        token: nlipSettings.token,
        username: nlipSettings.username
      }
    });

    if (response.ok) {
      const data = await response.json();
      const newToken = data.data.token;
      await setStorage({ [STORAGE_KEYS.TOKEN]: newToken });
      return newToken;
    }
  } catch (error) {
    console.error('Token刷新失败:', error);
    pendingRequests.forEach(({ reject }) => reject(error));
    throw error;
  } finally {
    isRefreshing = false;
    pendingRequests.forEach(resolve => resolve());
    pendingRequests = [];
  }
}

async function ensureSpace() {
  let { [STORAGE_KEYS.SPACE_ID]: spaceId } = await getStorage(STORAGE_KEYS.SPACE_ID);
  if (!spaceId) {
    const response = await makeRequest({
      path: API_PATHS.SPACE_CREATE,
      method: 'POST',
      body: {
        name: DEFAULT_SPACE,
        maxItems: 5,
        retentionDays: 1
      }
    });
    
    if (response.ok) {
      const data = await response.json();
      spaceId = data.data.space.id;
      await setStorage({ [STORAGE_KEYS.SPACE_ID]: spaceId });
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
    const { [STORAGE_KEYS.SETTINGS]: nlipSettings } = await getStorage(STORAGE_KEYS.SETTINGS);
    let spaceId = await ensureSpace();
    
    try {
      const response = await makeRequest({
        path: API_PATHS.CLIP_UPLOAD(spaceId),
        method: 'POST',
        body: {
          content: info.selectionText,
          contentType: 'text/plain',
          spaceId: spaceId
        }
      });
      
      if (!response.ok) throw new Error(NOTIFICATION.ERROR.UPLOAD_FAIL);
      chrome.notifications.create({
        type: 'basic',
        iconUrl: 'icon-128.png',
        title: NOTIFICATION.TITLE,
        message: NOTIFICATION.COPY_SUCCESS
      });
    } catch (error) {
      chrome.notifications.create({
        type: 'basic',
        iconUrl: 'icon-128.png',
        title: NOTIFICATION.ERROR.TITLE,
        message: error.message
      });
    }
  }
  
  if (info.menuItemId === 'paste-from-nlip') {
    const { [STORAGE_KEYS.SETTINGS]: nlipSettings } = await getStorage(STORAGE_KEYS.SETTINGS);
    let spaceId = await ensureSpace();
    
    const response = await makeRequest({
      path: API_PATHS.CLIP_LIST(spaceId),
      method: 'GET',
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
          title: NOTIFICATION.TITLE,
          message: NOTIFICATION.PASTE_SUCCESS
        });
      } else {
        chrome.notifications.create({
          type: 'basic',
          iconUrl: 'icon-128.png',
          title: NOTIFICATION.ERROR.TITLE,
          message: '获取剪贴板内容失败'
        });
      }
    }
  }
});

// 处理快捷键命令
chrome.commands.onCommand.addListener(async (command) => {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  const { [STORAGE_KEYS.SETTINGS]: nlipSettings } = await getStorage(STORAGE_KEYS.SETTINGS);
  let spaceId = await ensureSpace();

  try {
    if (command === 'copy-to-nlip') {
      const [selection] = await chrome.scripting.executeScript({
        target: { tabId: tab.id },
        func: () => window.getSelection().toString()
      });

      if (selection?.result) {
        const response = await makeRequest({
          path: API_PATHS.CLIP_UPLOAD(spaceId),
          method: 'POST',
          body: {
            content: selection.result,
            contentType: 'text/plain',
            spaceId: spaceId
          }
        });

        if (!response.ok) throw new Error(NOTIFICATION.ERROR.UPLOAD_FAIL);
        
        chrome.notifications.create({
          type: 'basic',
          iconUrl: 'icon-128.png',
          title: NOTIFICATION.TITLE,
          message: NOTIFICATION.COPY_SUCCESS
        });
      }
    }

    if (command === 'paste-from-nlip') {
      const response = await makeRequest({
        path: API_PATHS.CLIP_LIST(spaceId),
        method: 'GET',
        headers: { Authorization: `Bearer ${await getStoredToken()}` }
      });
      
      if (!response.ok) throw new Error(NOTIFICATION.ERROR.FETCH_FAIL);
      
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
          title: NOTIFICATION.TITLE,
          message: NOTIFICATION.PASTE_SUCCESS
        });
      }
    }
  } catch (error) {
    chrome.notifications.create({
      type: 'basic',
      iconUrl: 'icon-128.png',
      title: NOTIFICATION.ERROR.TITLE,
      message: error.message
    });
  }
}); 