module.exports = {
  env: {
    browser: true,
    es6: true,
  },
  extends: [
    'airbnb-base',
  ],
  globals: {
    Atomics: 'readonly',
    SharedArrayBuffer: 'readonly',
  },
  parserOptions: {
    ecmaVersion: 2018,
    sourceType: 'module',
  },
  rules: {
    'no-param-reassign': ['off'],
    'no-underscore-dangle': ['off'],
    'no-return-assign': ['off'],
    'no-restricted-syntax': ['warn'],
    'max-len': ['off'],
    'class-methods-use-this': ['off'],
    'no-plusplus': ['off'],
    'import/prefer-default-export': ['off'],
    'max-classes-per-file': ['off'],
    'object-shorthand': ['off'],
    'no-bitwise': ['warn'],
    'prefer-destructuring': ['off'],
    'no-lone-blocks':['off'],
    'camelcase': ['off'],
    'no-shadow': ['warn'],
  },
};
