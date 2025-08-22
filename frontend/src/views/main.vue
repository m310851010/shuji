<template>
  <Window>
    <Titlebar></Titlebar>
    <a-layout class="h-100">
      <div class="flex-vertical">
        <div class="flex-main">
          <a-layout-sider width="220" class="sider" theme="light">
            <div class="address">
              <a-space direction="vertical">
                <a-tag color="processing" v-for="item in areas" :key="item">{{ item }}</a-tag>
              </a-space>
            </div>

            <a-menu mode="inline" class="menu" v-model:selectedKeys="selectedKeys">
              <a-menu-item v-for="item in menus" :key="item.path" @click="() => $router.push(item.path)">
                <span>{{ item.name }}</span>
              </a-menu-item>
            </a-menu>
          </a-layout-sider>
        </div>
        <div style="text-align: center; margin-bottom: 20px; font-size: 20px">
          <SettingOutlined @click="handleSettingClick" style="cursor: pointer" />
          <div style="font-size: 14px; margin-top: 30px">
            <div>技术支持</div>
            <div>XXX-XXX-XXX</div>
            <div>北京数极智能科技有限公司</div>
          </div>
        </div>
      </div>

      <a-layout-content class="content">
        <RouterView></RouterView>
      </a-layout-content>
    </a-layout>
  </Window>
</template>

<script setup lang="ts">
  import { RouterView, useRoute } from 'vue-router';
  import Window from '@/components/Window.vue';
  import { SettingOutlined } from '@ant-design/icons-vue';
  import { useRouter } from 'vue-router';
  import { GetAreaConfig } from '@wailsjs/go';

  // 菜单
  const menus = ref([
    { name: '数据导入', path: '/main/data-import' },
    { name: '数据校验', path: '/main/data-check' },
    { name: '数据导出', path: '/main/data-export' },
    { name: '清单导入', path: '/main/manifest-import' },
    { name: '导入进度', path: '/main/import-process' },
    { name: 'DB合并', path: '/main/db-merge' }
  ]);

  // 选中的菜单
  const selectedKeys = ref<string[]>([]);
  const route = useRoute();
  const $router = useRouter();

  // 监听路由变化, 并更新选中的菜单
  watchEffect(() => (selectedKeys.value = [route.path]));

  const handleSettingClick = () => {
    $router.push('/main/setting');
  };

  const areas = ref<string[]>([]);
  onMounted(async () => {
    const result = await GetAreaConfig();
    if (result.ok && result.data.length > 0) {
      const [data] = result.data;
      areas.value = [data.province_name, data.city_name, data.country_name].filter(v => v);
    }
  });
</script>

<style scoped>
  .sider.ant-layout-sider {
    background-color: #f1f4f8;
  }

  .content {
    background-color: #fff;
    position: relative;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .menu {
    background-color: transparent;
  }
  :deep(.menu) .ant-menu-item {
    font-size: 23px;
    margin: 0;
    width: 100%;
    padding: 25px 0 !important;
    text-align: center;
    &.ant-menu-item-selected {
      background-color: #b5d5f0;
    }
  }

  .address {
    padding: 15px 0 20px;
    text-align: center;
  }
  :deep(.address) .ant-tag {
    text-align: center;
    display: block;
    font-size: 13px;
    padding: 2px 5px;
    min-width: 120px;
  }
</style>
