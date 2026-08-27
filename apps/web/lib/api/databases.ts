import { apiClient } from './client';

export interface DatabaseResponse {
  databases: string[];
}

export interface Table {
  name: string;
  type: string;
  rows: string;
  size: string;
}

export interface TablesResponse {
  tables: Table[];
}

export const databasesApi = {
  getDatabases: () => {
    return apiClient<DatabaseResponse>('/api/v1/databases', {
      method: 'GET',
    });
  },
  getTables: (database: string) => {
    return apiClient<TablesResponse>(`/api/v1/databases/${database}/tables`, {
      method: 'GET',
    });
  },
};
