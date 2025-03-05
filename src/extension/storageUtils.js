export const getStorage = (keys) => 
  new Promise(resolve => chrome.storage.local.get(keys, resolve));

export const setStorage = (items) =>
  new Promise(resolve => chrome.storage.local.set(items, resolve)); 