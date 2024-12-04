import React from 'react';
import { Card as AntCard } from 'antd';
import styles from './Card.module.scss';

interface CardProps {
  title?: React.ReactNode;
  extra?: React.ReactNode;
  children: React.ReactNode;
  loading?: boolean;
  className?: string;
}

const Card: React.FC<CardProps> = ({
  title,
  extra,
  children,
  loading = false,
  className
}) => {
  return (
    <AntCard
      title={title}
      extra={extra}
      loading={loading}
      className={`${styles.card} ${className}`}
    >
      {children}
    </AntCard>
  );
};

export default Card; 