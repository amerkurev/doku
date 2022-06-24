import React from 'react';
import { Container, Header, Menu } from 'semantic-ui-react';

function TopMenu() {
  return (
    <Menu pointing secondary size="small" fixed="top">
      <Container>
        <Menu.Item as="a" active>
          Dashboard
        </Menu.Item>
        <Menu.Item as="a">Images</Menu.Item>
        <Menu.Item as="a">Containers</Menu.Item>
        <Menu.Item as="a">Local Volumes</Menu.Item>
        <Menu.Item as="a">Bind Mounts</Menu.Item>
        <Menu.Item as="a">Logs</Menu.Item>
        <Menu.Item as="a">Build Cache</Menu.Item>
        <Menu.Menu position="right">
          <Menu.Item>
            <Header>{window.config.header}</Header>
          </Menu.Item>
        </Menu.Menu>
      </Container>
    </Menu>
  );
}

export default TopMenu;
