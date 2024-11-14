/* eslint-disable no-undef */
const path = require('path');
const webpack = require('webpack');
const TerserPlugin = require('terser-webpack-plugin');
// const { SourceMap } = require('module');

module.exports = {
    mode: 'development',
    entry: './index.js',
    devtool: 'source-map',
    output: {
        path: path.resolve(__dirname, 'dist'),
        filename: 'mark.umd.js',
        library: 'markjs',
        libraryTarget: 'umd',
        globalObject: 'this'
    },
    plugins: [
        new webpack.ProvidePlugin({
            process: 'process/browser',
            // 如果你需要提供全局变量，可以在这里配置
        }),
        new TerserPlugin() // 用于压缩输出文件
    ],
    externals: {
        // 如果你需要指定外部依赖，可以在这里配置
    }
};
