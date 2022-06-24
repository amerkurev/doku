import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import axios from 'axios';
import { makeURL } from './conf/net';

const initialState = {
  version: '',
};

export const getVersion = createAsyncThunk('app/getVersion', async () => {
  const response = await axios.get(makeURL('/version'));
  return response.data;
});

export const appSlice = createSlice({
  name: 'app',
  initialState,
  reducers: {
    cleanupApp: (state) => initialState,
  },
  extraReducers: {
    [getVersion.fulfilled]: (state, action) => {
      state.version = action.payload.version;
    },
  },
});

export const appReducer = appSlice.reducer;

export const { cleanupApp } = appSlice.actions;

export const selectVersion = (state) => state.app.version;
