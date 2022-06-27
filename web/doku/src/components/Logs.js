import React, { useReducer } from 'react';
import { Container, Message, Statistic, Table } from 'semantic-ui-react';
import { useSelector } from 'react-redux';
import { selectDockerDiskUsage, selectDockerLogSize, selectDockerLogSizeStatus } from '../AppSlice';
import { CHANGE_SORT, sortReducer, sortReducerInitializer } from '../util/sort';
import { sortBy } from 'lodash/collection';
import prettyBytes from 'pretty-bytes';
import statusPage from './StatusPage';
import { replaceWithNbsp } from '../util/fmt';

function Logs() {
  const diskUsage = useSelector(selectDockerDiskUsage);
  const logSize = useSelector(selectDockerLogSize);
  const logSizeStatus = useSelector(selectDockerLogSizeStatus);
  const [state, dispatch] = useReducer(sortReducer, sortReducerInitializer());

  const s = statusPage(logSize, logSizeStatus);
  if (s !== null) {
    return s;
  }

  let dataTable = null;

  if (Array.isArray(logSize.Logs) && logSize.Logs.length > 0) {
    const { column, direction } = state;
    const data = sortBy(logSize.Logs, [column]);
    if (direction === 'descending') {
      data.reverse();
    }
    dataTable = (
      <Table selectable sortable celled compact size="small">
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell
              sorted={column === 'ContainerID' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'ContainerID' })}>
              Container ID
            </Table.HeaderCell>
            <Table.HeaderCell
              sorted={column === 'ContainerName' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'ContainerName' })}>
              Container Name
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'Size' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'Size' })}>
              Size
            </Table.HeaderCell>
            <Table.HeaderCell>Log File</Table.HeaderCell>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {data.map(({ ContainerID, ContainerName, Path, Size, LastCheck }) => (
            <Table.Row key={ContainerID}>
              <Table.Cell>
                <small>
                  <code>{ContainerID.slice(0, 12)}</code>
                </small>
              </Table.Cell>
              <Table.Cell>{ContainerName}</Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(Size))}</Table.Cell>
              <Table.Cell>
                <small>{Path}</small>
              </Table.Cell>
            </Table.Row>
          ))}
        </Table.Body>
      </Table>
    );
  }

  const showWarning = diskUsage && Array.isArray(diskUsage.Containers) && diskUsage.Containers.length > 0 && dataTable === null;

  return (
    <Container>
      <Statistic>
        <Statistic.Label>Size of logs</Statistic.Label>
        <Statistic.Value>{replaceWithNbsp(prettyBytes(logSize.TotalSize))}</Statistic.Value>
      </Statistic>
      {showWarning ? <NoAccessWarning /> : null}
      {dataTable}
    </Container>
  );
}

function NoAccessWarning() {
  return (
    <Message warning size="tiny">
      <Message.Content>
        <Message.Header>
          <code>{'No access to the log files'}</code>
        </Message.Header>
        {'Although log files of the containers are present, the top-level directory (/) ' +
          'on the host machine has not been mounted into the Doku container.'}
        <br />
        {"Therefore Doku doesn't have access to log files and can't calculate the size of these files."}
      </Message.Content>
    </Message>
  );
}

export default Logs;
