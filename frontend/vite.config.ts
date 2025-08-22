import { resolve } from 'path';

import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import vueJsx from '@vitejs/plugin-vue-jsx';
import legacy from '@vitejs/plugin-legacy';
import AutoImport from 'unplugin-auto-import/vite';
import Components from 'unplugin-vue-components/vite';
import { AntDesignVueResolver } from 'unplugin-vue-components/resolvers';
import { theme } from 'ant-design-vue';

export const r = (...args: any[]) => resolve(__dirname, '.', ...args);

// https://vitejs.dev/config/
export default defineConfig(env => {
  const root = process.cwd();
  const envPrefix = 'M_';
  const envConfig = loadEnv(env.mode, root, envPrefix);

  const { defaultAlgorithm, defaultSeed } = theme;
  const mapToken = defaultAlgorithm(defaultSeed);

  return {
    base: envConfig.M_PUBLIC_PATH,
    envPrefix,
    root,
    build: {
      target: 'es2015',
      rollupOptions: {
        input: {
          main: r('index.html')
        }
      }
    },
    css: {
      preprocessorOptions: {
        less: {
          modifyVars: { ...mapToken, 'primary-color': '#1DA57A' },
          javascriptEnabled: true
        }
      }
    },
    plugins: [
      vue(),
      vueJsx(),
      AutoImport({
        imports: ['vue'],
        dts: r('src/auto-imports.d.ts')
      }),
      Components({
        resolvers: [
          AntDesignVueResolver({
            importStyle: false,
            // importStyle: 'less', // 不需要，v4 默认使用 CSS-in-JS
            resolveIcons: true // 如果使用图标
          })
        ]
      }),
      legacy()
    ],
    resolve: {
      alias: {
        '@/': `${resolve(__dirname, 'src')}/`,
        '@wailsapp/runtime': r('wailsjs/runtime/runtime'),
        '@wailsjs/go': r('wailsjs/go/main/App'),
        '@wailsjs/models': r('wailsjs/go/models')
      }
    }
  };
});
