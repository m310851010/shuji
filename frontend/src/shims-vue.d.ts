/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue';
  const component: DefineComponent<{}, {}, any>;
  export default component;
}

declare namespace JSX {
  interface IntrinsicElements {
    [elemName: string]: any;
  }
  interface Element {
    [elemName: string]: any;
  }
}

// 增强File
declare type EnhancedFile = File & {
  fullPath: string;
  isDirectory: boolean;
  isFile: boolean;
};
