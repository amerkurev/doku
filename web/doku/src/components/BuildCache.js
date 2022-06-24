import React from 'react';
import { Loader, Container, Segment, Statistic } from 'semantic-ui-react';
import { useSelector } from 'react-redux';
import { selectDockerDiskUsage, selectDockerDiskUsageStatus } from '../AppSlice';
import prettyBytes from 'pretty-bytes';

function BuildCache() {
  const diskUsage = useSelector(selectDockerDiskUsage);
  const diskUsageStatus = useSelector(selectDockerDiskUsageStatus);

  if (diskUsageStatus === 'loading' && diskUsage == null) {
    return (
      <Container>
        <Loader active>Loading</Loader>
      </Container>
    );
  } else if (diskUsage === null) {
    return <div />; // initial state
  }

  const builderSize = prettyBytes(diskUsage.BuilderSize);
  return (
    <Container>
      <Statistic>
        <Statistic.Label>Builder Size</Statistic.Label>
        <Statistic.Value>{builderSize}</Statistic.Value>
      </Statistic>
    </Container>
  );
}

export default BuildCache;
