import axios from 'axios';

// Centralized Axios client for the frontend
export const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // For HTTP-only cookies (refresh tokens)
});

// Interceptor architecture scaffold
apiClient.interceptors.request.use((config) => {
  // Inject access tokens if needed
  // Inject x-correlation-id for tracing
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    // Handle 401 token refresh flow here
    return Promise.reject(error);
  }
);
