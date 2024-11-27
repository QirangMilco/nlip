import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { User, AuthResponse } from '@/store/types';

interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  isInitialized: boolean;
  token: string | null;
  needChangePwd: boolean;
  loading: boolean;
}

const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  isInitialized: false,
  token: null,
  needChangePwd: false,
  loading: false
};

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setAuth: (state, action: PayloadAction<AuthResponse>) => {
      state.isAuthenticated = true;
      state.user = action.payload.user;
      state.isInitialized = true;
      state.token = action.payload.token;
      state.needChangePwd = action.payload.needChangePwd;
      state.loading = false;
    },
    clearAuth: (state) => {
      state.isAuthenticated = false;
      state.user = null;
      state.isInitialized = true;
      state.token = null;
      state.needChangePwd = false;
      state.loading = false;
    },
  },
});

export const { setAuth, clearAuth } = authSlice.actions;
export default authSlice.reducer; 