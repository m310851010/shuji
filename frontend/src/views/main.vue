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
            数据文件合并
          </div>
          <div
            class="independent-menu-item"
            @click="handleDbToExcelClick"
            :style="dbToExcelButtonStyle"
            @mouseover="handleDbToExcelMouseOver"
            @mouseleave="handleDbToExcelMouseLeave"
          >
            数据文件转Excel
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
            <!--            <div>技术支持</div>
            <div>XXX-XXX-XXX</div>-->
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
  import { computed, ref, watch, onMounted } from 'vue';
  import Window from '@/components/Window.vue';
  import { SettingOutlined } from '@ant-design/icons-vue';
  import { useRouter } from 'vue-router';
  import { GetAreaConfig, GetStateManifest, UpdateStateManifest } from '@wailsjs/go';
  import { Modal } from 'ant-design-vue';
  import { message } from 'ant-design-vue';

  // manifest 状态
   const manifestState = ref<any>(null);

   // 菜单
   const menus = ref([
     { name: '清单导入', path: '/main/manifest-import', disabled: false },
     { name: '数据导入', path: '/main/data-import', disabled: false },
     { name: '数据校验', path: '/main/data-check', disabled: false },
     { name: '数据导出', path: '/main/data-export', disabled: false },    
     { name: '导入进度', path: '/main/import-process', disabled: false }
     // { name: '数据文件合并', path: '/main/db-merge' }
   ]);

   /**
    * 监听 manifestState 的变化，自动更新菜单状态
    */
   watch(manifestState, (newValue) => {
     console.log('监听到 manifest 状态变化:', newValue);
     
     if (newValue === 1) {
       // 状态为 1：禁用清单导入菜单项
       setMenusDisabledStateByManifest(1);
     } else if (newValue === 2) {
       // 状态为 2：禁用除清单导入外的菜单项
       setMenusDisabledStateByManifest(2);
     } else if (newValue === 3) {
       // 状态为 3：解除所有菜单禁用
       setMenusDisabledStateByManifest(3);
     }
   }, { immediate: false });

  /**
      * 获取 state.json 中的 manifest 状态
      */
     const getManifestState = async () => {
       try {
         const result = await GetStateManifest();
         if (result.ok) {
           manifestState.value = result.data;
           console.log('获取到的 manifest 值:', result.data);
           
           // 根据 manifest 状态值执行不同逻辑
           if (result.data === null) {
             // 状态为 null：显示弹框询问用户
             showManifestDialog();
           } else if (result.data === 1) {
             // 状态为 1：禁用清单导入菜单项
             setMenusDisabledStateByManifest(1);
           } else if (result.data === 2) {
             // 状态为 2：禁用除清单导入外的菜单项
             
             $router.push('/main/manifest-import');
             setMenusDisabledStateByManifest(2);
             console.log('跳转--------------')
           } else if (result.data === 3) {
             // 状态为 3：解除所有菜单禁用
             setMenusDisabledStateByManifest(3);
           }
         } else {
           console.error('获取 manifest 值失败:', result.message);
           manifestState.value = null;
           // 获取失败时也显示弹框
           showManifestDialog();
         }
       } catch (error) {
         console.error('获取 manifest 状态时发生错误:', error);
         manifestState.value = null;
         // 发生错误时也显示弹框
         showManifestDialog();
       }
     };

     /**
      * 根据manifest状态设置菜单项的禁用状态
      * @param {number} manifestValue - manifest状态值（1、2或3）
      */
     const setMenusDisabledStateByManifest = (manifestValue: number) => {
       menus.value.forEach(menu => {
         if ('disabled' in menu) {
           if (manifestValue === 1) {
             // 状态为1：禁用清单导入菜单项
             menu.disabled = menu.path === '/main/manifest-import';
           } else if (manifestValue === 2) {
             // 状态为2：禁用除清单导入外的其他菜单项
             menu.disabled = menu.path !== '/main/manifest-import';
           } else if (manifestValue === 3) {
             // 状态为3：解除所有菜单禁用
             menu.disabled = false;
           }
         }
       });
       
     };

     /**
      * 禁用除清单导入外的其他菜单项
      */
     const disableMenusExceptManifestImport = () => {
       menus.value.forEach(menu => {
         if (menu.path === '/main/manifest-import') {
           menu.disabled = false;
         } else {
           menu.disabled = true;
         }
       });
     };

   /**
     * 显示清单上传确认弹框
     */
    const showManifestDialog = () => {
      Modal.confirm({
        title: '模式选择：',
        content: ' 请选择数据校验模式，如果掌握清单请优先使用 ” 有清单模式 “ ',
        okText: '有清单模式',
        cancelText: '无清单模式',
        // style: { marginLeft: '30px' },
        async onOk() {

          // 禁用除清单导入外的其他菜单项
          disableMenusExceptManifestImport();
          // 先跳转到清单导入页面
          $router.push('/main/manifest-import');
          selectedKeys.value = ['manifest-import'];
          // 更新 state.json 中的 manifest 字段为 2
          try {
            const result = await UpdateStateManifest(2);
            if (result.ok) {
              console.log('已将 manifest 状态设置为 2');
            } else {
            }
          } catch (error) {
            console.error('更新 manifest 状态时发生错误:', error);
          }
        },
        async onCancel() {
          // 更新 state.json 中的 manifest 字段为 1
          try {
            const result = await UpdateStateManifest(1);
            if (result.ok) {
              console.log('已将 manifest 状态设置为 1');
              manifestState.value = 1;
            } else {
              console.error('更新 manifest 状态失败:', result.message);
            }
          } catch (error) {
            console.error('更新 manifest 状态时发生错误:', error);
          }
        }
      });
    };

  /**
   * 处理菜单项点击事件
   * @param item 菜单项对象
   */
  const handleMenuClick = (item: any) => {
    // 如果菜单项被禁用，则显示提示并阻止跳转
    if (item.disabled) {
      
      message.error('请上传清单后操作！');
      return; // 阻止后续执行
    } else  {
      // 执行正常的路由跳转
      $router.push(item.path);
    }
    
    
  };

  // 选中的菜单
  const selectedKeys = ref<string[]>([]);
  const route = useRoute();
  const $router = useRouter();
  const isSettingRoute = computed(() => route.path === '/main/setting');

  /**
   * 监听路由变化, 并更新选中的菜单
   * 如果目标路由对应的菜单项被禁用，则不切换路由
   */
  watchEffect(() => {
    // 查找当前路由对应的菜单项
    const currentMenu = menus.value.find(menu => menu.path === route.path);
    
    // 如果菜单项存在且未被禁用，则更新选中状态
    if (currentMenu && !currentMenu.disabled) {
      selectedKeys.value = [route.path];
    } else if (currentMenu && currentMenu.disabled) {
      // 如果菜单项被禁用，阻止路由切换
      // 状态为2时，只允许跳转到清单导入页面
      if (manifestState.value === 2) {
        $router.replace('/main/manifest-import');
      } else {
        const lastValidRoute = selectedKeys.value[0] || '/main/data-import';
        $router.replace(lastValidRoute);
      }
    } else {
      // 如果找不到对应菜单项，使用默认路由
      selectedKeys.value = [route.path];
    }
  });

  

  /**
   * 处理设置按钮点击事件
   */
  const handleSettingClick = () => {
    $router.push('/main/setting');
  };

  /**
   * 数据文件合并按钮样式（响应式计算属性）
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
   * 处理数据文件合并按钮点击事件
   */
  const handleDbMergeClick = () => {
    $router.push('/main/db-merge');
  };

  /**
   * 处理数据文件合并按钮鼠标悬停事件
   */
  const handleDbMergeMouseOver = (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    target.style.backgroundColor = '#1A5284';
    target.style.color = '#ffffff';
  };

  /**
   * 处理数据文件合并按钮鼠标离开事件
   */
  const handleDbMergeMouseLeave = (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    const isSelected = route.path === '/main/db-merge';
    target.style.backgroundColor = isSelected ? '#1A5284' : '#ffffff';
    target.style.color = isSelected ? '#ffffff' : '#000000';
  };

  /**
   * 数据文件转Excel按钮样式（响应式计算属性）
   */
  const dbToExcelButtonStyle = computed(() => {
    const isSelected = route.path === '/main/db-to-excel';
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
   * 处理数据文件转Excel按钮点击事件
   */
  const handleDbToExcelClick = () => {
    $router.push('/main/db-to-excel');
  };

  /**
   * 处理数据文件转Excel按钮鼠标悬停事件
   */
  const handleDbToExcelMouseOver = (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    target.style.backgroundColor = '#1A5284';
    target.style.color = '#ffffff';
  };

  /**
   * 处理数据文件转Excel按钮鼠标离开事件
   */
  const handleDbToExcelMouseLeave = (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    const isSelected = route.path === '/main/db-to-excel';
    target.style.backgroundColor = isSelected ? '#1A5284' : '#ffffff';
    target.style.color = isSelected ? '#ffffff' : '#000000';
  };

  /**
   * 使用 watchEffect 监听 manifestState 的变化
   * 当其他组件更新 state.json 后，通过重新获取状态来保持同步
   */
  const refreshManifestState = async () => {
    await getManifestState();
  };
  
  // 暴露刷新方法给子组件使用
  provide('refreshManifestState', refreshManifestState);

  const areas = ref<string[]>([]);
  onMounted(async () => {
    // 获取区域配置
    const areaResult = await GetAreaConfig();
    if (areaResult.ok && areaResult.data) {
      const data = areaResult.data;
      areas.value = [data.province_name, data.city_name, data.country_name].filter(v => v);
    }
    
    // 初始获取 manifest 状态
    await getManifestState();
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
    &:hover:not(.ant-menu-item-disabled) {
      background-color: #1a5284;
      color: #ffffff;
    }
    &:hover {
      background-color: #1a5284;
      color: #ffffff;
    }
    
    /* 禁用状态样式 */
    &.disabled-menu-item {
      opacity: 0.5;
      cursor: pointer;
      background-color: #f5f5f5 !important;
      
      &:hover {
        background-color: #f5f5f5 !important;
        color: #999999 !important;
      }
    }
  }
  
  /* 禁用文本样式 */
  .disabled-text {
    color: #999999;
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
<style>
  .ant-btn-primary {
    background-color: #1a5284 !important;
    border-color: #1a5284 !important;
  }

  .ant-btn-primary:hover {
    background-color: #0f3a5f !important;
    border-color: #0f3a5f !important;
  }

  .ant-btn-primary:focus {
    background-color: #1a5284 !important;
    border-color: #1a5284 !important;
  }
</style>