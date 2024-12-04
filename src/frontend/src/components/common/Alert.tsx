import React from 'react';
import { Alert as AntAlert } from 'antd';
import { AlertProps as AntAlertProps } from 'antd/es/alert';
import styles from './Alert.module.scss';

interface AlertProps extends AntAlertProps {
  className?: string;
}

const Alert: React.FC<AlertProps> = ({
  className,
  ...props
}) => {
  return (
    <AntAlert
      className={`${styles.alert} ${className}`}
      {...props}
    />
  );
};

export default Alert; 