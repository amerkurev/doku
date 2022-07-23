import React, { useReducer } from 'react';
import { useSelector } from 'react-redux';
import {
  selectDockerDiskUsage,
  selectDockerDiskUsageStatus,
  selectTotalSizeImages,
  selectCountImages,
  selectIsDarkTheme,
  selectDockerContainerList,
} from '../AppSlice';
import { CHANGE_SORT, sortReducer, sortReducerInitializer } from '../util/sort';
import statusPage from './StatusPage';
import { sortBy } from 'lodash/collection';
import { Container, Icon, Message, Popup, Statistic, Table, Grid, Header } from 'semantic-ui-react';
import { prettyContainerName, prettyCount, prettyImageID, prettyUnixTime, replaceWithNbsp } from '../util/fmt';
import prettyBytes from 'pretty-bytes';

function Images() {
  const isDarkTheme = useSelector(selectIsDarkTheme);
  const diskUsage = useSelector(selectDockerDiskUsage);
  const diskUsageStatus = useSelector(selectDockerDiskUsageStatus);
  const containerList = useSelector(selectDockerContainerList);
  const totalSize = useSelector(selectTotalSizeImages);
  const count = useSelector(selectCountImages);
  const [state, dispatch] = useReducer(sortReducer, sortReducerInitializer());

  const s = statusPage(diskUsage, diskUsageStatus);
  if (s !== null) {
    return s;
  }

  let dataTable = null;

  if (Array.isArray(diskUsage.Images) && diskUsage.Images.length > 0) {
    const { column, direction } = state;
    const data = sortBy(
      diskUsage.Images.map((x) => {
        const containers = getContainers(containerList, x.Id);
        const repoTags = Array.isArray(x.RepoTags) ? x.RepoTags.join('\n') : '';
        const repoDigests = Array.isArray(x.RepoDigests) ? x.RepoDigests.join('\n') : '';
        const extra = {
          ID: x.Id,
          Containers: containers.length === 0 ? '-' : containers.join('\n'),
          ContainersNum: containers.length.toString() + containers.join(''), // for alphanumeric sort if length is equal
          RepoTags: repoTags,
          RepoDigests: repoDigests,
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
            <Table.HeaderCell
              sorted={column === 'RepoTags' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'RepoTags' })}>
              Repository Tags
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'Size' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'Size' })}>
              Size
            </Table.HeaderCell>
            <Table.HeaderCell
              textAlign="right"
              sorted={column === 'SharedSize' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'SharedSize' })}>
              Shared Size
            </Table.HeaderCell>
            <Table.HeaderCell
              sorted={column === 'ContainersNum' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'ContainersNum' })}>
              Containers
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
          {data.map(({ Containers, Created, ID, RepoDigests, RepoTags, SharedSize, Size }) => (
            <Table.Row key={ID}>
              <Table.Cell>
                <small>
                  <code>{prettyImageID(ID)}</code>
                </small>
              </Table.Cell>
              <Table.Cell style={{ whiteSpace: 'pre-line' }}>{RepoTags}</Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(Size))}</Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(SharedSize))}</Table.Cell>
              <Table.Cell style={{ whiteSpace: 'pre-line' }}>{Containers}</Table.Cell>
              <Table.Cell textAlign="center">{prettyUnixTime(Created)}</Table.Cell>
              <Popup
                inverted={isDarkTheme}
                wide="very"
                header="Digests"
                content={RepoDigests}
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
            <Header inverted={isDarkTheme}>Images {prettyCount(count)}</Header>
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
          <code>{'$ docker image prune'}</code>
        </Message.Header>
        Remove unused images. For more details, see{' '}
        <a rel="noreferrer" target="_blank" href="https://docs.docker.com/engine/reference/commandline/image_prune/">
          docker image prune.
        </a>
      </Message.Content>
    </Message>
  );
}

function getContainers(containers, ImageId) {
  const res = [];
  if (containers && Array.isArray(containers.Containers) && containers.Containers.length > 0) {
    for (let i = 0; i < containers.Containers.length; i++) {
      const x = containers.Containers[i];
      if (x.Image === ImageId) {
        res.push(prettyContainerName(x.Name));
      }
    }
  }
  return res;
}

export default Images;
