import React from 'react';
import { Container, Message } from 'semantic-ui-react';

function EmptyData() {
  return (
    <Container>
      <Message warning size="tiny">
        <Message.Content>
          <Message.Header>Sorry...</Message.Header>
          {"The data preparation process didn't have enough time. Refresh the page after a few seconds."}
        </Message.Content>
      </Message>
    </Container>
  );
}

export default EmptyData;
