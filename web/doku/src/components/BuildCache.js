import React, { useReducer } from 'react';
import { Loader, Container, Statistic, Table, Message, Popup, Icon } from 'semantic-ui-react';
import { useSelector } from 'react-redux';
import { selectDockerDiskUsage, selectDockerDiskUsageStatus } from '../AppSlice';
import prettyBytes from 'pretty-bytes';
import moment from 'moment';
import { sortBy } from 'lodash/collection';
import { CHANGE_SORT, ASC, DESC } from '../conf/constants';

function tableSortReducer(state, action) {
  switch (action.type) {
    case CHANGE_SORT:
      if (state.column === action.column) {
        return {
          ...state,
          direction: state.direction === ASC ? DESC : ASC,
        };
      }

      return {
        column: action.column,
        direction: ASC,
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
    direction: DESC,
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
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'ID' })}>
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
            <Table.HeaderCell
              width={1}
              sorted={column === 'Type' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'Type' })}>
              Type
            </Table.HeaderCell>
            <Table.HeaderCell
              width={1}
              textAlign="center"
              sorted={column === 'Shared' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'Shared' })}>
              Shared
            </Table.HeaderCell>
            <Table.HeaderCell
              width={1}
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
              <Table.Cell textAlign="center">
                <small>
                  <code>{ID}</code>
                </small>
              </Table.Cell>
              <Table.Cell textAlign="right">{prettyBytes(Size)}</Table.Cell>
              <Table.Cell textAlign="center">{UsageCount}</Table.Cell>
              <Table.Cell>{Type}</Table.Cell>
              <Table.Cell textAlign="center">{Shared ? 'yes' : 'no'}</Table.Cell>
              <Table.Cell textAlign="center">{InUse ? 'yes' : 'no'}</Table.Cell>
              <Table.Cell textAlign="center">{moment(LastUsedAt).format('YYYY-MM-DD\u00a0\u00a0HH:mm:ss Z')}</Table.Cell>
              <Popup
                wide="very"
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
      <Statistic>
        <Statistic.Label>Builder Size</Statistic.Label>
        <Statistic.Value>{prettyBytes(diskUsage.BuilderSize)}</Statistic.Value>
      </Statistic>
      <Message success size="tiny">
        <Message.Content>
          <Message.Header>
            <code>{'$ docker builder prune'}</code>
          </Message.Header>
          Remove build cache. See details of{' '}
          <a target="_blank" href="https://docs.docker.com/engine/reference/commandline/builder_prune/">
            docker builder prune
          </a>
        </Message.Content>
      </Message>
      {buildCacheTable}
    </Container>
  );
}

export default BuildCache;
