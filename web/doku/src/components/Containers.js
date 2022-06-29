import React, { useReducer } from 'react';
import { useSelector } from 'react-redux';
import { selectDockerContainerList, selectDockerContainerListStatus, selectTotalSizeContainers, selectCountContainers } from '../AppSlice';
import { CHANGE_SORT, sortReducer, sortReducerInitializer } from '../util/sort';
import statusPage from './StatusPage';
import { Container, Grid, Header, Message, Statistic, Table } from 'semantic-ui-react';
import { prettyContainerID, prettyContainerName, prettyCount, prettyTime, replaceWithNbsp } from '../util/fmt';
import prettyBytes from 'pretty-bytes';
import { sortBy } from 'lodash/collection';

function Containers() {
  const containerList = useSelector(selectDockerContainerList);
  const containerListStatus = useSelector(selectDockerContainerListStatus);
  const totalSize = useSelector(selectTotalSizeContainers);
  const count = useSelector(selectCountContainers);
  const [state, dispatch] = useReducer(sortReducer, sortReducerInitializer());

  const s = statusPage(containerList, containerListStatus);
  if (s !== null) {
    return s;
  }

  let dataTable = null;

  if (Array.isArray(containerList.Containers) && containerList.Containers.length > 0) {
    const { column, direction } = state;
    const data = sortBy(
      containerList.Containers.map((x) => {
        return { ...x, ...{ Status: x.State.Status, ID: x.Id } };
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
            <Table.HeaderCell sorted={column === 'ID' ? direction : null} onClick={() => dispatch({ type: CHANGE_SORT, column: 'ID' })}>
              ID
            </Table.HeaderCell>
            <Table.HeaderCell sorted={column === 'Name' ? direction : null} onClick={() => dispatch({ type: CHANGE_SORT, column: 'Name' })}>
              Name
            </Table.HeaderCell>
            <Table.HeaderCell
              sorted={column === 'ImageName' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'ImageName' })}>
              Image
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'SizeRw' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'SizeRw' })}>
              Size
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'SizeRootFs' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'SizeRootFs' })}>
              Virtual Size
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
          {data.map(({ ID, Name, Image, Command, Created, SizeRw, SizeRootFs, Status }) => (
            <Table.Row key={ID}>
              <Table.Cell>
                <small>
                  <code>{prettyContainerID(ID)}</code>
                </small>
              </Table.Cell>
              <Table.Cell style={{ whiteSpace: 'pre-line' }}>{prettyContainerName(Name)}</Table.Cell>
              <Table.Cell>{Image}</Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(SizeRw))}</Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(SizeRootFs))}</Table.Cell>
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
          <Grid.Column>
            <Statistic>
              <Statistic.Label>Total size</Statistic.Label>
              <Statistic.Value>{replaceWithNbsp(prettyBytes(totalSize))}</Statistic.Value>
            </Statistic>
          </Grid.Column>
          <Grid.Column textAlign="right" verticalAlign="bottom">
            <Header>Containers {prettyCount(count)}</Header>
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
          <code>{'$ docker container prune'}</code>
        </Message.Header>
        Remove all stopped containers. For more details, see{' '}
        <a rel="noreferrer" target="_blank" href="https://docs.docker.com/engine/reference/commandline/container_prune/">
          docker container prune.
        </a>
      </Message.Content>
    </Message>
  );
}

export default Containers;
