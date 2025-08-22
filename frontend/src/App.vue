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

  import { EventsOn } from '@wailsapp/runtime';
  import { ExitApp, GetAreaConfig, GetCurrentOSUser } from '@wailsjs/go';

  dayjs.locale('zh-cn');

  GetCurrentOSUser().then(res => {
    console.log(res);
  });
  EventsOn('onBeforeClose', async () => {
    await ExitApp();
  });
</script>

<style></style>
