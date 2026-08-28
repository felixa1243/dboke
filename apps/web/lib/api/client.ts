const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}

export async function apiClient<T>(path: string, options?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string> || {}),
  };

  if (typeof window !== 'undefined') {
    const csrf = localStorage.getItem('dboke_csrf_token');
    if (csrf) {
      headers['X-CSRF-Token'] = csrf;
    }
  }

  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    credentials: 'include',
    headers,
  });
  
  if (!res.ok) {
    let message = 'An API error occurred';
    try {
      const text = await res.text();
      if (text) {
        try {
          const errorData = JSON.parse(text);
          message = errorData.message || message;
        } catch {
          message = text;
        }
      }
    } catch {
      // Ignore further read errors
    }
    throw new ApiError(res.status, message);
  }
  
  return res.json();
}
