import React from 'react';
import { NavLink } from 'react-router-dom';
import { Container, Header, Menu, Icon } from 'semantic-ui-react';
import { useDispatch, useSelector } from 'react-redux';
import { selectIsDarkTheme, setDarkTheme, setLightTheme } from '../AppSlice';

const azureBackgroundColor = {
  backgroundColor: 'azure',
};

function TopMenu() {
  const dispatch = useDispatch();
  const isDarkTheme = useSelector(selectIsDarkTheme);

  return (
    <Menu pointing secondary size="small" fixed="top" inverted={isDarkTheme} style={isDarkTheme ? null : azureBackgroundColor}>
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
          Volumes
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
        <Menu.Item
          onClick={() => {
            isDarkTheme ? dispatch(setLightTheme()) : dispatch(setDarkTheme());
          }}>
          {isDarkTheme ? <Icon name="sun outline" /> : <Icon name="moon outline" />}
        </Menu.Item>
        <Menu.Menu position="right">
          <Menu.Item>
            <Header inverted={isDarkTheme}>{window.config.header}</Header>
          </Menu.Item>
        </Menu.Menu>
      </Container>
    </Menu>
  );
}

export default TopMenu;
