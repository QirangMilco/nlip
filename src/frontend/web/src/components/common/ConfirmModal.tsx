import React from 'react';
import { Modal } from 'antd';
import { ExclamationCircleOutlined } from '@ant-design/icons';
import styles from './ConfirmModal.module.scss';

interface ConfirmModalProps {
  title: string;
  content: React.ReactNode;
  visible: boolean;
  onConfirm: () => void;
  onCancel: () => void;
  confirmLoading?: boolean;
  type?: 'info' | 'success' | 'warning' | 'error';
}

const ConfirmModal: React.FC<ConfirmModalProps> = ({
  title,
  content,
  visible,
  onConfirm,
  onCancel,
  confirmLoading = false,
  type = 'warning'
}) => {
  return (
    <Modal
      title={
        <div className={styles.title}>
          <ExclamationCircleOutlined className={styles[type]} />
          <span>{title}</span>
        </div>
      }
      open={visible}
      onOk={onConfirm}
      onCancel={onCancel}
      confirmLoading={confirmLoading}
      okType={type === 'error' ? 'danger' : 'primary'}
      className={styles.modal}
    >
      <div className={styles.content}>{content}</div>
    </Modal>
  );
};

export default ConfirmModal; 