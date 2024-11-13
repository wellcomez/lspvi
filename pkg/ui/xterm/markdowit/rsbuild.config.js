// rsbuild.config.js
module.exports = {
  entry: './main.js', // 入口文件
  output: {
    file: 'bundle.umd.js', // 输出文件名
    format: 'umd', // 输出格式为 UMD
    name: 'MyModule', // 全局变量名，用于在浏览器中访问模块
    globals: {
      // 这里定义了 main.js 中导入的模块的全局变量名
      // 例如，如果 main.js 中导入了 lodash，你可以这样定义：
      // lodash: '_'
    }
  }
};
