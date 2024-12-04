import http from './http';
import { SpaceSettings } from '@/store/types';

// 获取服务器设置
export const getSettings = async (): Promise<SpaceSettings> => {
    const response = await http.get('/admin/settings');
    return response.data.space;
};

// 更新服务器设置
export const updateSettings = async (settings: Partial<SpaceSettings>): Promise<void> => {
    const response = await http.put('/admin/settings', settings);
    return response.data;
};
