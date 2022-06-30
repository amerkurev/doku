import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import axios from 'axios';
import { makeURL } from './util/net';
import { sumBy } from 'lodash/math';
import { getLocalStorage, setLocalStorage } from './util/localstorage';

const initialState = {
  theme: 'light-theme',
  version: '',

  dockerVersion: null,
  dockerVersionStatus: 'idle',

  dockerContainerList: null,
  dockerContainerListStatus: 'idle',

  dockerDiskUsage: null,
  dockerDiskUsageStatus: 'idle',

  dockerLogs: null,
  dockerLogsStatus: 'idle',

  dockerBindMounts: null,
  dockerBindMountsStatus: 'idle',

  totalSizeImages: 0,
  totalSizeContainers: 0,
  totalSizeVolumes: 0,
  totalSizeBindMounts: 0,
  totalSizeLogs: 0,
  totalSizeBuildCache: 0,

  countImages: 0,
  countContainers: 0,
  countVolumes: 0,
  countBindMounts: 0,
  countLogs: 0,
  countBuildCache: 0,
};

export const getVersion = createAsyncThunk('app/getVersion', async () => {
  const response = await axios.get(makeURL('/v0/version'));
  return response.data;
});

export const getDockerVersion = createAsyncThunk('app/getDockerVersion', async () => {
  const response = await axios.get(makeURL('/v0/docker/version'));
  return response.data;
});

export const getDockerContainerList = createAsyncThunk('app/getDockerContainerList', async () => {
  const response = await axios.get(makeURL('/v0/docker/containers'));
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

export const getDockerLogs = createAsyncThunk('app/getDockerLogs', async () => {
  const response = await axios.get(makeURL('/v0/docker/log-size'));
  return response.data;
});

export const getDockerBindMounts = createAsyncThunk('app/getDockerBindMounts', async () => {
  const response = await axios.get(makeURL('/v0/docker/bind-mounts'));
  return response.data;
});

const diskUsageFulfilled = (state, action) => {
  state.dockerDiskUsageStatus = 'idle';
  state.dockerDiskUsage = action.payload;

  const sharedSizes = [];
  state.totalSizeImages = sumBy(action.payload.Images, (x) => {
    const sharedSize = x.SharedSize;
    if (sharedSize) {
      if (sharedSizes.indexOf(sharedSize) > -1) {
        return x.Size - sharedSize;
      } else {
        sharedSizes.push(sharedSize);
      }
    }
    return x.Size;
  });
  state.totalSizeVolumes = sumBy(action.payload.Volumes, (x) => x.UsageData.Size);
  state.totalSizeBuildCache = action.payload.BuilderSize;

  state.countImages = action.payload.Images.length;
  state.countVolumes = action.payload.Volumes.length;
  state.countBuildCache = action.payload.BuildCache.length;
};

export const appSlice = createSlice({
  name: 'app',
  initialState,
  reducers: {
    cleanupApp: (state) => initialState,
    setupTheme: (state) => {
      state.theme = getLocalStorage('theme');
      document.getElementById('body').className = state.theme;
    },
    setDarkTheme: (state) => {
      state.theme = 'dark-theme';
      document.getElementById('body').className = state.theme;
      setLocalStorage('theme', state.theme);
    },
    setLightTheme: (state) => {
      state.theme = 'light-theme';
      document.getElementById('body').className = state.theme;
      setLocalStorage('theme', state.theme);
    },
  },
  extraReducers: {
    [getVersion.fulfilled]: (state, action) => {
      state.version = action.payload.version;
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
    // Docker Container List
    [getDockerContainerList.pending]: (state) => {
      state.dockerContainerListStatus = 'loading';
    },
    [getDockerContainerList.fulfilled]: (state, action) => {
      state.dockerContainerListStatus = 'idle';
      state.dockerContainerList = action.payload;
      state.totalSizeContainers = action.payload.TotalSize;
      state.countContainers = action.payload.Containers.length;
    },
    [getDockerContainerList.rejected]: (state, action) => {
      state.dockerContainerListStatus = 'idle';
    },
    // Docker Disk Usage
    [getDockerDiskUsage.pending]: (state) => {
      state.dockerDiskUsageStatus = 'loading';
    },
    [getDockerDiskUsage.fulfilled]: diskUsageFulfilled,
    [getDockerDiskUsage.rejected]: (state, action) => {
      state.dockerDiskUsageStatus = 'idle';
    },
    // Docker Disk Usage (long polling)
    [getDockerDiskUsageLongPolling.pending]: (state) => {
      state.dockerDiskUsageStatus = 'loading';
    },
    [getDockerDiskUsageLongPolling.fulfilled]: diskUsageFulfilled,
    [getDockerDiskUsageLongPolling.rejected]: (state, action) => {
      state.dockerDiskUsageStatus = 'idle';
    },
    // Docker Log Size
    [getDockerLogs.pending]: (state) => {
      state.dockerLogsStatus = 'loading';
    },
    [getDockerLogs.fulfilled]: (state, action) => {
      state.dockerLogsStatus = 'idle';
      state.dockerLogs = action.payload;
      state.totalSizeLogs = action.payload.TotalSize;
      state.countLogs = action.payload.Logs.length;
    },
    [getDockerLogs.rejected]: (state, action) => {
      state.dockerLogsStatus = 'idle';
    },
    // Docker Bind Mounts
    [getDockerBindMounts.pending]: (state) => {
      state.dockerBindMountsStatus = 'loading';
    },
    [getDockerBindMounts.fulfilled]: (state, action) => {
      state.dockerBindMountsStatus = 'idle';
      state.dockerBindMounts = action.payload;
      state.totalSizeBindMounts = action.payload.TotalSize;
      state.countBindMounts = action.payload.BindMounts.length;
    },
    [getDockerBindMounts.rejected]: (state, action) => {
      state.dockerBindMountsStatus = 'idle';
    },
  },
});

export const appReducer = appSlice.reducer;

export const { cleanupApp, setupTheme, setDarkTheme, setLightTheme } = appSlice.actions;

export const selectIsDarkTheme = (state) => state.app.theme === 'dark-theme';
export const selectVersion = (state) => state.app.version;

export const selectDockerVersion = (state) => state.app.dockerVersion;
export const selectDockerVersionStatus = (state) => state.app.dockerVersionStatus;

export const selectDockerContainerList = (state) => state.app.dockerContainerList;
export const selectDockerContainerListStatus = (state) => state.app.dockerContainerListStatus;

export const selectDockerDiskUsage = (state) => state.app.dockerDiskUsage;
export const selectDockerDiskUsageStatus = (state) => state.app.dockerDiskUsageStatus;

export const selectDockerLogs = (state) => state.app.dockerLogs;
export const selectDockerLogsStatus = (state) => state.app.dockerLogsStatus;

export const selectDockerBindMounts = (state) => state.app.dockerBindMounts;
export const selectDockerBindMountsStatus = (state) => state.app.dockerBindMountsStatus;

export const selectTotalSizeImages = (state) => state.app.totalSizeImages;
export const selectTotalSizeContainers = (state) => state.app.totalSizeContainers;
export const selectTotalSizeVolumes = (state) => state.app.totalSizeVolumes;
export const selectTotalSizeBindMounts = (state) => state.app.totalSizeBindMounts;
export const selectTotalSizeLogs = (state) => state.app.totalSizeLogs;
export const selectTotalSizeBuildCache = (state) => state.app.totalSizeBuildCache;

export const selectCountImages = (state) => state.app.countImages;
export const selectCountContainers = (state) => state.app.countContainers;
export const selectCountVolumes = (state) => state.app.countVolumes;
export const selectCountBindMounts = (state) => state.app.countBindMounts;
export const selectCountLogs = (state) => state.app.countLogs;
export const selectCountBuildCache = (state) => state.app.countBuildCache;
