import React from 'react';
import { Button } from 'antd';
import { ButtonProps } from 'antd/es/button';
import { Tooltip } from 'antd';
import styles from './IconButton.module.scss';

interface IconButtonProps extends ButtonProps {
  icon: React.ReactNode;
  tooltip?: string;
  className?: string;
}

const IconButton: React.FC<IconButtonProps> = ({
  icon,
  tooltip,
  className,
  ...props
}) => {
  const button = (
    <Button
      type="text"
      icon={icon}
      className={`${styles.button} ${className}`}
      {...props}
    />
  );

  if (tooltip) {
    return (
      <Tooltip title={tooltip}>
        {button}
      </Tooltip>
    );
  }

  return button;
};

export default IconButton; 