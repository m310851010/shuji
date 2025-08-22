import { createApp } from 'vue';
// import Antd from 'ant-design-vue';
import App from './App.vue';
import 'ant-design-vue/dist/reset.css';
import './style.less';
import router from './router';
import { ComponentPlugin } from '@/components';
import { setupDuplicateData } from './hook/useDuplicateData';
import { GetEnv } from '@wailsjs/go';

const app = createApp(App);
app.use(router);
// app.use(Antd);
app.use(ComponentPlugin);
app.mount('#app');

setupDuplicateData();

// OnFileDrop((x, y, paths) => {
//   console.log(x, y, paths);
// }, true);
