const { NODE_ENV } = process.env;

export const makeURL = function (path) {
  if (NODE_ENV === 'development') {
    return 'http://127.0.0.1:9090' + path;
  }
  return path;
};
