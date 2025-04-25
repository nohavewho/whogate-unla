import axios from 'axios';
import toast from 'react-hot-toast';

// Create an axios instance with default config
const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 10000,
});

// API endpoints
export const getMCPServers = async () => {
  try {
    const response = await api.get('/mcp-servers');
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.data?.error) {
      toast.error(error.response.data.error, {
        duration: 3000,
        position: 'bottom-right',
      });
    } else {
      toast.error('获取 MCP 服务器列表失败', {
        duration: 3000,
        position: 'bottom-right',
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
      toast.error(error.response.data.error, {
        duration: 3000,
        position: 'bottom-right',
      });
    } else {
      toast.error('获取 MCP 服务器失败', {
        duration: 3000,
        position: 'bottom-right',
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
      toast.error(error.response.data.error, {
        duration: 3000,
        position: 'bottom-right',
      });
    } else {
      toast.error('创建 MCP 服务器失败', {
        duration: 3000,
        position: 'bottom-right',
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
      toast.error(error.response.data.error, {
        duration: 3000,
        position: 'bottom-right',
      });
    } else {
      toast.error('更新 MCP 服务器配置失败', {
        duration: 3000,
        position: 'bottom-right',
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
      toast.error(error.response.data.error, {
        duration: 3000,
        position: 'bottom-right',
      });
    } else {
      toast.error('删除 MCP 服务器失败', {
        duration: 3000,
        position: 'bottom-right',
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
      toast.error(error.response.data.error, {
        duration: 3000,
        position: 'bottom-right',
      });
    } else {
      toast.error('同步 MCP 服务器失败', {
        duration: 3000,
        position: 'bottom-right',
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
      toast.error(error.response.data.error, {
        duration: 3000,
        position: 'bottom-right',
      });
    } else {
      toast.error('获取聊天消息失败', {
        duration: 3000,
        position: 'bottom-right',
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
      toast.error(error.response.data.error, {
        duration: 3000,
        position: 'bottom-right',
      });
    } else {
      toast.error('获取聊天会话列表失败', {
        duration: 3000,
        position: 'bottom-right',
      });
    }
    throw error;
  }
};

export default api;
