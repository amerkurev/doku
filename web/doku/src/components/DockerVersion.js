import React from 'react';
import { Segment, Grid } from 'semantic-ui-react';
import { useSelector } from 'react-redux';
import { selectDockerVersion, selectIsDarkTheme } from '../AppSlice';

function DockerVersion() {
  const isDarkTheme = useSelector(selectIsDarkTheme);
  const dockerVersion = useSelector(selectDockerVersion);

  if (dockerVersion === null) {
    return <div />; // initial state
  }

  let platform = 'Docker Version';
  if (dockerVersion.Platform && dockerVersion.Platform.Name) {
    platform = dockerVersion.Platform.Name;
  }

  return (
    <Segment vertical>
      <Grid columns={2} inverted={isDarkTheme}>
        <Grid.Row>
          <Grid.Column>
            {platform}
            {':\u00a0\u00a0'}
            <strong>{dockerVersion.Version}</strong>
            {`\u00a0(${dockerVersion.Os})`}
          </Grid.Column>
          <Grid.Column textAlign="right">
            {'API Version:\u00a0'}
            <strong>{dockerVersion.ApiVersion}</strong>
            {`\u00a0(min.\u00a0${dockerVersion.MinAPIVersion})`}
          </Grid.Column>
        </Grid.Row>
      </Grid>
    </Segment>
  );
}

export default DockerVersion;
