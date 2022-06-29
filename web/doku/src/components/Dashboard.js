import React, { useReducer } from 'react';
import { Container, Grid, Statistic, Table, Icon, Popup } from 'semantic-ui-react';
import PieChart from './PieChart';
import { prettyContainerName, prettyTime, replaceWithNbsp } from '../util/fmt';
import prettyBytes from 'pretty-bytes';
import { useSelector } from 'react-redux';
import {
  selectDockerContainerList,
  selectDockerContainerListStatus,
  selectDockerDiskUsage,
  selectDockerLogs,
  selectTotalSizeBuildCache,
  selectTotalSizeContainers,
  selectTotalSizeImages,
  selectTotalSizeLogs,
  selectTotalSizeVolumes,
} from '../AppSlice';
import Statistics from './Statistics';
import { CHANGE_SORT, sortReducer, sortReducerInitializer } from '../util/sort';
import statusPage from './StatusPage';
import { sortBy } from 'lodash/collection';

function getImageSize(diskUsage, imageName) {
  if (diskUsage && Array.isArray(diskUsage.Images) && diskUsage.Images.length > 0) {
    for (let i = 0; i < diskUsage.Images.length; i++) {
      const x = diskUsage.Images[i];
      if (Array.isArray(x.RepoTags) && x.RepoTags.indexOf(imageName) > -1) {
        return x.Size;
      }
    }
  }
  return 0;
}

function getVolumesSize(diskUsage, mounts) {
  let size = 0;
  if (diskUsage && Array.isArray(mounts) && mounts.length > 0) {
    const volumes = mounts.map((x) => x.Name);

    if (Array.isArray(diskUsage.Volumes) && diskUsage.Volumes.length > 0) {
      for (let i = 0; i < diskUsage.Volumes.length; i++) {
        const x = diskUsage.Volumes[i];
        if (volumes.indexOf(x.Name) > -1) {
          size += x.UsageData.Size;
        }
      }
    }
  }
  return size;
}

function getLogSize(logs, containerId) {
  if (logs && Array.isArray(logs.Logs) && logs.Logs.length > 0) {
    for (let i = 0; i < logs.Logs.length; i++) {
      const x = logs.Logs[i];
      if (x.ContainerID === containerId) {
        return x.Size;
      }
    }
  }
  return 0;
}

function Dashboard() {
  const containerList = useSelector(selectDockerContainerList);
  const containerListStatus = useSelector(selectDockerContainerListStatus);
  const diskUsage = useSelector(selectDockerDiskUsage);
  const logs = useSelector(selectDockerLogs);
  const [state, dispatch] = useReducer(sortReducer, sortReducerInitializer());

  const totalSizeImages = useSelector(selectTotalSizeImages);
  const totalSizeContainers = useSelector(selectTotalSizeContainers);
  const totalSizeVolumes = useSelector(selectTotalSizeVolumes);
  const totalSizeLogs = useSelector(selectTotalSizeLogs);
  const totalSizeBuildCache = useSelector(selectTotalSizeBuildCache);
  const totalSize = totalSizeImages + totalSizeContainers + totalSizeVolumes + totalSizeLogs + totalSizeBuildCache;

  const s = statusPage(containerList, containerListStatus);
  if (s !== null) {
    return s;
  }

  let dataTable = null;

  if (Array.isArray(containerList.Containers) && containerList.Containers.length > 0) {
    const { column, direction } = state;
    const data = sortBy(
      containerList.Containers.map((x) => {
        const computed = {
          ImageSize: getImageSize(diskUsage, x.Image),
          VolumesSize: getVolumesSize(diskUsage, x.Mounts),
          LogsSize: getLogSize(logs, x.ID),
        };
        return { ...x, ...{ Status: x.State.Status, ID: x.Id }, ...computed };
      }),
      [column]
    );
    if (direction === 'descending') {
      data.reverse();
    }

    dataTable = (
      <Table selectable sortable celled compact size="small">
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell sorted={column === 'Name' ? direction : null} onClick={() => dispatch({ type: CHANGE_SORT, column: 'Name' })}>
              Container
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'SizeRw' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'SizeRw' })}>
              {'Size RW '}
              <Popup
                wide="very"
                header="Size RW"
                content={'The size of files that have been created or changed by this container'}
                trigger={<Icon name="question circle outline" />}
              />
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'ImageSize' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'ImageSize' })}>
              Image Size
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'VolumesSize' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'VolumesSize' })}>
              Volumes Size
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'LogsSize' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'LogsSize' })}>
              Logs Size
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="center"
              sorted={column === 'Status' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'Status' })}>
              Status
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="center"
              sorted={column === 'Created' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'Created' })}>
              Created
            </Table.HeaderCell>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {data.map(({ ID, Name, Image, ImageSize, VolumesSize, LogsSize, Created, SizeRw, Status }) => (
            <Table.Row key={ID}>
              <Table.Cell style={{ whiteSpace: 'pre-line' }}>{prettyContainerName(Name)}</Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(SizeRw))}</Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(ImageSize))}</Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(VolumesSize))}</Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(LogsSize))}</Table.Cell>
              <Table.Cell textAlign="center">{Status}</Table.Cell>
              <Table.Cell textAlign="center">{prettyTime(Created)}</Table.Cell>
            </Table.Row>
          ))}
        </Table.Body>
      </Table>
    );
  }

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
            <Statistics />
          </Grid.Column>
        </Grid.Row>
      </Grid>
      <br />
      {dataTable}
    </Container>
  );
}

export default Dashboard;
