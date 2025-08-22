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
  { path: '/', name: 'index', redirect: '/login', meta: { title: '首页' } },
  { path: '/select-address', name: 'select-address', component: () => import('../views/select-address.vue'), meta: { title: '选择区域' } },
  { path: '/login', name: 'login', component: () => import('../views/login.vue'), meta: { title: '登录' } },
  {
    path: '/main',
    name: 'main',
    redirect: '/main/data-import',
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
