<template>
  <div class="page-header">
    <a-radio-group v-model:value="tableTab" @change="handleTabChange">
      <a-radio-button v-for="item in TableOptions" :key="item.value" :value="item.value" class="tab-button">
        {{ item.label }}
      </a-radio-button>
    </a-radio-group>
  </div>

  <transition :name="transitionName" mode="out-in" :css="true">
    <div class="page-content" :key="tableTab">
      <template v-for="item in TableOptions">
        <DataImportTab v-if="item.value === tableTab" v-model="dataMap[tableTab]" />
      </template>
    </div>
  </transition>
</template>

<script setup lang="ts">
  import DataImportTab from './components/DataImportTab.vue';
  import { TableOptions, TableType, TableTypeName } from '@/views/constant';
  import { ValidateTable1File, ValidateTable2File, ValidateTable3File, ValidateAttachment2File } from '@wailsjs/go';

  const ImportTable1 = {},
    ImportTable2 = {},
    ImportTable3 = {},
    ImportAttachment2 = {};

  const tableTab = ref<TableType>(TableType.table1);
  const data_1 = reactive({
    selectedFiles: [],
    name: TableTypeName.table1,
    tableType: TableType.table1,
    checkFunc: ValidateTable1File,
    importFunc: ImportTable1
  });

  const data_2 = reactive({
    selectedFiles: [],
    name: TableTypeName.table2,
    tableType: TableType.table2,
    checkFunc: ValidateTable2File,
    importFunc: ImportTable2
  });

  const data_3 = reactive({
    selectedFiles: [],
    name: TableTypeName.table3,
    tableType: TableType.table3,
    checkFunc: ValidateTable3File,
    importFunc: ImportTable3
  });

  const data_4 = reactive({
    selectedFiles: [],
    name: TableTypeName.attachment2,
    tableType: TableType.attachment2,
    checkFunc: ValidateAttachment2File,
    importFunc: ImportAttachment2
  });

  const dataMap = {
    [TableType.table1]: data_1,
    [TableType.table2]: data_2,
    [TableType.table3]: data_3,
    [TableType.attachment2]: data_4
  } as Record<string, any>;

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
