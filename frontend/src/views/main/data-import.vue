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
  import { TableOptions, TableType } from '@/views/constant';
  import {
    ImportTable1,
    ImportTable2,
    ImportTable3,
    ImportAttachment2,
    ValidateTable1File,
    ValidateTable2File,
    ValidateTable3File,
    ValidateAttachment2File
  } from '@wailsjs/go';

  const tableTab = ref<TableType>(TableType.table1);
  const data_1 = reactive({
    selectedFiles: [],
    name: '表1',
    checkFunc: ValidateTable1File,
    importFunc: ImportTable1
  });

  const data_2 = reactive({
    selectedFiles: [],
    name: '表2',
    checkFunc: ValidateTable2File,
    importFunc: ImportTable2
  });

  const data_3 = reactive({
    selectedFiles: [],
    name: '表3',
    checkFunc: ValidateTable3File,
    importFunc: ImportTable3
  });

  const data_4 = reactive({
    selectedFiles: [],
    name: '附件2',
    checkFunc: ValidateAttachment2File,
    importFunc: ImportAttachment2
  });

  const dataMap = {
    [TableType.table1]: data_1,
    [TableType.table2]: data_2,
    [TableType.table3]: data_3,
    [TableType.attachment2]: data_4
  } as Record<string, any>;

  const currentStep = ref(0);
  const selectedFiles = ref<File[]>([
    { name: '附表3 固定资产投资项目节能审查煤炭消费情况汇总表.xlsx' },
    { name: '附表3 固定资产投资项目节能审查煤炭消费情况汇总表.xlsx' }
  ] as unknown as File[]);
  const showValidationResult = ref(false);
  const validationResult = ref('格式异常');

  let previousIndex: number = 0;
  const transitionName = ref('slide');

  const items = ref([
    {
      title: '文件导入',
      description: '选择并导入文件'
    },
    {
      title: '文件校验',
      description: '验证文件格式和内容'
    },
    {
      title: '导入完成',
      description: '文件导入成功'
    }
  ]);

  // 处理标签切换
  const handleTabChange = () => {
    const index = TableOptions!.findIndex(v => v.value === tableTab.value);
    transitionName.value = index > previousIndex ? 'slide' : 'slide-reverse';
    previousIndex = index;
  };

  // 处理导入按钮点击
  const handleUploadClick = () => {
    if (selectedFiles.value) {
      // 模拟文件校验过程
      currentStep.value = 1;
      showValidationResult.value = true;

      // 模拟校验后导入完成
      setTimeout(() => {
        currentStep.value = 2;
      }, 1500);
    }
  };

  // 处理文件选择
  const handleFilesSelected = (files: FileList) => {
    // selectedFiles.value = files;
    currentStep.value = 0;
    showValidationResult.value = false;
  };

  // 处理导入成功
  const handleUploadSuccess = () => {
    currentStep.value = 2;
  };
</script>

<style scoped></style>
