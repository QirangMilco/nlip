/**
 * 安全复制文本到剪贴板的工具函数
 * @param text 需要复制的文本内容
 * @param onSuccess 复制成功回调（可选）
 * @param onError 复制失败回调（可选）
 */
export const copyToClipboard = async (
  text: string,
  onSuccess?: () => void,
  onError?: (error: unknown) => void
): Promise<void> => {
  try {
    // 优先使用现代 Clipboard API
    if (window.isSecureContext && navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(text);
      onSuccess?.();
      return;
    }

    // 降级方案：使用传统方法
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed'; // 避免触发滚动
    document.body.appendChild(textArea);
    textArea.select();

    const success = document.execCommand('copy');
    document.body.removeChild(textArea);

    if (!success) {
      throw new Error('无法访问剪贴板');
    }
    onSuccess?.();
  } catch (err) {
    onError?.(err);
    throw err; // 允许调用方继续处理错误
  }
}; 