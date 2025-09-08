<template>
  <a-config-provider
    :locale="zhCN"
  >
      <RouterView />
  </a-config-provider>
</template>

<script lang="tsx" setup>
  import { RouterView } from 'vue-router';
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
