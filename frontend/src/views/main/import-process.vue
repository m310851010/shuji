<template>
  <div class="wh-100 flex-vertical">
    <div class="page-header">
      <a-radio-group v-model:value="tableTab" @change="handleTabChange">
        <a-radio-button v-for="item in tableOptions" :key="item.value" :value="item.value" class="tab-button">
          {{ item.label }}
        </a-radio-button>
      </a-radio-group>
    </div>

    <transition :name="transitionName" mode="out-in" :css="true">
      <div class="page-content flex-vertical" :key="tableTab">
        <template v-for="item in tableOptions">
          <ImportProcessTab v-if="item.value === tableTab" :tableType="item.value" />
        </template>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
  import { TableOptions, TableType } from '@/views/constant';
  import ImportProcessTab from './components/ImportProcessTab.vue';
  import { GetAreaConfig } from '@wailsjs/go';

  const tableTab = ref<TableType>(TableType.table1);
  const tableOptions = ref<any>(TableOptions);

  onMounted(async () => {
    const areaResult = await GetAreaConfig();
    if (areaResult.ok && areaResult.data?.country_name) {
      // @ts-ignore
      const [t, t2] = TableOptions;
      tableOptions.value = [t, t2];
    }
  });

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
