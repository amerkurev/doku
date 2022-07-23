import React from 'react';
import { Container, Image, Menu } from 'semantic-ui-react';
import GitHubLogo from '../assets/logo/github-logo.svg';
import { useSelector } from 'react-redux';
import { selectVersion } from '../AppSlice';

const verRegExp = /(^v\d+.\d+.\d+)-.*$/;

function Footer() {
  const revision = useSelector(selectVersion);
  let version = '';
  if (typeof revision === 'string' || revision instanceof String) {
    version = revision;
    if (verRegExp.test(revision)) {
      const found = revision.match(verRegExp);
      if (found.length > 1) {
        version = found[1];
      }
    }
  }

  return (
    <Menu inverted borderless size="mini" fixed="bottom">
      <Container textAlign="center">
        <Menu.Item>
          <Image centered size="mini" verticalAlign="middle" src={GitHubLogo} href="https://github.com/amerkurev/doku" />
        </Menu.Item>
        <Menu.Menu position="right">
          <Menu.Item as="h5">{version}</Menu.Item>
        </Menu.Menu>
      </Container>
    </Menu>
  );
}

export default Footer;
