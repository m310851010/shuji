<template>
  <a-row type="flex"  justify="end" style="margin-bottom: 10px">
    <a-button v-if="!model.checkFinished" type="primary" @click="handleCheckClick" :loading="model.isChecking">校验</a-button>
    <a-button v-else @click="handleBackClick">返回</a-button>
  </a-row>

  <div class="box-grey no-bg" style="height: 340px">
    <div v-if="model.passed == null">
      <h1 style="text-align: center; margin-top: 100px; color: #999">点击上面“校验”按钮开始自动校验</h1>
    </div>

    <a-row type="flex" v-else align="center" justify="space-between" :vertical="true" class="h-100">
      <div></div>
      <div style="font-size: 24px" :style="{ color: model.passed ? '#52c41a' : '#ff4d4f' }">
        数据{{ model.passed ? '已' : '未' }}通过自动校验
      </div>
      <div v-if="model.errorMessage" style="max-height: 255px; overflow: auto">
        <pre style="white-space: break-spaces; line-height: 25px">{{ model.errorMessage }}</pre>
      </div>
      <div>
        <a-button type="primary" v-if="model.canDownloadReport" @click="handleDownloadReport">导出校验报告</a-button>
      </div>
    </a-row>
  </div>
</template>

<script setup lang="tsx">
  import { openInfoModal, openModal } from '@/components/useModal';
  import TodoCoverTable from './TodoCoverTable.vue';
  import { getFileName } from '@/util';
  import { GetCachePath, ModelDataCheckReportDownload, Removefile } from '@wailsjs/go';
  import { db, main } from '@wailsjs/models';
  import { TableTypeName } from '@/views/constant';

  const model = defineModel({
    type: Object,
    default: () => ({
      checkFunc: null,
      coverFunc: null,
      isChecking: false,
      checkFinished: false,
      tableType: ''
    })
  });

  const handleDownloadReport = async () => {
    await ModelDataCheckReportDownload(model.value.tableType);
  };

  const handleBackClick = async () => {
    model.value.passed = null;
    model.value.checkFinished = false;
    const cachePath = await GetCachePath(model.value.tableType);
    // @ts-ignore
    await Removefile(cachePath + '/' + TableTypeName[model.value.tableType] + '校验报告.zip');
  };

  const handleCheckClick = async () => {
    const handleResult = (result: db.QueryResult) => {
      const data = result.data || {};
      model.value.canDownloadReport = data.hasExportReport;
      model.value.passed = !data.hasFailedFiles;
      model.value.isChecking = false;
      model.value.errorMessage = result.message;
      model.value.checkFinished = true;
    };

    model.value.isChecking = true;
    const result = await model.value.checkFunc();
    console.log('自动校验结果', result);
    if (!result.ok) {
      model.value.isChecking = false;
      openInfoModal({
        title: '校验失败',
        content: result.message
      });
      return;
    }
    const { cover_files } = result.data;

    // 有覆盖文件
    if (cover_files.length) {
      const confirmCoverList = cover_files.map((f: string) => ({
        fileName: getFileName(f),
        fullPath: f
      }));

      let todoCoverList: string[] = [];
      return openModal({
        width: 800,
        title: '文件覆盖确认',
        content: () => (
          <>
            <h3 style="color: #faad14;margin-bottom:15px;text-align:center">以下文件包含重复数据，请确认数据是否覆盖？</h3>
            <div style="max-height: 350px; overflow: auto">
              <TodoCoverTable
                fileList={confirmCoverList}
                onUpdateFileList={(val: any) => {
                  todoCoverList = val;
                }}
              />
            </div>
          </>
        ),
        onOk: async () => {
          await model.value.coverFunc(todoCoverList);
          handleResult(result);
        },
        onCancel: async () => {
          await model.value.coverFunc([]);
          handleResult(result);
        }
      });
    }

    handleResult(result);
  };
</script>
