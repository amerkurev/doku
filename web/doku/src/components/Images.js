import React, { useReducer } from 'react';
import { useSelector } from 'react-redux';
import { selectDockerDiskUsage, selectDockerDiskUsageStatus } from '../AppSlice';
import { CHANGE_SORT, sortReducer, sortReducerInitializer } from '../util/sort';
import statusPage from './StatusPage';
import { sortBy } from 'lodash/collection';
import { sumBy } from 'lodash/math';
import { Container, Icon, Message, Popup, Statistic, Table, Grid, Header } from 'semantic-ui-react';
import { prettyImageID, prettyUnixTime, replaceWithNbsp } from '../util/fmt';
import prettyBytes from 'pretty-bytes';

function Images() {
  const diskUsage = useSelector(selectDockerDiskUsage);
  const diskUsageStatus = useSelector(selectDockerDiskUsageStatus);
  const [state, dispatch] = useReducer(sortReducer, sortReducerInitializer());

  const s = statusPage(diskUsage, diskUsageStatus);
  if (s !== null) {
    return s;
  }

  let dataTable = null;
  let totalSize = 0;

  if (Array.isArray(diskUsage.Images) && diskUsage.Images.length > 0) {
    const { column, direction } = state;
    const data = sortBy(
      diskUsage.Images.map((x) => {
        const repoTags = Array.isArray(x.RepoTags) ? x.RepoTags.join('\n') : '';
        const repoDigests = Array.isArray(x.RepoDigests) ? x.RepoDigests.join('\n') : '';
        return { ...x, ...{ RepoTags: repoTags, RepoDigests: repoDigests, ID: x.Id } };
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
              textAlign="center"
              sorted={column === 'Containers' ? direction : null}
              onClick={() => dispatch({ type: CHANGE_SORT, column: 'Containers' })}>
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
          {data.map(({ Containers, Created, ID, Labels, ParentId, RepoDigests, RepoTags, SharedSize, Size, VirtualSize }) => (
            <Table.Row key={ID}>
              <Table.Cell>
                <small>
                  <code>{prettyImageID(ID)}</code>
                </small>
              </Table.Cell>
              <Table.Cell style={{ whiteSpace: 'pre-line' }}>{RepoTags}</Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(Size))}</Table.Cell>
              <Table.Cell textAlign="right">{replaceWithNbsp(prettyBytes(SharedSize))}</Table.Cell>
              <Table.Cell textAlign="center">{Containers}</Table.Cell>
              <Table.Cell textAlign="center">{prettyUnixTime(Created)}</Table.Cell>
              <Popup
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

    const sharedSizes = [];

    totalSize = sumBy(data, (x) => {
      const sharedSize = x.SharedSize;
      if (sharedSize) {
        if (sharedSizes.indexOf(sharedSize) > -1) {
          return x.Size - sharedSize;
        } else {
          sharedSizes.push(sharedSize);
        }
      }
      return x.Size;
    });
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
            <Header>Images</Header>
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
        Remove unused images. See details of{' '}
        <a rel="noreferrer" target="_blank" href="https://docs.docker.com/engine/reference/commandline/image_prune/">
          docker image prune
        </a>
      </Message.Content>
    </Message>
  );
}

export default Images;
