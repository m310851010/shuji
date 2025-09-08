<template>
  <!-- 头部区域 -->
  <div class="page-header">
    <a-row type="flex" justify="space-between" style="width: 100%">
      <div class="span-line">
        <a-radio-group v-model:value="checkTab" @change="handleCheckTabChange">
          <a-radio-button v-for="item in CheckTypeOptions" :key="item.value" :value="item.value" class="tab-button-check">
            {{ item.label }}
          </a-radio-button>
        </a-radio-group>
      </div>

      <a-radio-group v-model:value="tableTab" @change="handleTabChange">
        <a-radio-button v-for="item in TableOptions" :key="item.value" :value="item.value" class="tab-button">
          {{ item.label }}
        </a-radio-button>
      </a-radio-group>
    </a-row>
  </div>

  <transition :name="transitionName" mode="out-in" :css="true">
    <div class="page-content" :key="tableTab + checkTab">
      <template v-for="item in TableOptions">
        <DataCheckModel v-if="checkTab === CheckType.model && item.value === tableTab" v-model="modelCheckMap[item.value]" />
        <DataCheckManual v-if="checkTab === CheckType.manual && item.value === tableTab" :tableType="item.value" />
      </template>
    </div>
  </transition>
</template>

<script setup lang="tsx">
  import { CheckType, CheckTypeOptions, TableOptions, TableType, TableTypeName } from '@/views/constant';
  import DataCheckModel from './components/DataCheckModel.vue';
  import DataCheckManual from './components/DataCheckManual.vue';
  import {
    ModelDataCheckAttachment2,
    ModelDataCheckTable1,
    ModelDataCheckTable2,
    ModelDataCheckTable3,
    ModelDataCoverAttachment2,
    ModelDataCoverTable1,
    ModelDataCoverTable2,
    ModelDataCoverTable3
  } from '@wailsjs/go';

  let previousIndex: number = 0;
  const transitionName = ref('slide');

  const tableTab = ref<TableType>(TableType.table1);
  const checkTab = ref<CheckType>(CheckType.model);

  const handleCheckTabChange = (value: CheckType) => {
    transitionName.value = 'slide';
  };

  // 处理标签切换
  const handleTabChange = (value: 'enterprise' | 'device') => {
    const index = TableOptions!.findIndex(v => v.value === tableTab.value);
    transitionName.value = index > previousIndex ? 'slide' : 'slide-reverse';
    previousIndex = index;
  };

  const modelCheckMap = getModelCheckModel();

  function getModelCheckModel() {
    const data_1 = reactive({
      name: TableTypeName.table1,
      tableType: TableType.table1,
      isChecking: false,
      checkFunc: ModelDataCheckTable1,
      coverFunc: ModelDataCoverTable1
    });

    const data_2 = reactive({
      name: TableTypeName.table2,
      tableType: TableType.table2,
      isChecking: false,
      checkFunc: ModelDataCheckTable2,
      coverFunc: ModelDataCoverTable2
    });

    const data_3 = reactive({
      name: TableTypeName.table3,
      tableType: TableType.table3,
      isChecking: false,
      checkFunc: ModelDataCheckTable3,
      coverFunc: ModelDataCoverTable3
    });

    const data_4 = reactive({
      name: TableTypeName.attachment2,
      tableType: TableType.attachment2,
      isChecking: false,
      checkFunc: ModelDataCheckAttachment2,
      coverFunc: ModelDataCoverAttachment2
    });

    return {
      [TableType.table1]: data_1,
      [TableType.table2]: data_2,
      [TableType.table3]: data_3,
      [TableType.attachment2]: data_4
    } as Record<string, any>;
  }
</script>

<style scoped>
  .tab-button-check.ant-radio-button-wrapper {
    padding: 0 15px;
    border: 0 !important;
    font-size: 16px;
    font-weight: 600;
    &:before {
      content: none;
      background-color: transparent;
    }
    &:after {
      background-color: transparent;
      content: none;
    }
  }

  .span-line {
    position: relative;
    margin-right: 20px;
    &:before {
      content: '';
      position: absolute;
      left: 50%;
      transform: translate(-50%, -50%);
      z-index: 2;
      top: 50%;
      font-size: 18px;
    }
  }
</style>
