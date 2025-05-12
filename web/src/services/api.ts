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

// Tenant related API functions
export const getTenants = async () => {
  try {
    const response = await api.get('/auth/tenants');
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data?.error) {
      toast.error(t('errors.fetch_tenants'), {
        duration: 3000,
      });
    } else {
      toast.error(t('errors.fetch_tenants'), {
        duration: 3000,
      });
    }
    throw error;
  }
};

export const getTenant = async (name: string) => {
  try {
    const response = await api.get(`/auth/tenants/${name}`);
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data?.error) {
      toast.error(t('errors.fetch_tenant'), {
        duration: 3000,
      });
    } else {
      toast.error(t('errors.fetch_tenant'), {
        duration: 3000,
      });
    }
    throw error;
  }
};

export const createTenant = async (data: { name: string; prefix: string; description: string }) => {
  try {
    // 确保前缀以/开头
    let prefix = data.prefix;
    if (prefix && !prefix.startsWith('/')) {
      prefix = `/${prefix}`;
    }

    // 检查是否为根级别目录
    if (prefix === '/') {
      toast.error(t('tenants.root_prefix_not_allowed'), {
        duration: 3000,
      });
      throw new Error('Root prefix not allowed');
    }

    // 先获取所有租户，检查前缀冲突
    const tenants = await getTenants();
    if (checkPrefixConflict(prefix, tenants.map((t: any) => t.prefix))) {
      toast.error(t('tenants.prefix_path_conflict'), {
        duration: 3000,
      });
      throw new Error('Prefix path conflict');
    }

    const response = await api.post('/auth/tenants', {
      ...data,
      prefix,
    });
    toast.success(t('tenants.add_success'), {
      duration: 3000,
    });
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 409) {
      // 检查具体的错误消息，区分名称冲突和前缀冲突
      const errorMessage = error.response.data?.error;
      if (errorMessage === "Tenant name already exists") {
        toast.error(t('tenants.name_conflict'), {
          duration: 3000,
        });
      } else {
        toast.error(t('tenants.prefix_conflict'), {
          duration: 3000,
        });
      }
    } else if (!(error instanceof Error && 
               (error.message === 'Root prefix not allowed' || 
                error.message === 'Prefix path conflict'))) {
      toast.error(t('errors.create_tenant'), {
        duration: 3000,
      });
    }
    throw error;
  }
};

// 检查前缀是否与现有前缀有冲突（相同、父路径或子路径）
const checkPrefixConflict = (prefix: string, existingPrefixes: string[], excludePrefix?: string): boolean => {
  for (const existingPrefix of existingPrefixes) {
    // 跳过当前正在编辑的前缀（仅在更新时使用）
    if (excludePrefix && existingPrefix === excludePrefix) {
      continue;
    }

    // 检查是否为父路径 - 例如 /a 是 /a/b 的父路径
    if (prefix.startsWith(existingPrefix + '/') || existingPrefix === prefix) {
      return true;
    }

    // 检查是否为子路径 - 例如 /a/b 是 /a 的子路径
    if (existingPrefix.startsWith(prefix + '/')) {
      return true;
    }
  }
  return false;
};

export const updateTenant = async (data: { name: string; prefix?: string; description?: string; isActive?: boolean }) => {
  try {
    // 如果提供了前缀，确保前缀以/开头
    if (data.prefix) {
      let prefix = data.prefix;
      if (!prefix.startsWith('/')) {
        prefix = `/${prefix}`;
      }

      // 检查是否为根级别目录
      if (prefix === '/') {
        toast.error(t('tenants.root_prefix_not_allowed'), {
          duration: 3000,
        });
        throw new Error('Root prefix not allowed');
      }

      // 获取当前租户信息
      const currentTenant = await getTenant(data.name);
      
      // 先获取所有租户，检查前缀冲突
      const tenants = await getTenants();
      if (checkPrefixConflict(prefix, tenants.map((t: any) => t.prefix), currentTenant.prefix)) {
        toast.error(t('tenants.prefix_path_conflict'), {
          duration: 3000,
        });
        throw new Error('Prefix path conflict');
      }

      data.prefix = prefix;
    }

    const response = await api.put('/auth/tenants', data);
    toast.success(t('tenants.edit_success'), {
      duration: 3000,
    });
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 409) {
      // 检查具体的错误消息，区分名称冲突和前缀冲突
      const errorMessage = error.response.data?.error;
      if (errorMessage === "Tenant name already exists") {
        toast.error(t('tenants.name_conflict'), {
          duration: 3000,
        });
      } else {
        toast.error(t('tenants.prefix_conflict'), {
          duration: 3000,
        });
      }
    } else if (!(error instanceof Error && 
               (error.message === 'Root prefix not allowed' || 
                error.message === 'Prefix path conflict'))) {
      toast.error(t('errors.update_tenant'), {
        duration: 3000,
      });
    }
    throw error;
  }
};

export const deleteTenant = async (name: string) => {
  try {
    const response = await api.delete(`/auth/tenants/${name}`);
    toast.success(t('tenants.delete_success'), {
      duration: 3000,
    });
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data?.error) {
      toast.error(t('errors.delete_tenant'), {
        duration: 3000,
      });
    } else {
      toast.error(t('tenants.delete_failed'), {
        duration: 3000,
      });
    }
    throw error;
  }
};

// User related API functions
export const getUsers = async () => {
  try {
    const response = await api.get('/auth/users');
    return response.data;
  } catch (error) {
    toast.error(t('errors.fetch_users'), {
      duration: 3000,
    });
    throw error;
  }
};

export const getUser = async (username: string) => {
  try {
    const response = await api.get(`/auth/users/${username}`);
    return response.data;
  } catch (error) {
    toast.error(t('errors.fetch_user'), {
      duration: 3000,
    });
    throw error;
  }
};

export const createUser = async (data: { username: string; password: string; role: 'admin' | 'normal' }) => {
  try {
    const response = await api.post('/auth/users', data);
    toast.success(t('users.add_success'), {
      duration: 3000,
    });
    return response.data;
  } catch (error) {
    toast.error(t('users.add_failed'), {
      duration: 3000,
    });
    throw error;
  }
};

export const updateUser = async (data: { username: string; password?: string; role?: 'admin' | 'normal'; isActive?: boolean }) => {
  try {
    const response = await api.put('/auth/users', data);
    toast.success(t('users.edit_success'), {
      duration: 3000,
    });
    return response.data;
  } catch (error) {
    toast.error(t('users.edit_failed'), {
      duration: 3000,
    });
    throw error;
  }
};

export const deleteUser = async (username: string) => {
  try {
    const response = await api.delete(`/auth/users/${username}`);
    toast.success(t('users.delete_success'), {
      duration: 3000,
    });
    return response.data;
  } catch (error) {
    toast.error(t('users.delete_failed'), {
      duration: 3000,
    });
    throw error;
  }
};

export const toggleUserStatus = async (username: string, isActive: boolean) => {
  try {
    const response = await api.put('/auth/users', {
      username,
      isActive,
    });
    toast.success(isActive ? t('users.enable_success') : t('users.disable_success'), {
      duration: 3000,
    });
    return response.data;
  } catch (error) {
    toast.error(isActive ? t('users.enable_failed') : t('users.disable_failed'), {
      duration: 3000,
    });
    throw error;
  }
};

export default api;
