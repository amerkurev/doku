export const CHANGE_SORT = 'CHANGE_SORT';
export const ASC = 'ascending';
export const DESC = 'descending';

export function sortReducerInitializer(column) {
  return {
    column: column || 'Size',
    direction: DESC,
  };
}

export function sortReducer(state, action) {
  switch (action.type) {
    case CHANGE_SORT:
      if (state.column === action.column) {
        return {
          ...state,
          direction: state.direction === ASC ? DESC : ASC,
        };
      }

      return {
        column: action.column,
        direction: ASC,
      };
    default:
      throw new Error();
  }
}
