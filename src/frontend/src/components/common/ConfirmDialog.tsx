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

const ConfirmDialog: React.FC<ConfirmDialogProps> = ({
  title,
  content,
  onConfirm,
  okText = '确定',
  cancelText = '取消',
  type = 'warning'
}) => {
  const { confirm } = Modal;

  const showConfirm = () => {
    confirm({
      title,
      icon: <ExclamationCircleOutlined />,
      content,
      okText,
      cancelText,
      okType: type === 'error' ? 'danger' : 'primary',
      onOk: onConfirm,
    });
  };

  return showConfirm();
};

export default ConfirmDialog; 