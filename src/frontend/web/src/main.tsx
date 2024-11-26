import React from 'react';
import ReactDOM from 'react-dom/client';
import { Provider } from 'react-redux';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { store } from './store';
import App from './App';

// 导入样式
import 'antd/dist/reset.css';
import './styles/index.scss';

// 使用立即执行函数来处理异步操作
(async () => {
  try {
    const rootElement = document.getElementById('root');
    if (!rootElement) throw new Error('Failed to find the root element');

    const root = ReactDOM.createRoot(rootElement);

    root.render(
      <Provider store={store}>
        <ConfigProvider locale={zhCN}>
          <App />
        </ConfigProvider>
      </Provider>
    );
  } catch (error) {
    console.error('Application initialization failed:', error);
  }
})(); 