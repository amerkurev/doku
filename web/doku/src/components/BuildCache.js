import React, { useReducer } from 'react';
import { Container, Statistic, Table, Message, Popup, Icon, Grid, Header } from 'semantic-ui-react';
import { useSelector } from 'react-redux';
import {
  selectCountBuildCache,
  selectDockerDiskUsage,
  selectDockerDiskUsageStatus,
  selectIsDarkTheme,
  selectTotalSizeBuildCache,
} from '../AppSlice';
import prettyBytes from 'pretty-bytes';
import { sortBy } from 'lodash/collection';
import { CHANGE_SORT, sortReducer, sortReducerInitializer } from '../util/sort';
import statusPage from './StatusPage';
import { prettyCount, prettyTime, replaceWithNbsp } from '../util/fmt';

function BuildCache() {
  const isDarkTheme = useSelector(selectIsDarkTheme);
  const diskUsage = useSelector(selectDockerDiskUsage);
  const diskUsageStatus = useSelector(selectDockerDiskUsageStatus);
  const totalSize = useSelector(selectTotalSizeBuildCache);
  const count = useSelector(selectCountBuildCache);
  const [state, dispatch] = useReducer(sortReducer, sortReducerInitializer());

  const s = statusPage(diskUsage, diskUsageStatus);
  if (s !== null) {
    return s;
  }

  let dataTable = null;

  if (Array.isArray(diskUsage.BuildCache) && diskUsage.BuildCache.length > 0) {
    const { column, direction } = state;
    const data = sortBy(diskUsage.BuildCache, [column]);
    if (direction === 'descending') {
      data.reverse();
    }

    dataTable = (
      <Table selectable sortable celled compact size="small" inverted={isDarkTheme}>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell sorted={column === 'ID' ? direction : null} onClick={() => dispatch({ type: CHANGE_SORT, column: 'ID' })}>
              Build Cache ID
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'Size' ? direction : null}
              onClick={() => dispatch({ type: 'CHANGE_SORT', column: 'Size' })}>
              Size
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="center"
              sorted={column === 'UsageCount' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'UsageCount' })}>
              Usage Count
            </Table.HeaderCell>
            <Table.HeaderCell sorted={column === 'Type' ? direction : null} onClick={() => dispatch({ type: CHANGE_SORT, column: 'Type' })}>
              Type
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="center"
              sorted={column === 'Shared' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'Shared' })}>
              Shared
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="center"
              sorted={column === 'InUse' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'InUse' })}>
              In Use
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="center"
              sorted={column === 'LastUsedAt' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'LastUsedAt' })}>
              Last Used At
            </Table.HeaderCell>
            <Table.HeaderCell />
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {data.map(({ ID, Description, Size, UsageCount, InUse, Shared, Type, LastUsedAt, CreatedAt, Parent }) => (
            <Table.Row key={ID}>
              <Table.Cell>
                <small>
                  <code>{ID}</code>
                </small>
              </Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(Size))}</Table.Cell>
              <Table.Cell textAlign="center">{UsageCount}</Table.Cell>
              <Table.Cell>{Type}</Table.Cell>
              <Table.Cell textAlign="center">{Shared ? 'yes' : 'no'}</Table.Cell>
              <Table.Cell textAlign="center">{InUse ? 'yes' : 'no'}</Table.Cell>
              <Table.Cell textAlign="center">{prettyTime(LastUsedAt)}</Table.Cell>
              <Popup
                inverted={isDarkTheme}
                wide="very"
                header="Description"
                content={Description}
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
            <Header>Build Cache {prettyCount(count)}</Header>
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
          <code>{'$ docker builder prune'}</code>
        </Message.Header>
        Remove build cache. For more details, see{' '}
        <a rel="noreferrer" target="_blank" href="https://docs.docker.com/engine/reference/commandline/builder_prune/">
          docker builder prune.
        </a>
      </Message.Content>
    </Message>
  );
}

export default BuildCache;
