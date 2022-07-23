import moment from 'moment';

export const replaceWithNbsp = function (s) {
  return s.replace(/ /g, '\u00a0');
};

export const prettyCount = function (c) {
  return c > 0 ? `(${c})` : '';
};

export const prettyContainerName = function (s) {
  if (s.length > 0 && s[0] === '/') {
    return s.slice(1);
  }
  return s;
};

export const prettyContainerID = function (s) {
  return s.slice(0, 12);
};

export const prettyImageID = function (s) {
  if (s.startsWith('sha256:')) {
    return s.slice(7, 19);
  }
  return s.slice(0, 12);
};

const logPathRegExp = new RegExp('(/var/lib/docker/containers/\\w+/\\w{8})\\w+(-json.log)');

export const prettyLogPath = function (s) {
  if (logPathRegExp.test(s)) {
    const found = s.match(logPathRegExp);
    if (found.length === 3) {
      return found[1] + '***' + found[2];
    }
  }
  return s;
};

export const prettyTime = function (t) {
  return moment(t).format('YYYY-MM-DD\u00a0\u00a0HH:mm:ss Z');
};

export const prettyUnixTime = function (t) {
  return moment.unix(t).format('YYYY-MM-DD\u00a0\u00a0HH:mm:ss Z');
};
