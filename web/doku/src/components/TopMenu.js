import React from 'react';
import { NavLink } from 'react-router-dom';
import { Container, Header, Menu } from 'semantic-ui-react';

const styles = {
  backgroundColor: 'azure',
};

function TopMenu() {
  return (
    <Menu pointing secondary size="small" fixed="top" style={styles}>
      <Container>
        <Menu.Item as={NavLink} to="/">
          Dashboard
        </Menu.Item>
        <Menu.Item as={NavLink} to="/images/">
          Images
        </Menu.Item>
        <Menu.Item as={NavLink} to="/containers/">
          Containers
        </Menu.Item>
        <Menu.Item as={NavLink} to="/volumes/">
          Local Volumes
        </Menu.Item>
        <Menu.Item as={NavLink} to="/bind-mounts/">
          Bind Mounts
        </Menu.Item>
        <Menu.Item as={NavLink} to="/logs/">
          Logs
        </Menu.Item>
        <Menu.Item as={NavLink} to="/build-cache/">
          Build Cache
        </Menu.Item>
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
