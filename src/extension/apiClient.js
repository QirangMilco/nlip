import { getStorage } from './storageUtils';
import { STORAGE_KEYS, AUTH_WHITELIST } from './constants';

export const makeRequest = async (request) => {
  const { [STORAGE_KEYS.SETTINGS]: nlipSettings } = await getStorage(STORAGE_KEYS.SETTINGS);
  const { [STORAGE_KEYS.TOKEN]: token } = await getStorage(STORAGE_KEYS.TOKEN);
  
  const defaultHeaders = {
    'Content-Type': 'application/json'
  };

  if (!AUTH_WHITELIST.includes(request.path)) {
    defaultHeaders.Authorization = `Bearer ${token}`;
  }

  return fetch(`${nlipSettings.url}${request.path}`, {
    method: request.method || 'GET',
    headers: { ...defaultHeaders, ...request.headers },
    body: request.body ? JSON.stringify(request.body) : undefined
  });
}; 