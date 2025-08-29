<template>
  <Window>
    <Titlebar></Titlebar>
    <a-layout class="h-100">
      <div class="flex-vertical border-right">
        <div class="flex-main">
          <div class="title-container">
            <h2 class="main-title">煤炭摸底数据校验工具</h2>
          </div>

          <!-- 灰色间隔线 -->
          <div class="divider-line"></div>

          <a-layout-sider width="220" class="sider" theme="light">
            <div class="address">
              <a-tag v-for="(item, index) in areas" :key="item" color="#6BA2D4" class="area-tag">{{ item }}</a-tag>
            </div>

            <div class="divider-line"></div>

            <a-menu mode="inline" class="menu" v-model:selectedKeys="selectedKeys">
              <a-menu-item v-for="item in menus" :key="item.path" @click="() => $router.push(item.path)">
                <span>{{ item.name }}</span>
              </a-menu-item>
            </a-menu>
          </a-layout-sider>

          <div class="divider-line"></div>
        </div>
        <!-- 灰色间隔线 -->
        <div class="divider-line"></div>

        <div>
          <div
            class="independent-menu-item"
            @click="handleDbMergeClick"
            :style="dbMergeButtonStyle"
            @mouseover="handleDbMergeMouseOver"
            @mouseleave="handleDbMergeMouseLeave"
          >
            DB文件合并
          </div>
        </div>

        <!-- 灰色间隔线 -->
        <div class="divider-line"></div>

        <div class="bottom-section">
          <SettingOutlined
            :class="['setting-icon', { 'setting-icon-active': isSettingRoute }]"
            @click="handleSettingClick"
            style="cursor: pointer"
          />

          <div class="support-info">
            <div>技术支持</div>
            <div>XXX-XXX-XXX</div>
            <!-- <div>北京数极智能科技有限公司</div> -->
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
  import { computed } from 'vue';
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
    { name: '导入进度', path: '/main/import-process' }
    // { name: 'DB文件合并', path: '/main/db-merge' }
  ]);

  // 选中的菜单
  const selectedKeys = ref<string[]>([]);
  const route = useRoute();
  const $router = useRouter();
  const isSettingRoute = computed(() => route.path === '/main/setting');

  // 监听路由变化, 并更新选中的菜单
  watchEffect(() => (selectedKeys.value = [route.path]));

  /**
   * 处理设置按钮点击事件
   */
  const handleSettingClick = () => {
    $router.push('/main/setting');
  };

  /**
   * DB文件合并按钮样式（响应式计算属性）
   */
  const dbMergeButtonStyle = computed(() => {
    const isSelected = route.path === '/main/db-merge';
    return {
      fontSize: '18px',
      margin: '8px 16px',
      width: 'calc(100% - 32px)',
      padding: '12px 24px',
      textAlign: 'center' as const,
      cursor: 'pointer',
      backgroundColor: isSelected ? '#1A5284' : '#ffffff',
      color: isSelected ? '#ffffff' : '#000000',
      transition: 'background-color 0.3s, color 0.3s',
      borderRadius: '25px',
      border: '1px solid #d9d9d9'
    };
  });

  /**
   * 处理DB文件合并按钮点击事件
   */
  const handleDbMergeClick = () => {
    $router.push('/main/db-merge');
  };

  /**
   * 处理DB文件合并按钮鼠标悬停事件
   */
  const handleDbMergeMouseOver = (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    target.style.backgroundColor = '#1A5284';
    target.style.color = '#ffffff';
  };

  /**
   * 处理DB文件合并按钮鼠标离开事件
   */
  const handleDbMergeMouseLeave = (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    const isSelected = route.path === '/main/db-merge';
    target.style.backgroundColor = isSelected ? '#1A5284' : '#ffffff';
    target.style.color = isSelected ? '#ffffff' : '#000000';
  };

  const areas = ref<string[]>([]);
  onMounted(async () => {
    const result = await GetAreaConfig();
    if (result.ok && result.data) {
      const data = result.data;
      areas.value = [data.province_name, data.city_name, data.country_name].filter(v => v);
    }
  });
</script>

<style scoped>
  .sider.ant-layout-sider {
    background-color: #f9fafb;
  }
  ::v-deep .ant-menu-light.ant-menu-root.ant-menu-inline {
    border-right: 0px;
  }
  .content {
    background-color: #fff;
    position: relative;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .menu {
    margin-top: 10px;
    background-color: transparent;
  }
  :deep(.menu) .ant-menu-item {
    font-size: 23px;
    margin: 0;
    width: 100%;
    padding: 25px 0 !important;
    text-align: center;
    transition:
      background-color 0.3s,
      color 0.3s;
    &.ant-menu-item-selected {
      background-color: #1a5284;
      color: #ffffff;
      font-weight: 500;
    }
    &:hover {
      background-color: #1a5284;
      color: #ffffff;
    }
  }

  .address {
    padding: 15px 0 20px;
    text-align: center;
    display: flex;
    justify-content: center;
    flex-direction: column;
  }
  :deep(.address) .ant-tag {
    text-align: center;
    display: block;
    font-size: 13px;
    padding: 2px 5px;
    min-width: 120px;
    border-radius: 3px;
  }

  /* 图标 hover 颜色 */
  .setting-icon {
    font-size: 20px;
    transition: color 0.3s;
  }
  .setting-icon-active {
    color: #1a5284;
  }
  .setting-icon:hover {
    color: #1a5284;
    font-weight: 600;
  }

  /* 灰色间隔线样式 */
  .divider-line {
    width: 100%;
    margin: 0 auto;
    border: 0.8px solid #e8e8e8c0;
  }

  /* 标题容器样式 */
  .title-container {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 32px;
    margin: 10px 20px;
  }

  /* 主标题样式 */
  .main-title {
    margin: 0;
    font-size: 18px;
    font-weight: 800;
    color: #1a5284;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  }

  /* 技术支持信息样式 */
  .support-info {
    font-size: 14px;
    margin: 20px 0;
  }

  /* 底部区域样式 */
  .bottom-section {
    text-align: center;
    padding-top: 20px;
    background-color: #f9fafb;
  }

  /* 地址标签间距样式 */
  .area-tag {
    margin: 4px 8px !important;
  }
</style>
