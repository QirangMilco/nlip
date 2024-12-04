// 定义公共路由常量
export const PUBLIC_ROUTES = [
  '/login',
  '/register',
  '/change-password',
  '/clips/public-space',
] as const;

// 检查路径是否匹配公共路由
export const isPublicRoute = (path: string): boolean => {
  return PUBLIC_ROUTES.some(route => path.includes(route));
}; 