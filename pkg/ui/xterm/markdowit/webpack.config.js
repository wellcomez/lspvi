const path = require('path');

module.exports = {
  entry: './main.js', // 入口文件
  output: {
    filename: 'mark.js', // 打包后的文件名
    path: path.resolve(__dirname, 'dist'), // 打包后的文件路径
  },
  module: {
    rules: [
      {
        test: /\.m?js$/,
        exclude: /(node_modules|bower_components)/,
        use: {
          loader: 'babel-loader',
          options: {
            presets: ['@babel/preset-env'],
          },
        },
      },
    ],
  },
};
