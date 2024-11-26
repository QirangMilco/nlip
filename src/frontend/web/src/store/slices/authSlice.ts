import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { User, AuthResponse } from '@/store/types';

interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  isInitialized: boolean;
  token: string | null;
}

const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  isInitialized: false,
  token: null,
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
    },
    clearAuth: (state) => {
      state.isAuthenticated = false;
      state.user = null;
      state.isInitialized = true;
      state.token = null;
    },
  },
});

export const { setAuth, clearAuth } = authSlice.actions;
export default authSlice.reducer; 