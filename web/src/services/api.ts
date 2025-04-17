import axios from 'axios';
import toast from 'react-hot-toast';

// Create an axios instance with default config
const api = axios.create({
  baseURL: '/api', // Assuming the API is served under /api
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// API endpoints
export const getMCPServers = async () => {
  try {
    const response = await api.get('/configs');
    return response.data;
  } catch (error) {
    toast.error('获取 MCP 服务器列表失败', {
      duration: 3000,
      position: 'bottom-right',
    });
    throw error;
  }
};

export default api;
