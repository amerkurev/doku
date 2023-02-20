import React, { useEffect, useLayoutEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { Container } from 'semantic-ui-react';

import Dashboard from './components/Dashboard';
import Images from './components/Images';
import Containers from './components/Containers';
import Volumes from './components/Volumes';
import Logs from './components/Logs';
import BindMounts from './components/BindMounts';
import BuildCache from './components/BuildCache';
import TopMenu from './components/TopMenu';
import Footer from './components/Footer';
import {
  getVersion,
  getDiskUsage,
  getDockerVersion,
  getDockerDiskUsage,
  getDockerLogs,
  getDockerBindMounts,
  getDockerContainerList,
  setupTheme,
} from './AppSlice';

function App() {
  const dispatch = useDispatch();

  useLayoutEffect(() => {
    dispatch(setupTheme());
  });

  useEffect(() => {
    dispatch(getVersion());
    dispatch(getDiskUsage());
    dispatch(getDockerVersion());
    dispatch(getDockerDiskUsage());
    dispatch(getDockerContainerList());
    dispatch(getDockerLogs());
    dispatch(getDockerBindMounts());
  }, [dispatch]);

  return (
    <BrowserRouter basename="/">
      <TopMenu />
      <Container style={{ paddingTop: '6em', paddingBottom: '10em' }}>
        <Routes>
          <Route path="/" element={<Dashboard />} exact />
          <Route path="/images" element={<Images />} exact />
          <Route path="/containers" element={<Containers />} exact />
          <Route path="/volumes" element={<Volumes />} exact />
          <Route path="/bind-mounts" element={<BindMounts />} exact />
          <Route path="/logs" element={<Logs />} exact />
          <Route path="/build-cache" element={<BuildCache />} exact />
        </Routes>
      </Container>
      <Footer />
    </BrowserRouter>
  );
}

export default App;
