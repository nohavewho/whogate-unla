import axios from 'axios';

import { toast } from '../utils/toast';
import { t } from 'i18next';

// Create an axios instance with default config
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Clear any existing token
      window.localStorage.removeItem('token');
      // Only redirect if not already on login page
      if (window.location.pathname !== '/login') {
        window.location.href = '/login';
      }
      // If already on login page, do not redirect, just clear token
    }
    return Promise.reject(error);
  }
);

// Add request interceptor to add token to headers
api.interceptors.request.use(
  (config) => {
    const token = window.localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// API endpoints
export const getMCPServers = async () => {
  try {
    const response = await api.get('/mcp-servers');
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data?.error) {
      toast.error(t('errors.fetch_mcp_servers'), {
        duration: 3000,
      });
    } else {
      toast.error(t('errors.fetch_mcp_servers'), {
        duration: 3000,
      });
    }
    throw error;
  }
};

export const getMCPServer = async (name: string) => {
  try {
    const response = await api.get(`/mcp-servers/${name}`);
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data?.error) {
      toast.error(t('errors.fetch_mcp_server'), {
        duration: 3000,
      });
    } else {
      toast.error(t('errors.fetch_mcp_server'), {
        duration: 3000,
      });
    }
    throw error;
  }
};

export const createMCPServer = async (config: string) => {
  try {
    const response = await api.post('/mcp-servers', config, {
      headers: {
        'Content-Type': 'application/yaml',
      },
    });
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data?.error) {
      toast.error(t('errors.create_mcp_server'), {
        duration: 3000,
      });
    } else {
      toast.error(t('errors.create_mcp_server'), {
        duration: 3000,
      });
    }
    throw error;
  }
};

export const updateMCPServer = async (name: string, config: string) => {
  try {
    const response = await api.put(`/mcp-servers/${name}`, config, {
      headers: {
        'Content-Type': 'application/yaml',
      },
    });
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data?.error) {
      toast.error(t('errors.update_mcp_server'), {
        duration: 3000,
      });
    } else {
      toast.error(t('errors.update_mcp_server'), {
        duration: 3000,
      });
    }
    throw error;
  }
};

export const deleteMCPServer = async (name: string) => {
  try {
    const response = await api.delete(`/mcp-servers/${name}`);
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data?.error) {
      toast.error(t('errors.delete_mcp_server'), {
        duration: 3000,
      });
    } else {
      toast.error(t('errors.delete_mcp_server'), {
        duration: 3000,
      });
    }
    throw error;
  }
};

export const syncMCPServers = async () => {
  try {
    const response = await api.post('/mcp-servers/sync');
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data?.error) {
      toast.error(t('errors.sync_mcp_server'), {
        duration: 3000,
      });
    } else {
      toast.error(t('errors.sync_mcp_server'), {
        duration: 3000,
      });
    }
    throw error;
  }
};

export const getChatMessages = async (sessionId: string, page: number = 1, pageSize: number = 20) => {
  try {
    const response = await api.get(`/chat/sessions/${sessionId}/messages`, {
      params: {
        page,
        pageSize,
      },
    });
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data?.error) {
      toast.error(t('errors.fetch_chat_messages'), {
        duration: 3000,
      });
    } else {
      toast.error(t('errors.fetch_chat_messages'), {
        duration: 3000,
      });
    }
    throw error;
  }
};

export const getChatSessions = async () => {
  try {
    const response = await api.get('/chat/sessions');
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data?.error) {
      toast.error(t('errors.fetch_chat_sessions'), {
        duration: 3000,
      });
    } else {
      toast.error(t('errors.fetch_chat_sessions'), {
        duration: 3000,
      });
    }
    throw error;
  }
};

export const importOpenAPI = async (file: File) => {
  try {
    const formData = new FormData();
    formData.append('file', file);

    const response = await api.post('/openapi/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data?.error) {
      toast.error(t('errors.import_openapi_failed'), {
        duration: 3000,
      });
    } else {
      toast.error(t('errors.import_openapi_failed'), {
        duration: 3000,
      });
    }
    throw error;
  }
};

export default api;
