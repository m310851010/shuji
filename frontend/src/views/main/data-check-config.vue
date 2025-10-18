<template>
  <div class="page-header"></div>
  <div class="page-content">
    <div v-for="(item, index) in data.configs" :key="item.key">
      <div class="form-item-title">{{ index + 1 }}、{{ item.title }}</div>
      <p>{{ item.desc }}</p>
      <div class="form-item">
        <label>预留差异空间：</label>
        <a-input-number
          class="form-item-input"
          v-model:value="item.value"
          :min="0"
          :max="100"
          :formatter="(value: any) => `${value || 0}%`"
          :parser="(value: any) => value.replace('%', '')"
          :precision="0"
        ></a-input-number>
        <span class="form-item-remark">注：默认0，表示不留空间，完全相等校验，如设置20%，表示差异在20%以内算合格</span>
      </div>
      <div v-if="index < data.configs.length - 1" class="form-item-divider"></div>
    </div>
  </div>

  <div class="page-footer">
    <a-space :size="20">
      <a-button type="primary" class="config-button" @click="saveConfig">保存</a-button>
      <a-button type="primary" class="config-button" @click="resetConfig">恢复默认</a-button>
    </a-space>
  </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue';

  const data = reactive({
    configs: [
      {
        key: 'table1Space',
        value: 0,
        defaultValue: 0,
        title: '规上企业煤炭消费信息附表1，预留差异空间配置',
        desc: '说明：除煤炭洗选和煤制品加工用途以外，应当确保“耗煤总量(实物量，万吨)”数值与“原煤消费(实物量，万吨)”“洗精煤消费(实物量，万吨)”“其他煤炭消费(实物量，万吨)”加和的数值相等。若超出预留空间以上的较大偏差，应在“备注”列中说明原因。'
      },
      {
        key: 'table1SpaceUse',
        value: 0,
        defaultValue: 0,
        title: '规上企业煤炭消费信息附表1主要用途情况，预留差异空间配置',
        desc: '说明：除煤炭洗选和煤制品加工用途以外，应当确保附表1煤炭消费主要信息部分，每个企业的“耗煤总量(实物量，万吨)”与附表1主要用途情况，企业的“投入量”之和相等。若超出预留空间以上的较大偏差，应在“备注”列中说明原因。'
      },
      {
        key: 'table1SpaceEq',
        value: 0,
        defaultValue: 0,
        title: '规上企业煤炭消费信息附表1重点耗煤装置（设备），预留差异空间配置',
        desc: '说明：除煤炭洗选和煤制品加工用途以外，应当确保附表1煤炭消费主要信息部分，每个企业的“耗煤总量(实物量，万吨)”与附表1重点耗煤装置（设备）情况这重点耗煤装置的年耗煤量之和相等。若超出预留空间以上的较大偏差，应在“备注”列中说明原因。'
      },
      {
        key: 'table1SpaceXf',
        value: 0,
        defaultValue: 0,
        title: '对于某一个区域（省、市、县）的煤炭消费数据，预留差异空间配置',
        desc: '说明：按照现行统计规则，除煤炭洗选和煤制品加工用途以外，对于某一个区域（省、市、区）的煤炭消费数据，本年度的“煤合计”与“原煤”“洗精煤”“其他”应该相等。若超出预留空间以上的较大偏差，应在“备注”列中说明原因。'
      },
      {
        key: 'table1SpaceXf',
        value: 0,
        defaultValue: 0,
        title: '对于某一个区域（省、市、县）的煤炭消费数据，预留差异空间配置',
        desc: '说明：按照现行统计规则，对于某一个区域（省、市、区）的煤炭消费数据，本年度的“煤合计”数值与“能源加工转换”和“终端消费”之和应该相等。若超出预留空间以上的较大偏差，应在“备注”列中说明原因。'
      }
    ]
  });

  const saveConfig = () => {
    data.configs.forEach(item => {
      item.defaultValue = item.value;
    });
    message.success('保存成功');
  };

  const resetConfig = () => {
    data.configs.forEach(item => {
      item.value = item.defaultValue;
    });
    message.success('恢复默认成功');
  };
</script>

<style scoped lang="less">
  :deep(.ant-input-number-input) {
    text-align: center;
  }
  .form-item {
    label {
      display: inline-block;
      width: 120px;
      font-size: 16px;
    }
  }
  .form-item-title {
    font-size: 18px;
  }
  .form-item-input {
    width: 150px;
    margin-right: 10px;
    font-size: 18px;
  }
  .form-item-divider {
    border-bottom: 1px solid #e8e8e8;
    margin: 10px 0;
  }
  .form-item-remark {
    font-size: 13px;
    color: #f00;
  }

  .config-button {
    width: 100px;
  }
</style>
