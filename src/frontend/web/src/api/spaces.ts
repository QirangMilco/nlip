import http from './http';
import { CreateSpaceRequest, UpdateSpaceRequest } from '@/store/types';
import { Space } from '@/store/slices/spaceSlice';

export const listSpaces = async (): Promise<Space[]> => {
  const response = await http.get('/spaces/list');
  return response.data;
};

export const createSpace = async (data: CreateSpaceRequest): Promise<Space> => {
  const response = await http.post('/spaces/create', data);
  return response.data;
};

export const updateSpace = async (id: string, data: UpdateSpaceRequest): Promise<Space> => {
  const response = await http.put(`/spaces/${id}`, data);
  return response.data;
};

export const deleteSpace = async (id: string): Promise<void> => {
  await http.delete(`/spaces/${id}`);
}; 