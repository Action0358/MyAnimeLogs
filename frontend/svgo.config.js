module.exports = {
  multipass: true,
  js2svg: {
    indent: 2,
    pretty: true,
  },
  plugins: [
    {
      name: 'preset-default',
      params: {
        overrides: {
          // 数値の精度を1桁に
          cleanupNumericValues: {
            floatPrecision: 1,
          },
        },
      },
    },
    // 属性をアルファベット順にソート
    'sortAttrs',
  ],
}
