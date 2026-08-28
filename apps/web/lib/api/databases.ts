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
  getTableSchema: (database: string, table: string) => {
    return apiClient<{ columns: any[] }>(`/api/v1/databases/${database}/tables/${table}/columns`, {
      method: 'GET',
    });
  },
  executeQuery: (database: string, query: string, params: any[] = [], limit: number = 200, offset: number = 0) => {
    return apiClient<any>(`/api/v1/databases/${database}/query`, {
      method: 'POST',
      body: JSON.stringify({ query, params, limit, offset }),
    });
  },
};
