import React, { useReducer } from 'react';
import { useSelector } from 'react-redux';
import {
  selectDockerContainerList,
  selectDockerContainerListStatus,
  selectTotalSizeContainers,
  selectCountContainers,
  selectIsDarkTheme,
  selectDockerDiskUsage,
} from '../AppSlice';
import { CHANGE_SORT, sortReducer, sortReducerInitializer } from '../util/sort';
import statusPage from './StatusPage';
import { Container, Grid, Header, Icon, Message, Popup, Statistic, Table } from 'semantic-ui-react';
import { prettyContainerID, prettyContainerName, prettyCount, prettyTime, replaceWithNbsp } from '../util/fmt';
import prettyBytes from 'pretty-bytes';
import { sortBy } from 'lodash/collection';

function Containers() {
  const isDarkTheme = useSelector(selectIsDarkTheme);
  const diskUsage = useSelector(selectDockerDiskUsage);
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
        const extra = {
          ID: x.Id,
          Size: x.SizeRw,
          Status: x.State.Status,
          ImageNames: getImageNames(diskUsage, x.Image),
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
            <Table.HeaderCell sorted={column === 'ID' ? direction : null} onClick={() => dispatch({ type: CHANGE_SORT, column: 'ID' })}>
              ID
            </Table.HeaderCell>
            <Table.HeaderCell sorted={column === 'Name' ? direction : null} onClick={() => dispatch({ type: CHANGE_SORT, column: 'Name' })}>
              Name
            </Table.HeaderCell>
            <Table.HeaderCell
              sorted={column === 'ImageNames' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'ImageNames' })}>
              Image
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'Size' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'Size' })}>
              {'Size RW '}
              <Popup
                inverted={isDarkTheme}
                wide="very"
                header="Size RW"
                content={'The size of files that have been created or changed by this container'}
                trigger={<Icon name="question circle outline" />}
              />
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'SizeRootFs' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'SizeRootFs' })}>
              {'Virtual Size '}
              <Popup
                inverted={isDarkTheme}
                wide="very"
                header="Virtual Size"
                content={'The total size of all the files in this container'}
                trigger={<Icon name="question circle outline" />}
              />
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
            <Table.HeaderCell />
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {data.map(({ ID, Name, Image, ImageNames, Created, Size, SizeRootFs, Status }) => (
            <Table.Row key={ID}>
              <Table.Cell>
                <small>
                  <code>{prettyContainerID(ID)}</code>
                </small>
              </Table.Cell>
              <Table.Cell>{prettyContainerName(Name)}</Table.Cell>
              <Table.Cell style={{ whiteSpace: 'pre-line' }}>{ImageNames}</Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(Size))}</Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(SizeRootFs))}</Table.Cell>
              <Table.Cell textAlign="center">{Status}</Table.Cell>
              <Table.Cell textAlign="center">{prettyTime(Created)}</Table.Cell>
              <Popup
                inverted={isDarkTheme}
                wide="very"
                header="Image ID"
                content={Image}
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
            <Header inverted={isDarkTheme}>Containers {prettyCount(count)}</Header>
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

function getImageNames(diskUsage, ImageID) {
  if (diskUsage && Array.isArray(diskUsage.Images) && diskUsage.Images.length > 0) {
    for (let i = 0; i < diskUsage.Images.length; i++) {
      const x = diskUsage.Images[i];
      if (x.Id === ImageID) {
        return x.RepoTags.join('\n');
      }
    }
  }
  return ImageID;
}

export default Containers;
