import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Clip } from '@/store/types';

interface ClipState {
  clips: Clip[];
  loading: boolean;
  error: string | null;
  currentClip: Clip | null;
}

const initialState: ClipState = {
  clips: [],
  loading: false,
  error: null,
  currentClip: null,
};

const clipSlice = createSlice({
  name: 'clip',
  initialState,
  reducers: {
    setClips: (state, action: PayloadAction<Clip[]>) => {
      state.clips = action.payload;
      state.error = null;
    },
    setCurrentClip: (state, action: PayloadAction<Clip>) => {
      state.currentClip = action.payload;
      state.error = null;
    },
    addClip: (state, action: PayloadAction<Clip>) => {
      state.clips.unshift(action.payload);
      state.error = null;
    },
    deleteClip: (state, action: PayloadAction<string>) => {
      state.clips = state.clips.filter(clip => clip.id !== action.payload);
      if (state.currentClip?.id === action.payload) {
        state.currentClip = null;
      }
      state.error = null;
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
    setError: (state, action: PayloadAction<string>) => {
      state.error = action.payload;
      state.loading = false;
    },
  },
});

export const {
  setClips,
  setCurrentClip,
  addClip,
  deleteClip,
  setLoading,
  setError,
} = clipSlice.actions;

export default clipSlice.reducer; 