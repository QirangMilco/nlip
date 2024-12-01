import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { User } from '../types';

export interface AuthState {
  token: string | null;
  user: User | null;
  needChangePwd: boolean;
  isLoading: boolean;
  error: string | null;
}

interface SetAuthPayload {
  token: string;
  user: User;
  needChangePwd: boolean;
}

const initialState: AuthState = {
  token: null,
  user: null,
  needChangePwd: false,
  isLoading: false,
  error: null
};

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setAuth: (state, action: PayloadAction<SetAuthPayload>) => {
      state.token = action.payload.token;
      state.user = action.payload.user;
      state.needChangePwd = action.payload.needChangePwd;
    },
    setNeedChangePwd: (state, action: PayloadAction<boolean>) => {
      state.needChangePwd = action.payload;
    },
    clearAuth: (state) => {
      state.token = null;
      state.user = null;
      state.needChangePwd = false;
    }
  }
});

export const { setAuth, setNeedChangePwd, clearAuth } = authSlice.actions;
export default authSlice.reducer; 