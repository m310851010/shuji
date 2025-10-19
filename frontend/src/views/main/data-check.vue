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
  import { openInfoModal } from '@/components/useModal';
  import { message } from 'ant-design-vue';
  import UploadComponent from './components/Upload.vue';
  import { Copyfile, GetCachePath, OpenSaveDialog, Removefile, ValidateData } from '@wailsjs/go';
  import { main } from '@wailsjs/models';
  import { CurrentArea } from '../constant';


  const model = reactive({
    isChecking: false,
    checkFinished: false,
    zipPath: '',
    selectedFiles: [],
    passed: false,
    errorMessage: '',
    canDownloadReport: false
  });

  const handleDownloadReport = async () => {
    if (!model.zipPath) {
      return;
    }

    const result = await OpenSaveDialog(new main.FileDialogOptions({
      title: '导出校验报告',
      defaultFilename: '校验报告.zip',
    }));

    if (result.canceled) {
      return;
    }

    const copyResult = await Copyfile(model.zipPath, result.filePaths[0]);
    if (!copyResult.ok) {
      openInfoModal({ content: copyResult.data });
      return;
    }
    message.success('导出校验报告成功');
  };

  const handleBackClick = async () => {
    model.passed = false;
    model.checkFinished = false;
    // @ts-ignore
    await Removefile(model.zipPath);
  };

  const handleCheckClick = async () => {
    const handleResult = (result: main.QueryResult) => {
      const data = result.data;
      model.canDownloadReport = result.ok && data;
      // 如果校验通过，并且没有错误，则认为数据通过
      model.passed = result.ok && !data;
      model.isChecking = false;
      model.errorMessage = (result.message || '').replace(/\n/g, '<br>');
      model.checkFinished = true;
      model.zipPath = result.ok ? data as string : '';
    };

    model.isChecking = true;
    model.zipPath = '';
    const result = await ValidateData(CurrentArea.province, model.selectedFiles.map((file: any) => file.fullPath));
    handleResult(result);
  };
</script>

<style scoped>
  .header-tip {
    color: red;
    font-size: 14px;
  }
</style>
