<template>
  <!-- 头部区域 -->
  <div class="page-header header-tip">说明：附表1、附表2和区域表之间进行关联校验，请务必同时上传excel表格数据。</div>

  <div class="page-content">
    <div class="box-grey">
      <div class="bottom-line">
        <span class="title">数据导入及校验</span>
        <a-button
          v-if="!model.checkFinished"
          type="primary"
          @click="handleCheckClick"
          :loading="model.isChecking"
          :disabled="!model.selectedFiles?.length"
        >
          校验
        </a-button>
        <a-button v-else @click="handleBackClick">返回</a-button>
      </div>

      <!-- 文件导入区域 -->
      <UploadComponent v-if="!model.checkFinished" v-model="model.selectedFiles">
        <div>提示：一次只能导入一个省份的4个excel文件（.xlsx/.xls），支持批量选择，并且与选择的省份保持一致。</div>
        <div>选择文件后，点击上方按钮开始校验。</div>
        <div>数据校验完成后，会提示下载校验结果，自行下载压缩包后，打开excel检查单元格是否有标红色、黄色，按单元格的批注进行修改。</div>
      </UploadComponent>

      <div v-else class="box-grey no-bg" style="height: 400px">
        <div v-if="model.passed == null">
          <h1 style="text-align: center; margin-top: 100px; color: #999">点击上面“校验”按钮开始自动校验</h1>
        </div>

        <a-row type="flex" v-else align="middle" justify="space-between" style="flex-direction: column" class="h-100">
          <div style="font-size: 24px" :style="{ color: model.passed ? '#52c41a' : '#ff4d4f' }">
            数据{{ model.passed ? '已' : '未' }}通过自动校验
          </div>

          <div v-if="model.errorMessage" style="max-height: 255px; width: 100%; overflow: auto">
            <div style="white-space: break-spaces; line-height: 25px" v-html="model.errorMessage"></div>
          </div>
          <div>
            <a-button type="primary" v-if="model.canDownloadReport" @click="handleDownloadReport">导出校验报告</a-button>
          </div>
        </a-row>
      </div>
    </div>
  </div>
</template>

<script setup lang="tsx">
  import { openInfoModal, openModal } from '@/components/useModal';
  import UploadComponent from './components/Upload.vue';
  import { GetCachePath, Removefile } from '@wailsjs/go';
  import { main } from '@wailsjs/models';
  import { TableTypeName } from '@/views/constant';

  const model = reactive({
    isChecking: false,
    checkFinished: false,
    tableType: '',
    selectedFiles: [],
    passed: false,
    errorMessage: `const handleBackClick = async () => {
   `,
    canDownloadReport: true
  });

  const handleDownloadReport = async () => {};

  const handleBackClick = async () => {
    model.passed = null;
    model.checkFinished = false;
    const cachePath = await GetCachePath(model.tableType);
    // @ts-ignore
    await Removefile(cachePath + '/' + TableTypeName[model.tableType] + '校验报告.zip');
  };

  const handleCheckClick = async () => {
    const handleResult = (result: main.QueryResult) => {
      const data = result.data || {};
      model.canDownloadReport = data.hasExportReport;
      model.passed = !data.hasFailedFiles;
      model.isChecking = false;
      model.errorMessage = (result.message || '').replace(/\n/g, '<br>');
      model.checkFinished = true;
    };

    model.isChecking = true;
    const result = await model.checkFunc();
    console.log('自动校验结果', result);
    if (!result.ok) {
      model.isChecking = false;
      openInfoModal({
        title: '校验失败',
        content: result.message
      });
      return;
    }
    handleResult(result);
  };
</script>

<style scoped>
  .header-tip {
    color: red;
    font-size: 14px;
  }
</style>
