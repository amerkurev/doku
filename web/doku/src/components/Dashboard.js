import React from 'react';
import { Container, Grid, Segment, Statistic } from 'semantic-ui-react';
import PieChart from './PieChart';
import { replaceWithNbsp } from '../util/fmt';
import prettyBytes from 'pretty-bytes';
import { useSelector } from 'react-redux';
import {
  selectTotalSizeBuildCache,
  selectTotalSizeContainers,
  selectTotalSizeImages,
  selectTotalSizeLogs,
  selectTotalSizeVolumes,
} from '../AppSlice';

function Dashboard() {
  const totalSizeImages = useSelector(selectTotalSizeImages);
  const totalSizeContainers = useSelector(selectTotalSizeContainers);
  const totalSizeVolumes = useSelector(selectTotalSizeVolumes);
  const totalSizeLogs = useSelector(selectTotalSizeLogs);
  const totalSizeBuildCache = useSelector(selectTotalSizeBuildCache);
  const totalSize = totalSizeImages + totalSizeContainers + totalSizeVolumes + totalSizeLogs + totalSizeBuildCache;

  return (
    <Container>
      <Grid columns={2}>
        <Grid.Row>
          <Grid.Column textAlign="right">
            <PieChart />
          </Grid.Column>
          <Grid.Column textAlign="center">
            <Container style={{ marginTop: '60px' }}>
              <Statistic>
                <Statistic.Label>Docker disk space usage</Statistic.Label>
                <Statistic.Value>{replaceWithNbsp(prettyBytes(totalSize))}</Statistic.Value>
              </Statistic>
            </Container>
            <Grid columns={2}>
              <Grid.Row />
              <Grid.Row>
                <Grid.Column textAlign="right">
                  <h4>{replaceWithNbsp(prettyBytes(totalSizeImages))}</h4>
                </Grid.Column>
                <Grid.Column textAlign="left">
                  <h4>Images</h4>
                </Grid.Column>
              </Grid.Row>
              <Grid.Row>
                <Grid.Column textAlign="right">
                  <h4>{replaceWithNbsp(prettyBytes(totalSizeContainers))}</h4>
                </Grid.Column>
                <Grid.Column textAlign="left">
                  <h4>Containers</h4>
                </Grid.Column>
              </Grid.Row>
              <Grid.Row>
                <Grid.Column textAlign="right">
                  <h4>{replaceWithNbsp(prettyBytes(totalSizeVolumes))}</h4>
                </Grid.Column>
                <Grid.Column textAlign="left">
                  <h4>Volumes</h4>
                </Grid.Column>
              </Grid.Row>
              <Grid.Row>
                <Grid.Column textAlign="right">
                  <h4>{replaceWithNbsp(prettyBytes(totalSizeLogs))}</h4>
                </Grid.Column>
                <Grid.Column textAlign="left">
                  <h4>Logs</h4>
                </Grid.Column>
              </Grid.Row>
              <Grid.Row>
                <Grid.Column textAlign="right">
                  <h4>{replaceWithNbsp(prettyBytes(totalSizeBuildCache))}</h4>
                </Grid.Column>
                <Grid.Column textAlign="left">
                  <h4>Build Cache</h4>
                </Grid.Column>
              </Grid.Row>
            </Grid>
          </Grid.Column>
        </Grid.Row>
      </Grid>
    </Container>
  );
}

export default Dashboard;
