<template>
  <Window>
    <Titlebar></Titlebar>
    <a-layout class="h-100">
      <div class="flex-vertical border-right">
        <div class="flex-main">
          <div class="title-container">
            <h2 class="main-title">数据校验工具</h2>
          </div>

          <!-- 灰色间隔线 -->
          <div class="divider-line"></div>

          <a-layout-sider width="220" class="sider" theme="light">
            <div class="address">
              <a-select
              size="small"
              class="area-tag"
            v-model:value="province"
            show-search
            placeholder="请选择省"
            @change="handleProvinceChange"
            :options="provinceOptions"
            :filter-option="filterOption"
          ></a-select>
            </div>

            <div class="divider-line"></div>

            <a-menu mode="inline" class="menu" v-model:selectedKeys="selectedKeys">
              <a-menu-item
                  v-for="item in menus"
                  :key="item.path"
                  :disabled="item.disabled"
                  @click="handleMenuClick(item)"
                  :class="{ 'disabled-menu-item': item.disabled }"
              >
                <span :class="{ 'disabled-text': item.disabled }">{{ item.name }}</span>
              </a-menu-item>
            </a-menu>
          </a-layout-sider>
        </div>
      </div>

      <a-layout-content class="content">
        <router-view v-slot="{ Component }">
          <keep-alive :exclude="['setting']">
            <component :is="Component" :key="route.path" />
          </keep-alive>
        </router-view>
      </a-layout-content>
    </a-layout>
  </Window>
</template>

<script setup lang="ts">
import { RouterView, useRoute } from 'vue-router';
import { computed, ref, watch, onMounted } from 'vue';
import Window from '@/components/Window.vue';
import { useRouter } from 'vue-router';
import { GetChinaAreaStr } from '@wailsjs/go';
import type { SelectProps } from 'ant-design-vue';
import { CurrentArea } from './constant';

// 菜单
const menus = ref([
  { name: '数据导入及校验', path: '/main/data-check', disabled: false },
  { name: '数据校验配置', path: '/main/data-check-config', disabled: false },
]);

  const province = ref('');
  const provinceOptions = ref<SelectProps['options']>([]);
  const filterOption = (input: string, option: any) => {
    return option.label.indexOf(input) >= 0;
  };

/**
 * 处理菜单项点击事件
 * @param item 菜单项对象
 */
const handleMenuClick = (item: any) => {
    $router.push(item.path);
};

// 选中的菜单
const selectedKeys = computed(() => [route.path]);
const route = useRoute();
const $router = useRouter();

let LOCATION_DATA: any[] = [];
onMounted(async () => {
  const res = await GetChinaAreaStr();
  LOCATION_DATA = JSON.parse(res.data);

  provinceOptions.value = LOCATION_DATA.map((item: any) => ({
    value: item.code,
    label: item.name
  }));

  province.value = LOCATION_DATA[0].code;
  CurrentArea.province = province.value;
});

const handleProvinceChange = (value: string) => {
  CurrentArea.province = LOCATION_DATA.find((item: any) => item.code === value)?.name || '';
};
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
    background-color: #035da4;
    color: #ffffff;
    font-weight: 500;
  }
  &:hover:not(.ant-menu-item-disabled) {
    background-color: #035da4;
    color: #ffffff;
  }
  &:hover {
    background-color: #035da4;
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

:deep(.address) {
  .ant-select-selector {
    text-align: center;
    display: block;
    font-size: 13px;
    border-radius: 3px  !important;
    background-color: rgb(107, 162, 212) !important;
    color: #fff !important;
    border: none !important;

    .ant-select-selection-item {
      color: #fff !important;
    }
  }

  .ant-select-suffix{
    color: #fff !important;
  }

  .ant-select-clear {
    background-color: transparent;
    color: #fff;
  }
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
  color: #035da4;
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

:deep(.ant-menu-inline) .ant-menu-item-selected::after{
  content: none;
}
</style>
<style>
.ant-btn-primary {
  background-color: #035da4 !important;
  border-color: #035da4 !important;
}

.ant-btn-primary:hover {
  background-color: #0f3a5f !important;
  border-color: #0f3a5f !important;
}

.ant-btn-primary:focus {
  background-color: #035da4 !important;
  border-color: #035da4 !important;
}
</style>
