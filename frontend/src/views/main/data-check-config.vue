<template>
  <div class="page-header"></div>
  <div class="page-content">
    <div v-for="(item, index) in data.configs" :key="index">
      <div class="form-item-title">{{ index + 1 }}、{{ item.title }}</div>
      <p>{{ item.desc }}</p>
      <div class="form-item" v-for="field in item.fields" :key="field.key">
        <label>{{ field.label }}：</label>
        <a-input-number
          class="form-item-input"
          v-model:value="field.value" 
          :min="0"
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
import { GetAppConfig, WriteAppConfig } from '@wailsjs/go';
import { main } from '@wailsjs/models';

  const data = reactive({
    configs: [
      {
        title: '规上企业煤炭消费信息附表1，预留差异空间配置',
        desc: '说明：除煤炭洗选和煤制品加工用途以外，应当确保“耗煤总量(实物量，万吨)”数值与“原煤消费(实物量，万吨)”“洗精煤消费(实物量，万吨)”“其他煤炭消费(实物量，万吨)”加和的数值相等。若超出预留空间以上的较大偏差，应在“备注”列中说明原因。',
        fields: [{key: 'thresholdTotalCoalConsumption', label: '预留差异空间', value: 0}]
      },
      {
        key: 'thresholdMainUsage',
        title: '规上企业煤炭消费信息附表1主要用途情况，预留差异空间配置',
        desc: '说明：除煤炭洗选和煤制品加工用途以外，应当确保附表1煤炭消费主要信息部分，每个企业的“耗煤总量(实物量，万吨)”与附表1主要用途情况，企业的“投入量”之和相等。若超出预留空间以上的较大偏差，应在“备注”列中说明原因。',
        fields: [{key: 'thresholdMainUsage', label: '预留差异空间', value: 0}]
      },
      {
        title: '规上企业煤炭消费信息附表1重点耗煤装置（设备），预留差异空间配置',
        desc: '说明：除煤炭洗选和煤制品加工用途以外，应当确保附表1煤炭消费主要信息部分，每个企业的“耗煤总量(实物量，万吨)”与附表1重点耗煤装置（设备）情况这重点耗煤装置的年耗煤量之和相等。若超出预留空间以上的较大偏差，应在“备注”列中说明原因。',
        fields: [{key: 'thresholdCoalEquipment', label: '预留差异空间', value: 0}]
      },
      {
        title: '对于某一个区域（省、市、县）年度煤合计数值，预留差异空间配置',
        desc: '说明：按照现行统计规则，除煤炭洗选和煤制品加工用途以外，对于某一个区域（省、市、区）的煤炭消费数据，本年度的“煤合计”与“原煤”“洗精煤”“其他”应该相等。若超出预留空间以上的较大偏差，应在“备注”列中说明原因。',
        fields: [{key: 'thresholdTotalCoal_province', label: '省预留差异空间', value: 0}, {key: 'thresholdTotalCoal_city', label: '市预留差异空间', value: 0}, {key: 'thresholdTotalCoal_country', label: '县预留差异空间', value: 0}]
      },
      {
        title: '对于某一个区域（省、市、县）的煤炭消费品种数据，预留差异空间配置',
        desc: '说明：按照现行统计规则，对于某一个区域（省、市、区）的煤炭消费数据，本年度的“煤合计”数值与“能源加工转换”和“终端消费”之和应该相等。若超出预留空间以上的较大偏差，应在“备注”列中说明原因。',
        fields: [{key: 'thresholdEnergyTypes', label: '预留差异空间', value: 0}]
      },
      {
        title: '对于某一个区域（省、市、县）的煤炭消费数据，预留差异空间配置',
        desc: '说明：按照现行统计规则，对于某一个区域（省、市、区）的煤炭消费数据，本年度的“煤合计”数值与“能源加工转换”和“终端消费”之和应该相等。若超出预留空间以上的较大偏差，应在“备注”列中说明原因。',
        fields: [{key: 'thresholdCoalConsumption', label: '预留差异空间', value: 0}]
      },
      {
        title: '对于某一个区域（省、市）的煤炭消费数据，本年度“煤合计”数值与下辖所有区域（市、县）“煤合计”数值，预留差异空间配置',
        desc: '说明：对于某一个区域（省、市）的煤炭消费数据，本年度“煤合计”数值与下辖所有区域（市、县）“煤合计”数值之和应当近似相等。若超出预留空间以上的较大偏差，应在“备注”列中说明原因。',
        fields: [{key: 'thresholdTotalCoal_area_province', label: '省预留差异空间', value: 0}, {key: 'thresholdTotalCoal_area_city', label: '市预留差异空间', value: 0}, {key: 'thresholdTotalCoal_area_country', label: '县预留差异空间', value: 0}]
      }
    ]
  });

  const setConfig = (config: any) => {
    data.configs.forEach(item => {
      item.fields.forEach(field => {
         // @ts-ignore
        field.value = config[field.key] * 100;;
      });
    });
  };

  onMounted(async () => {
    const config = await GetAppConfig();
    if (!config) {
      message.error('获取应用配置失败');
      return;
    }
    setConfig(config);
  });
  

  // 设置配置
  const writeAppConfig = async (config: Record<string, number>) => {
    const _config = new main.AppConfig(config);
    const result = await WriteAppConfig(_config);
    if (!result.ok) {
      message.error(result.message);
      return false;
    }
    return true;
  };

  // 保存配置
  const saveConfig = async () => {
    const config: Record<string, number> = {};
    data.configs.forEach(item => {
      item.fields.forEach(field => {
        config[field.key] = field.value / 100;
      });
    });
    const success = await writeAppConfig(config);
    if (success) {
     message.success('保存成功');
    }
  };

  // 恢复默认配置
  const resetConfig = async() => {
    const config: Record<string, number> = {
       thresholdTotalCoalConsumption: 0,
       thresholdMainUsage: 0,
       thresholdCoalEquipment: 0,
       thresholdTotalCoal_province: 0,
       thresholdTotalCoal_city: 0,
       thresholdTotalCoal_country: 0,
       thresholdEnergyTypes: 0,
       thresholdCoalConsumption: 0,
       thresholdTotalCoal_area_province: 0,
       thresholdTotalCoal_area_city: 0,
       thresholdTotalCoal_area_country: 0
    }
    const success = await writeAppConfig(config);
    if (success) {
      message.success('恢复默认成功');
      setConfig(config);
    }
  };
</script>

<style scoped lang="less">
  :deep(.ant-input-number-input) {
    text-align: center;
  }
  .form-item {
    margin-bottom: 8px;
    label {
      display: inline-block;
      width: 130px;
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
