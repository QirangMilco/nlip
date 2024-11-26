import { configureStore } from '@reduxjs/toolkit';
import authReducer from './slices/authSlice';
import spaceReducer from './slices/spaceSlice';
import clipReducer from './slices/clipSlice';

export const store = configureStore({
  reducer: {
    auth: authReducer,
    space: spaceReducer,
    clip: clipReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        ignoredActions: ['auth/setAuth'],
      },
    }),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

// 导出hooks类型
export type AppThunk<ReturnType = void> = ThunkAction<
  ReturnType,
  RootState,
  unknown,
  Action<string>
>; 