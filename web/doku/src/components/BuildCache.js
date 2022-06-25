import React, { useReducer } from 'react';
import { Loader, Container, Statistic, Table, Grid, Popup } from 'semantic-ui-react';
import { useSelector } from 'react-redux';
import { selectDockerDiskUsage, selectDockerDiskUsageStatus } from '../AppSlice';
import prettyBytes from 'pretty-bytes';
import moment from 'moment';
import { sortBy } from 'lodash/collection';

function tableSortReducer(state, action) {
  switch (action.type) {
    case 'CHANGE_SORT':
      if (state.column === action.column) {
        return {
          ...state,
          // data: state.data.slice().reverse(),
          direction: state.direction === 'ascending' ? 'descending' : 'ascending',
        };
      }

      return {
        column: action.column,
        // data: sortBy(state.data, [action.column]),
        direction: 'ascending',
      };
    default:
      throw new Error();
  }
}

function BuildCache() {
  const diskUsage = useSelector(selectDockerDiskUsage);
  const diskUsageStatus = useSelector(selectDockerDiskUsageStatus);

  const [state, dispatch] = useReducer(tableSortReducer, {
    column: 'Size',
    direction: 'descending',
  });

  if (diskUsageStatus === 'loading' && diskUsage == null) {
    return (
      <Container>
        <Loader active>Loading</Loader>
      </Container>
    );
  } else if (diskUsage === null) {
    return <div />; // initial state
  }

  let buildCacheTable = null;
  if (Array.isArray(diskUsage.BuildCache) && diskUsage.BuildCache.length > 0) {
    const { column, direction } = state;
    const data = sortBy(diskUsage.BuildCache, [column]);
    if (direction === 'descending') {
      data.reverse();
    }

    buildCacheTable = (
      <Table selectable sortable celled compact size="small">
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell
              width={1}
              sorted={column === 'ID' ? direction : null}
              onClick={() => dispatch({ type: 'CHANGE_SORT', column: 'ID' })}>
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
              onClick={() => dispatch({ type: 'CHANGE_SORT', column: 'UsageCount' })}>
              Usage count
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="center"
              sorted={column === 'Type' ? direction : null}
              onClick={() => dispatch({ type: 'CHANGE_SORT', column: 'Type' })}>
              Info
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="center"
              sorted={column === 'LastUsedAt' ? direction : null}
              onClick={() => dispatch({ type: 'CHANGE_SORT', column: 'LastUsedAt' })}>
              Last used at
            </Table.HeaderCell>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {data.map(({ ID, Description, Size, UsageCount, InUse, Shared, Type, LastUsedAt, CreatedAt, Parent }) => (
            <Table.Row key={ID}>
              <Popup
                wide="very"
                content={Description}
                trigger={
                  <Table.Cell textAlign="center">
                    <code>{ID}</code>
                  </Table.Cell>
                }
              />
              <Table.Cell textAlign="right">{prettyBytes(Size)}</Table.Cell>
              <Table.Cell textAlign="center">{UsageCount}</Table.Cell>
              <Table.Cell>
                <Grid textAlign="center" columns={3}>
                  <Grid.Row>
                    <Grid.Column>type={Type}</Grid.Column>
                    <Grid.Column>shared={Shared ? 'yes' : 'no'}</Grid.Column>
                    <Grid.Column>in_use={InUse ? 'yes' : 'no'}</Grid.Column>
                  </Grid.Row>
                </Grid>
              </Table.Cell>
              <Table.Cell textAlign="center">{moment(LastUsedAt).format('YYYY-MM-DD\u00a0\u00a0HH:mm:ss Z')}</Table.Cell>
            </Table.Row>
          ))}
        </Table.Body>
      </Table>
    );
  }

  return (
    <Container>
      <Statistic>
        <Statistic.Label>Builder Size</Statistic.Label>
        <Statistic.Value>{prettyBytes(diskUsage.BuilderSize)}</Statistic.Value>
      </Statistic>
      {buildCacheTable}
    </Container>
  );
}

export default BuildCache;
