<template>
  <a-config-provider :locale="zhCN">
    <RouterView />
  </a-config-provider>
</template>

<script lang="tsx" setup>
  import { RouterView } from 'vue-router';
  import zhCN from 'ant-design-vue/es/locale/zh_CN';
  import dayjs from 'dayjs';
  import 'dayjs/locale/zh-cn';

  import { EventsOn } from '@wailsapp/runtime';
  import { ExitApp } from '@wailsjs/go';
  import { useFileDrop, useSupportFileDrop } from '@/hook/useFileDrop';
  import { userGlobalDragAndDrop } from '@/util/preventDragAndDrop';

  dayjs.locale('zh-cn');
  EventsOn('onBeforeClose', async () => {
    await ExitApp();
  });

  useFileDrop();
  const isSupport = useSupportFileDrop();
  const { addEvent, removeEvent } = userGlobalDragAndDrop();

  const listenerDrop = (val: boolean) => {
    if (val) {
      console.log('移除事件');
      removeEvent();
    } else {
      console.log('添加事件');
      addEvent();
    }
  };

  onMounted(() => {
    listenerDrop(isSupport.value);
  });

  watch(isSupport, newVal => {
    console.log(`watch isSupport=${isSupport} newVal=${newVal}`);
    listenerDrop(newVal);
  });
</script>

<style></style>
