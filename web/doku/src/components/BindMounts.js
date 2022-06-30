import React, { useReducer } from 'react';
import { Container, Statistic, Table, Icon, Message, Grid, Header } from 'semantic-ui-react';
import { useSelector } from 'react-redux';
import {
  selectDockerBindMounts,
  selectDockerBindMountsStatus,
  selectTotalSizeBindMounts,
  selectCountBindMounts,
  selectIsDarkTheme,
} from '../AppSlice';
import { CHANGE_SORT, sortReducer, sortReducerInitializer } from '../util/sort';
import statusPage from './StatusPage';
import { sortBy } from 'lodash/collection';
import { prettyCount, prettyTime, replaceWithNbsp } from '../util/fmt';
import prettyBytes from 'pretty-bytes';
import { findIndex } from 'lodash/array';

function BindMounts() {
  const isDarkTheme = useSelector(selectIsDarkTheme);
  const bindMounts = useSelector(selectDockerBindMounts);
  const bindMountsStatus = useSelector(selectDockerBindMountsStatus);
  const totalSize = useSelector(selectTotalSizeBindMounts);
  const count = useSelector(selectCountBindMounts);
  const [state, dispatch] = useReducer(sortReducer, sortReducerInitializer());

  const s = statusPage(bindMounts, bindMountsStatus);
  if (s !== null) {
    return s;
  }

  let dataTable = null;

  if (Array.isArray(bindMounts.BindMounts) && bindMounts.BindMounts.length > 0) {
    const { column, direction } = state;
    const data = sortBy(bindMounts.BindMounts, [column]);
    if (direction === 'descending') {
      data.reverse();
    }

    const customView = (prepared, err, value) => {
      if (!prepared) return <Icon loading name="spinner" />;
      if (err.length > 0) return '-'; // error
      return value;
    };

    dataTable = (
      <Table selectable sortable celled compact size="small" inverted={isDarkTheme}>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell sorted={column === 'Path' ? direction : null} onClick={() => dispatch({ type: CHANGE_SORT, column: 'Path' })}>
              Path
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'Size' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'Size' })}>
              Size
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'Files' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'Files' })}>
              Files
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="center"
              sorted={column === 'ReadOnly' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'ReadOnly' })}>
              Read Only
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="center"
              sorted={column === 'LastCheck' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'LastCheck' })}>
              Last Check
            </Table.HeaderCell>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {data.map(({ Path, Size, IsDir, Files, ReadOnly, LastCheck, Prepared, Err }) => (
            <Table.Row key={Path}>
              <Table.Cell>{Path}</Table.Cell>
              <Table.Cell textAlign="right">{customView(Prepared, Err, replaceWithNbsp(prettyBytes(Size)))}</Table.Cell>
              <Table.Cell textAlign="right">{customView(Prepared, Err, IsDir ? Files : 1)}</Table.Cell>
              <Table.Cell textAlign="center">{ReadOnly ? 'yes' : 'no'}</Table.Cell>
              <Table.Cell textAlign="center">{prettyTime(LastCheck)}</Table.Cell>
            </Table.Row>
          ))}
        </Table.Body>
      </Table>
    );
  }

  const showWarning = bindMounts.BindMounts && Array.isArray(bindMounts.BindMounts) && findIndex(bindMounts.BindMounts, (m) => m.Err) > -1;

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
            <Header>Bind Mounts {prettyCount(count)}</Header>
          </Grid.Column>
        </Grid.Row>
      </Grid>
      {showWarning ? <NoAccessWarning /> : null}
      <HelpText />
      {dataTable}
    </Container>
  );
}

function NoAccessWarning() {
  return (
    <Message warning size="tiny">
      <Message.Content>
        <Message.Header>
          <code>{'No access to some mounted files or directories'}</code>
        </Message.Header>
        {"Doku doesn't have access to some mounted files or directories and can't calculate the size of these files."}
      </Message.Content>
    </Message>
  );
}

function HelpText() {
  return (
    <Message info size="tiny">
      <Message.Content>
        <Message.Header>
          <Icon name="info circle" />
          <code>{'Note'}</code>
        </Message.Header>
        When you use a bind mount, a file or directory on the host machine is mounted into a container. The file or directory is referenced
        by its absolute path on the host machine. In this case, disk space usage is not directly related to Docker. For more details, see{' '}
        <a rel="noreferrer" target="_blank" href="https://docs.docker.com/storage/bind-mounts/">
          bind mounts.
        </a>
      </Message.Content>
    </Message>
  );
}

export default BindMounts;
