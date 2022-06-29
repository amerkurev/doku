import React, { useReducer } from 'react';
import { useSelector } from 'react-redux';
import { selectDockerDiskUsage, selectDockerDiskUsageStatus, selectTotalSizeVolumes, selectCountVolumes } from '../AppSlice';
import { CHANGE_SORT, sortReducer, sortReducerInitializer } from '../util/sort';
import statusPage from './StatusPage';
import { sortBy } from 'lodash/collection';
import { Container, Grid, Header, Icon, Message, Popup, Statistic, Table } from 'semantic-ui-react';
import { prettyCount, prettyTime, replaceWithNbsp } from '../util/fmt';
import prettyBytes from 'pretty-bytes';

function Volumes() {
  const diskUsage = useSelector(selectDockerDiskUsage);
  const diskUsageStatus = useSelector(selectDockerDiskUsageStatus);
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
        return { ...x, ...{ RefCount: x.UsageData.RefCount, Size: x.UsageData.Size } };
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
              Name
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'Size' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'Size' })}>
              Size
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'RefCount' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'RefCount' })}>
              Ref.Count
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
          {data.map(({ CreatedAt, Driver, Labels, Mountpoint, Name, Options, Scope, Size, RefCount }) => (
            <Table.Row key={Name}>
              <Table.Cell>
                <small>
                  <code>{Name}</code>
                </small>
              </Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(Size))}</Table.Cell>
              <Table.Cell textAlign="right">{RefCount}</Table.Cell>
              <Table.Cell textAlign="center">{Driver}</Table.Cell>
              <Table.Cell textAlign="center">{Scope}</Table.Cell>
              <Table.Cell textAlign="center">{prettyTime(CreatedAt)}</Table.Cell>
              <Popup
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
            <Statistic>
              <Statistic.Label>Total size</Statistic.Label>
              <Statistic.Value>{replaceWithNbsp(prettyBytes(totalSize))}</Statistic.Value>
            </Statistic>
          </Grid.Column>
          <Grid.Column textAlign="right" verticalAlign="bottom">
            <Header>Volumes {prettyCount(count)}</Header>
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

export default Volumes;
