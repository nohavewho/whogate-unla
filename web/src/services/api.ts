import axios from 'axios';

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
    console.error('Error fetching MCP servers:', error);
    throw error;
  }
};

export default api;
