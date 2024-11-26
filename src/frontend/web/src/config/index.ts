import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:3000/api/v1/nlip',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  }
});

export { api }; 