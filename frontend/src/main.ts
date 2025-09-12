import { createApp } from 'vue';
import App from './App.vue';
import './style.less';
import router from './router';
import { ComponentPlugin } from '@/components';
import 'ant-design-vue/dist/antd.less';

const app = createApp(App);
app.use(router);
// app.use(Antd);
app.use(ComponentPlugin);
app.mount('#app');

setTimeout(() => {
  // @ts-ignore
  const wapp = window.go.main.App;
  for (const k in wapp) {
    const fn = wapp[k];
    wapp[k] = function () {
      const args = [...arguments];
      console.log('函数:' + k, '参数==', args);
      const res = fn.apply(app, args);
      return res
        .then((data: any) => {
          console.log('函数:' + k, '返回值==', data);
          return data;
        })
        .catch((err: any) => {
          console.log('函数:' + k, '错误==', err);
          return err;
        });
    };
  }
});
