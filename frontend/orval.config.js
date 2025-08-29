module.exports = {
  myanimelog: {
    input: '../openapi/openapi.yaml',
    output: {
      target: './src/lib/api/generated/myanimelog.ts',
      client: 'react-query',
      mock: true,
    },
  },
}
