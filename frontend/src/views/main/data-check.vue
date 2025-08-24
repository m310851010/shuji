<template>
  <!-- 头部区域 -->
  <div class="page-header">
    <a-flex justify="space-between">
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
    </a-flex>
  </div>

  <transition :name="transitionName" mode="out-in" :css="true">
    <div class="page-content" :key="tableTab + checkTab">
      <template v-for="item in TableOptions">
        <DataCheckModel v-if="checkTab === CheckType.model && item.value === tableTab" />
        <DataCheckManual v-if="checkTab === CheckType.manual && item.value === tableTab" />
      </template>
    </div>
  </transition>
</template>

<script setup lang="tsx">
  import { CheckType, CheckTypeOptions, TableOptions, TableType } from '@/views/constant';
  import DataCheckModel from './components/DataCheckModel.vue';
  import DataCheckManual from './components/DataCheckManual.vue';

  let previousIndex: number = 0;
  const transitionName = ref('slide');

  const tableTab = ref<TableType>(TableType.table1);
  const checkTab = ref<CheckType>(CheckType.model);

  // 状态管理
  const activeTab = ref<'enterprise' | 'device'>('enterprise');
  const currentStep = ref(0);

  const handleCheckTabChange = (value: CheckType) => {
    transitionName.value = 'slide';
    previousIndex = 0;
    tableTab.value = TableType.table1;
  };

  // 处理标签切换
  const handleTabChange = (value: 'enterprise' | 'device') => {
    const index = TableOptions!.findIndex(v => v.value === tableTab.value);
    transitionName.value = index > previousIndex ? 'slide' : 'slide-reverse';
    previousIndex = index;
  };
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
