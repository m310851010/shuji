<template>
  <a-config-provider
    :locale="zhCN"
    :theme="{
      algorithm: [theme.defaultAlgorithm],
      token: {
        colorPrimary: '#0078d7',
        borderRadius: '2px',
        wireframe: true,
        colorBgMask: '#00000036'
      }
    }"
  >
    <a-style-provider hash-priority="high" :transformers="[legacyLogicalPropertiesTransformer]">
      <RouterView />
    </a-style-provider>
  </a-config-provider>
</template>

<script lang="tsx" setup>
  import { RouterView } from 'vue-router';
  import { legacyLogicalPropertiesTransformer, theme, Button, Alert } from 'ant-design-vue';
  import zhCN from 'ant-design-vue/es/locale/zh_CN';
  import dayjs from 'dayjs';
  import 'dayjs/locale/zh-cn';

  import { EventsOn, OnFileDropOff } from '@wailsapp/runtime';
  import { ExitApp } from '@wailsjs/go';
  import { useFileDrop } from '@/hook/useFileDrop';

  dayjs.locale('zh-cn');
  EventsOn('onBeforeClose', async () => {
    await ExitApp();
  });

  onMounted(() => {
    useFileDrop();
  });
  onUnmounted(() => {
    OnFileDropOff();
  });
</script>

<style></style>
