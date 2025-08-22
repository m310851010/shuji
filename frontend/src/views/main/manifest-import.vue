<template>
  <div class="page-header">
    <a-radio-group v-model:value="activeTab" @change="handleTabChange">
      <a-radio-button v-for="item in ManifestTypeOptions" :key="item.value" :value="item.value" class="tab-button">
        {{ item.label }}
      </a-radio-button>
    </a-radio-group>
  </div>

  <transition :name="transitionName" mode="out-in" :css="true">
    <div class="page-content" :key="activeTab">
      <template v-for="item in ManifestTypeOptions">
        <ManifestImportTab v-if="item.value === activeTab" v-model="dataMap[item.value]"></ManifestImportTab>
      </template>
    </div>
  </transition>
</template>

<script setup lang="ts">
  import ManifestImportTab from './components/ManifestImportTab.vue';
  import { ManifestType, ManifestTypeOptions } from '@/views/constant';
  import { ImportEnterpriseList, ImportKeyEquipmentList, ValidateEnterpriseListFile, ValidateKeyEquipmentListFile } from '@wailsjs/go';

  // 状态管理
  const activeTab = ref<ManifestType>(ManifestType.enterprise);
  let previousIndex: number = 0;
  const transitionName = ref('slide');

  const dataMap = {
    [ManifestType.enterprise]: reactive({
      selectedFiles: [],
      name: '企业清单',
      checkFunc: ValidateEnterpriseListFile,
      importFunc: ImportEnterpriseList
    }),
    [ManifestType.equipment]: reactive({
      selectedFiles: [],
      name: '装置清单',
      checkFunc: ValidateKeyEquipmentListFile,
      importFunc: ImportKeyEquipmentList
    })
  } as Record<string, any>;

  // 处理标签切换
  const handleTabChange = () => {
    const index = ManifestTypeOptions!.findIndex(v => v.value === activeTab.value);
    transitionName.value = index > previousIndex ? 'slide' : 'slide-reverse';
    previousIndex = index;
  };
</script>

<style scoped></style>
