import React from 'react';
import { Grid } from 'semantic-ui-react';
import { replaceWithNbsp } from '../util/fmt';
import prettyBytes from 'pretty-bytes';
import { useSelector } from 'react-redux';
import {
  selectIsDarkTheme,
  selectTotalSizeBuildCache,
  selectTotalSizeContainers,
  selectTotalSizeImages,
  selectTotalSizeLogs,
  selectTotalSizeVolumes,
} from '../AppSlice';

function Statistics() {
  const isDarkTheme = useSelector(selectIsDarkTheme);
  const totalSizeImages = useSelector(selectTotalSizeImages);
  const totalSizeContainers = useSelector(selectTotalSizeContainers);
  const totalSizeVolumes = useSelector(selectTotalSizeVolumes);
  const totalSizeLogs = useSelector(selectTotalSizeLogs);
  const totalSizeBuildCache = useSelector(selectTotalSizeBuildCache);

  return (
    <Grid columns={2} inverted={isDarkTheme}>
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
  );
}

export default Statistics;
