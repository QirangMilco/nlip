import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export interface Space {
  id: string;
  name: string;
  type: 'public' | 'private';
  ownerId: string;
  maxItems: number;
  retentionDays: number;
  invitedUsers?: Record<string, 'edit' | 'view'>;
  createdAt: string;
  updatedAt: string;
}

interface SpaceState {
  spaces: Space[];
  currentSpace: Space | null;
  loading: boolean;
  error: string | null;
}

const initialState: SpaceState = {
  spaces: [],
  currentSpace: null,
  loading: false,
  error: null,
};

const spaceSlice = createSlice({
  name: 'space',
  initialState,
  reducers: {
    setSpaces: (state, action: PayloadAction<Space[]>) => {
      state.spaces = action.payload;
      state.error = null;
    },
    setCurrentSpace: (state, action: PayloadAction<Space>) => {
      state.currentSpace = action.payload;
      state.error = null;
    },
    addSpace: (state, action: PayloadAction<Space>) => {
      state.spaces.push(action.payload);
      state.error = null;
    },
    updateSpace: (state, action: PayloadAction<Space>) => {
      const index = state.spaces.findIndex(space => space.id === action.payload.id);
      if (index !== -1) {
        state.spaces[index] = action.payload;
      }
      if (state.currentSpace?.id === action.payload.id) {
        state.currentSpace = action.payload;
      }
      state.error = null;
    },
    deleteSpace: (state, action: PayloadAction<string>) => {
      state.spaces = state.spaces.filter(space => space.id !== action.payload);
      if (state.currentSpace?.id === action.payload) {
        state.currentSpace = null;
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
  setSpaces,
  setCurrentSpace,
  addSpace,
  updateSpace,
  deleteSpace,
  setLoading,
  setError,
} = spaceSlice.actions;

export default spaceSlice.reducer; 