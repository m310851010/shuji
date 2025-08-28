<template>
  <div class="wh-100 flex-vertical">
    <div class="page-header">
      <a-radio-group v-model:value="tableTab" @change="handleTabChange">
        <a-radio-button v-for="item in TableOptions" :key="item.value" :value="item.value" class="tab-button">
          {{ item.label }}
        </a-radio-button>
      </a-radio-group>
    </div>

    <transition :name="transitionName" mode="out-in" :css="true">
      <div class="page-content flex-vertical" :key="tableTab">
        <template v-for="item in TableOptions">
          <ImportProcessTab v-if="item.value === tableTab" :tableType="item.value" />
        </template>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
  import { TableOptions, TableType } from '@/views/constant';
  import ImportProcessTab from './components/ImportProcessTab.vue';

  const tableTab = ref<TableType>(TableType.table1);

  let previousIndex: number = 0;
  const transitionName = ref('slide');

  // 处理标签切换
  const handleTabChange = () => {
    const index = TableOptions!.findIndex(v => v.value === tableTab.value);
    transitionName.value = index > previousIndex ? 'slide' : 'slide-reverse';
    previousIndex = index;
  };
</script>

<style scoped></style>
