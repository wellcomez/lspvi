import resolve from 'rollup-plugin-node-resolve';
import commonjs from 'rollup-plugin-commonjs';
import babel from '@rollup/plugin-babel';
import { terser } from 'rollup-plugin-terser';
const isProduction = process.env.NODE_ENV === 'production';
export default {
    input: 'index.js',
    output: [
        {
            file: 'dist/mark.umd.js',
            format: 'umd',
            sourcemap: true,
            name: 'markjs',
            globals: {
                // 如果你的库依赖其他库，这里可以指定全局变量
            }
        },
        {
            file: 'dist/mark.esm.js',
            format: 'esm'
        }
    ],
    plugins: [
        resolve(),
        commonjs(),
        babel({
            exclude: 'node_modules/**',
            presets: ['@babel/preset-env']
        }),
        terser()
    ],
    external: [
        // 如果你的库依赖其他库，这里可以指定外部依赖
    ]
};