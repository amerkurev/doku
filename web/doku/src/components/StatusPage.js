import React from 'react';
import Loading from './Loading';
import { isEmpty } from 'lodash';
import EmptyData from './EmptyData';

function statusPage(data, status) {
  if (status === 'loading' && data == null) {
    return <Loading />; // data loading
  }

  if (data === null) {
    return <div />; // initial state
  }

  if (isEmpty(data)) {
    return <EmptyData />; // data not ready yet
  }

  return null;
}

export default statusPage;
