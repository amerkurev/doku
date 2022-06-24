import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import axios from 'axios';
import { makeURL } from './conf/net';

const initialState = {
  version: '',
  sizeCalcProgress: null,

  dockerVersion: null,
  dockerVersionStatus: 'idle',

  dockerDiskUsage: null,
  dockerDiskUsageStatus: 'idle',

  dockerLogSize: null,
  dockerLogSizeStatus: 'idle',

  dockerBindMounts: null,
  dockerBindMountsStatus: 'idle',
};

export const getVersion = createAsyncThunk('app/getVersion', async () => {
  const response = await axios.get(makeURL('/v0/version'));
  return response.data;
});

export const getSizeCalcProgress = createAsyncThunk('app/getSizeCalcProgress', async () => {
  const response = await axios.get(makeURL('/v0/size-calc-progress'));
  return response.data;
});

export const getDockerVersion = createAsyncThunk('app/getDockerVersion', async () => {
  const response = await axios.get(makeURL('/v0/docker/version'));
  return response.data;
});

export const getDockerDiskUsage = createAsyncThunk('app/getDockerDiskUsage', async () => {
  const response = await axios.get(makeURL('/v0/docker/disk-usage'));
  return response.data;
});

export const getDockerDiskUsageLongPolling = createAsyncThunk('app/getDockerDiskUsageLongPolling', async () => {
  const response = await axios.get(makeURL('/v0/_/docker/disk-usage'));
  return response.data;
});

export const getDockerLogSize = createAsyncThunk('app/getDockerLogSize', async () => {
  const response = await axios.get(makeURL('/v0/docker/log-size'));
  return response.data;
});

export const getDockerBindMounts = createAsyncThunk('app/getDockerBindMounts', async () => {
  const response = await axios.get(makeURL('/v0/docker/bind-mounts'));
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
    [getSizeCalcProgress.fulfilled]: (state, action) => {
      state.sizeCalcProgress = action.payload.version;
    },
    // Docker Version
    [getDockerVersion.pending]: (state) => {
      state.dockerVersionStatus = 'loading';
    },
    [getDockerVersion.fulfilled]: (state, action) => {
      state.dockerVersionStatus = 'idle';
      state.dockerVersion = action.payload;
    },
    [getDockerVersion.rejected]: (state, action) => {
      state.dockerVersionStatus = 'idle';
    },
    // Docker Disk Usage
    [getDockerDiskUsage.pending]: (state) => {
      state.dockerDiskUsageStatus = 'loading';
    },
    [getDockerDiskUsage.fulfilled]: (state, action) => {
      state.dockerDiskUsageStatus = 'idle';
      state.dockerDiskUsage = action.payload;
    },
    [getDockerDiskUsage.rejected]: (state, action) => {
      state.dockerDiskUsageStatus = 'idle';
    },
    // Docker Disk Usage (long polling)
    [getDockerDiskUsageLongPolling.pending]: (state) => {
      state.dockerDiskUsageStatus = 'loading';
    },
    [getDockerDiskUsageLongPolling.fulfilled]: (state, action) => {
      state.dockerDiskUsageStatus = 'idle';
      state.dockerDiskUsage = action.payload;
    },
    [getDockerDiskUsageLongPolling.rejected]: (state, action) => {
      state.dockerDiskUsageStatus = 'idle';
    },
    // Docker Log Size
    [getDockerLogSize.pending]: (state) => {
      state.dockerLogSizeStatus = 'loading';
    },
    [getDockerLogSize.fulfilled]: (state, action) => {
      state.dockerLogSizeStatus = 'idle';
      state.dockerLogSize = action.payload;
    },
    [getDockerLogSize.rejected]: (state, action) => {
      state.dockerLogSizeStatus = 'idle';
    },
    // Docker Bind Mounts
    [getDockerBindMounts.pending]: (state) => {
      state.dockerBindMountsStatus = 'loading';
    },
    [getDockerBindMounts.fulfilled]: (state, action) => {
      state.dockerBindMountsStatus = 'idle';
      state.dockerBindMounts = action.payload;
    },
    [getDockerBindMounts.rejected]: (state, action) => {
      state.dockerBindMountsStatus = 'idle';
    },
  },
});

export const appReducer = appSlice.reducer;

export const { cleanupApp } = appSlice.actions;

export const selectVersion = (state) => state.app.version;
