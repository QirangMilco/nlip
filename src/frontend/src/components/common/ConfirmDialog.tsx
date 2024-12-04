import React from 'react';
import { Modal } from 'antd';
import { ExclamationCircleOutlined } from '@ant-design/icons';

interface ConfirmDialogProps {
  title: string;
  content: React.ReactNode;
  onConfirm: () => void | Promise<void>;
  okText?: string;
  cancelText?: string;
  type?: 'info' | 'success' | 'warning' | 'error';
}

const ConfirmDialog: React.FC<ConfirmDialogProps> = (props) => {
  const { confirm } = Modal;

  React.useEffect(() => {
    confirm({
      title: props.title,
      icon: <ExclamationCircleOutlined />,
      content: props.content,
      okText: props.okText ?? '确定',
      cancelText: props.cancelText ?? '取消',
      okType: props.type === 'error' ? 'danger' : 'primary',
      onOk: props.onConfirm,
    });
  }, []);

  return null;
};

export default ConfirmDialog; 