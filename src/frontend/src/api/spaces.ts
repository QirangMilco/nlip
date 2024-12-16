import http from './http';
import { CreateSpaceRequest, UpdateSpaceRequest, SpaceStats, VerifyInviteTokenResponse, Collaborator } from '@/store/types';
import { Space } from '@/store/types';

export const listSpaces = async (): Promise<Space[]> => {
  const response = await http.get('/spaces/list');
  return response.data.spaces;
};

export const createSpace = async (data: CreateSpaceRequest) => {
  const response = await http.post('/spaces/create', data);
  return response.data;
};

export const updateSpace = async (id: string, data: UpdateSpaceRequest): Promise<Space> => {
  const response = await http.put(`/spaces/${id}`, data);
  return response.data;
};

export const deleteSpace = async (spaceId: string) => {
  const response = await http.delete(`/spaces/${spaceId}`);
  return response.data;
};

export const getSpace = async (id: string): Promise<Space> => {
  const response = await http.get(`/spaces/${id}`);
  return response.data;
};

export const updateCollaboratorPermission = async (
  spaceId: string,
  collaboratorId: string,
  permission: 'edit' | 'view'
  ) => {
    const response = await http.put(`/spaces/${spaceId}/collaborators/update-permissions`, {
    collaboratorId,
    permission,
  });
  return response.data;
};

export const removeCollaborator = async (
  spaceId: string,
  collaboratorId: string
) => {
  const response = await http.delete(`/spaces/${spaceId}/collaborators/remove`, {
    data: { collaboratorId }
  });
  return response.data;
};

export const inviteCollaborator = async (
  spaceId: string,
  email: string,
  permission: 'edit' | 'view'
) => {
  const response = await http.post(`/spaces/${spaceId}/collaborators/invite`, {
    email,
    permission,
  });
  return {
    inviteLink: response.data.inviteLink,
    token: response.data.token
  };
};

// 验证邀请token
export const verifyInviteToken = async (token: string): Promise<VerifyInviteTokenResponse> => {
  const response = await http.post(`/spaces/collaborators/verify-invite`, { token });
  return response.data;
};

// 接受邀请
export const acceptInvite = async (token: string) => {
  const response = await http.post(`/spaces/collaborators/accept-invite`, { token });
  return response.data;
};

export const getSpaceStats = async (spaceId: string): Promise<SpaceStats> => {
  const response = await http.get(`/spaces/${spaceId}/stats`);
  return response.data;
};

export const getSpaceCollaborators = async (spaceId: string): Promise<Collaborator[]> => {
  const response = await http.get(`/spaces/${spaceId}/collaborators`);
  return response.data;
}; 