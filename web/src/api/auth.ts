import api from '../services/api';

export const getCurrentUser = async () => {
  try {
    const response = await api.get('/auth/user/info');
    console.log('API response:', response);
    return response;
  } catch (error) {
    console.error('API error:', error);
    throw error;
  }
}; 