import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  withCredentials: true,
});

api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

api.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    const originalRequest = error.config;
    
    // Check if error is 401 and not a retry
    if (error.response && error.response.status === 401 && !originalRequest._retry) {
      // Avoid infinite loops for login/refresh requests themselves
      if (originalRequest.url.includes('/auth/login') || originalRequest.url.includes('/auth/refresh')) {
        return Promise.reject(error);
      }
      
      originalRequest._retry = true;
      
      try {
        // Dynamically import store to avoid circular dependency
        const { useAuthStore } = await import('@/stores/auth');
        const authStore = useAuthStore();
        
        const newToken = await authStore.refreshAccessToken();
        
        // Update header
        originalRequest.headers.Authorization = `Bearer ${newToken}`;
        // Retry
        return api(originalRequest);
      } catch (refreshError) {
         // Auto logout is handled in store, but we can enforce redirect here if needed
         window.location.href = '/login';
         return Promise.reject(refreshError);
      }
    }
    return Promise.reject(error);
  }
);

export default api;
