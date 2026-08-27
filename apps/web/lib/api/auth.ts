import { apiClient } from './client';

export interface LoginParams {
  dbType: string;
  port: string;
  username: string;
  password?: string;
}

export interface LoginResponse {
  message: string;
  dbType: string;
  csrfToken: string;
}

export const authApi = {
  login: (data: LoginParams) => {
    return apiClient<LoginResponse>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },
  
  logout: () => {
    return apiClient<{ message: string }>('/api/v1/auth/logout', {
      method: 'POST',
    });
  },
};
