import React, { useReducer } from 'react';
import { Container, Grid, Header, Message, Statistic, Table } from 'semantic-ui-react';
import { useSelector } from 'react-redux';
import {
  selectDockerContainerList,
  selectDockerLogs,
  selectDockerLogsStatus,
  selectTotalSizeLogs,
  selectCountLogs,
  selectIsDarkTheme,
} from '../AppSlice';
import { CHANGE_SORT, sortReducer, sortReducerInitializer } from '../util/sort';
import { sortBy } from 'lodash/collection';
import prettyBytes from 'pretty-bytes';
import statusPage from './StatusPage';
import { prettyContainerID, prettyContainerName, prettyCount, prettyLogPath, replaceWithNbsp } from '../util/fmt';

function Logs() {
  const isDarkTheme = useSelector(selectIsDarkTheme);
  const containerList = useSelector(selectDockerContainerList);
  const logs = useSelector(selectDockerLogs);
  const logsStatus = useSelector(selectDockerLogsStatus);
  const totalSize = useSelector(selectTotalSizeLogs);
  const count = useSelector(selectCountLogs);
  const [state, dispatch] = useReducer(sortReducer, sortReducerInitializer());

  const s = statusPage(logs, logsStatus);
  if (s !== null) {
    return s;
  }

  let dataTable = null;

  if (Array.isArray(logs.Logs) && logs.Logs.length > 0) {
    const { column, direction } = state;
    const data = sortBy(logs.Logs, [column]);
    if (direction === 'descending') {
      data.reverse();
    }

    dataTable = (
      <Table selectable sortable celled compact size="small" inverted={isDarkTheme}>
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
          {data.map(({ ContainerID, ContainerName, Path, Size }) => (
            <Table.Row key={ContainerID}>
              <Table.Cell>
                <small>
                  <code>{prettyContainerID(ContainerID)}</code>
                </small>
              </Table.Cell>
              <Table.Cell>{prettyContainerName(ContainerName)}</Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(Size))}</Table.Cell>
              <Table.Cell>
                <small>{prettyLogPath(Path)}</small>
              </Table.Cell>
            </Table.Row>
          ))}
        </Table.Body>
      </Table>
    );
  }

  const showWarning = containerList && Array.isArray(containerList.Containers) && containerList.Containers.length > 0 && dataTable === null;

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
            <Header inverted={isDarkTheme}>Logs {prettyCount(count)}</Header>
          </Grid.Column>
        </Grid.Row>
      </Grid>
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
