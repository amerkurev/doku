// https://prettier.io/docs/en/options.html
module.exports = {
  printWidth: 140,
  useTabs: false,
  tabWidth: 2,
  semi: true,
  singleQuote: true,
  trailingComma: 'es5',
  bracketSpacing: true,
  bracketSameLine: true,
  overrides: [
    {
      files: ['*.html'],
      options: {
        trailingComma: 'none',
      },
    },
  ],
};
