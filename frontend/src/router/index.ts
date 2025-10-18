import { createRouter, createWebHashHistory, RouteRecordRaw } from 'vue-router';
import main from '../views/main.vue';
const modules = import.meta.glob('../views/main/*.vue');

const mainRoutes: RouteRecordRaw[] = [];

for (let key in modules) {
  const name = key.replace('../views/', '').replace('.vue', '');
  let obj = {
    path: `/${name}`,
    name: `${name.replace(/\/+/, '-')}`,
    component: modules[key]
  };
  mainRoutes.push(obj);
}

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'index', redirect: '/main/data-check', meta: { title: '首页' } },
  {
    path: '/main',
    name: 'main',
    redirect: '/main/data-check',
    component: main,
    meta: { title: '主页' },
    children: mainRoutes
  }
];

const router = createRouter({
  history: createWebHashHistory(''),
  routes
});

export default router;
