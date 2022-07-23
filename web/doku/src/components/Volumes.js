import React, { useReducer } from 'react';
import { useSelector } from 'react-redux';
import {
  selectDockerDiskUsage,
  selectDockerDiskUsageStatus,
  selectTotalSizeVolumes,
  selectCountVolumes,
  selectDockerContainerList,
  selectIsDarkTheme,
} from '../AppSlice';
import { CHANGE_SORT, sortReducer, sortReducerInitializer } from '../util/sort';
import statusPage from './StatusPage';
import { sortBy } from 'lodash/collection';
import { Container, Grid, Header, Icon, Message, Popup, Statistic, Table } from 'semantic-ui-react';
import { prettyContainerName, prettyCount, prettyTime, replaceWithNbsp } from '../util/fmt';
import prettyBytes from 'pretty-bytes';

function Volumes() {
  const isDarkTheme = useSelector(selectIsDarkTheme);
  const diskUsage = useSelector(selectDockerDiskUsage);
  const diskUsageStatus = useSelector(selectDockerDiskUsageStatus);
  const containerList = useSelector(selectDockerContainerList);
  const totalSize = useSelector(selectTotalSizeVolumes);
  const count = useSelector(selectCountVolumes);
  const [state, dispatch] = useReducer(sortReducer, sortReducerInitializer());

  const s = statusPage(diskUsage, diskUsageStatus);
  if (s !== null) {
    return s;
  }

  let dataTable = null;

  if (Array.isArray(diskUsage.Volumes) && diskUsage.Volumes.length > 0) {
    const { column, direction } = state;
    const data = sortBy(
      diskUsage.Volumes.map((x) => {
        const containers = getContainers(containerList, x.Name);
        const extra = {
          Containers: containers.length === 0 ? '-' : containers.join('\n'),
          ContainersNum: containers.length,
          RefCount: x.UsageData.RefCount,
          Size: x.UsageData.Size,
        };
        return { ...x, ...extra };
      }),
      [column]
    );
    if (direction === 'descending') {
      data.reverse();
    }

    dataTable = (
      <Table selectable sortable celled compact size="small" inverted={isDarkTheme}>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell sorted={column === 'Name' ? direction : null} onClick={() => dispatch({ type: CHANGE_SORT, column: 'Name' })}>
              Name
            </Table.HeaderCell>
            <Table.HeaderCell
              sorted={column === 'ContainersNum' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'ContainersNum' })}>
              Containers
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'Size' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'Size' })}>
              Size
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="center"
              sorted={column === 'Driver' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'Driver' })}>
              Driver
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="center"
              sorted={column === 'Scope' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'Scope' })}>
              Scope
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="center"
              sorted={column === 'CreatedAt' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'CreatedAt' })}>
              Created At
            </Table.HeaderCell>
            <Table.HeaderCell />
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {data.map(({ CreatedAt, Driver, Containers, Mountpoint, Name, Scope, Size }) => (
            <Table.Row key={Name}>
              <Table.Cell>
                <small>
                  <code>{Name}</code>
                </small>
              </Table.Cell>
              <Table.Cell style={{ whiteSpace: 'pre-line' }}>{Containers}</Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(Size))}</Table.Cell>
              <Table.Cell textAlign="center">{Driver}</Table.Cell>
              <Table.Cell textAlign="center">{Scope}</Table.Cell>
              <Table.Cell textAlign="center">{prettyTime(CreatedAt)}</Table.Cell>
              <Popup
                inverted={isDarkTheme}
                wide="very"
                header="Mountpoint"
                content={Mountpoint}
                trigger={
                  <Table.Cell textAlign="center">
                    <Icon name="question circle outline" />
                  </Table.Cell>
                }
              />
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
          <Grid.Column>
            <Statistic inverted={isDarkTheme}>
              <Statistic.Label>Total size</Statistic.Label>
              <Statistic.Value>{replaceWithNbsp(prettyBytes(totalSize))}</Statistic.Value>
            </Statistic>
          </Grid.Column>
          <Grid.Column textAlign="right" verticalAlign="bottom">
            <Header inverted={isDarkTheme}>Volumes {prettyCount(count)}</Header>
          </Grid.Column>
        </Grid.Row>
      </Grid>
      <HelpText />
      {dataTable}
    </Container>
  );
}

function HelpText() {
  return (
    <Message success size="tiny">
      <Message.Content>
        <Message.Header>
          <code>{'$ docker volume prune'}</code>
        </Message.Header>
        Remove all unused local volumes. For more details, see{' '}
        <a rel="noreferrer" target="_blank" href="https://docs.docker.com/engine/reference/commandline/volume_prune/">
          docker volume prune.
        </a>
      </Message.Content>
    </Message>
  );
}

function getContainers(containers, volumeName) {
  const res = [];
  if (containers && Array.isArray(containers.Containers) && containers.Containers.length > 0) {
    for (let i = 0; i < containers.Containers.length; i++) {
      const x = containers.Containers[i];
      const volumes = x.Mounts.map((m) => m.Name);
      if (volumes.indexOf(volumeName) > -1) {
        res.push(prettyContainerName(x.Name));
      }
    }
  }
  return res;
}

export default Volumes;
