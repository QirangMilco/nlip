import React from 'react';
import { Badge } from 'antd';
import { PresetStatusColorType } from 'antd/es/_util/colors';
import styles from './Status.module.scss';

interface StatusProps {
  type: PresetStatusColorType;
  text: string;
  className?: string;
}

const Status: React.FC<StatusProps> = ({
  type,
  text,
  className
}) => {
  return (
    <Badge
      status={type}
      text={text}
      className={`${styles.status} ${className}`}
    />
  );
};

export default Status; 