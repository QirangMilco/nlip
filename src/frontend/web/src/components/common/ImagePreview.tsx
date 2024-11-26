import React from 'react';
import { Image } from 'antd';
import styles from './ImagePreview.module.scss';

interface ImagePreviewProps {
  src: string;
  alt?: string;
  width?: number | string;
  height?: number | string;
  className?: string;
}

const ImagePreview: React.FC<ImagePreviewProps> = ({
  src,
  alt = '',
  width,
  height,
  className
}) => {
  return (
    <div className={`${styles.container} ${className}`}>
      <Image
        src={src}
        alt={alt}
        width={width}
        height={height}
        placeholder={
          <div className={styles.placeholder}>
            加载中...
          </div>
        }
        fallback="/images/image-error.png"
        preview={{
          mask: '预览',
          maskClassName: styles.mask
        }}
      />
    </div>
  );
};

export default ImagePreview; 