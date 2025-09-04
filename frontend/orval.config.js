module.exports = {
  myanimelogs: {
    input: '../openapi/openapi.yaml',
    output: {
      target: './src/lib/api/generated/myanimelogs.ts',
      client: 'react-query',
      mock: true,
    },
  },
}
