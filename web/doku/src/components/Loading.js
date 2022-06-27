import React from 'react';
import { Container, Loader } from 'semantic-ui-react';

function Loading() {
  return (
    <Container>
      <Loader active>Loading</Loader>
    </Container>
  );
}

export default Loading;
