<template>
  <a-flex justify="flex-end" style="margin-bottom: 10px">
    <a-button type="primary" @click="handleCheckClick" :loading="model.isChecking">校验</a-button>
  </a-flex>

  <div class="box-grey no-bg" style="height: 340px">
    <a-flex align="center" justify="space-between" :vertical="true" class="h-100">
      <div></div>
      <div style="font-size: 24px" :style="{ color: passed ? '#52c41a' : '#ff4d4f' }">数据{{ passed ? '已' : '未' }}通过模型校验</div>
      <div>
        <a-button type="primary" v-if="!passed" @click="handleDownloadReport">下载模型报告</a-button>
      </div>
    </a-flex>
  </div>
</template>

<script setup lang="tsx">
import {openInfoModal, openModal} from '@/components/useModal';
  import TodoCoverTable from './TodoCoverTable.vue';
import {getFileName} from '@/util';
import {ModelDataCheckReportDownload} from '@wailsjs/go';

  const passed = ref(false);

  const model = defineModel({
    type: Object,
    default: () => ({
      checkFunc: null,
      coverFunc: null,
      isChecking: false,
      tableType: ''
    })
  });

  const handleDownloadReport = async () => {
    const result = await ModelDataCheckReportDownload(model.value.tableType);
    console.log(result);
  }


  const handleCheckClick = async () => {
    const handleResult = (hasFailedFiles: boolean) => {
      passed.value = !hasFailedFiles;
      model.value.isChecking = false;
    }

    model.value.isChecking = true;
    const result = await  model.value.checkFunc();
    console.log("模型校验结果", result);
    if (!result.ok) {
      model.value.isChecking = false;
     openInfoModal({
       title: '校验失败',
       content: result.message
     })
      return;
    }
    const {cover_files, hasFailedFiles} = result.data;

    // 有覆盖文件
    if (cover_files.length) {
      const confirmCoverList = cover_files.map((f: string) => ({
        fileName: getFileName(f),
        fullPath: f,
      }));

      let todoCoverList: string[] = [];
      return openModal({
        width: 800,
        title: '文件覆盖确认',
        content: () => (
            <>
              <h3 style="color: #f5222d;margin-bottom:15px;text-align:center">以下文件已被导入，是否覆盖？</h3>
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
          handleResult(hasFailedFiles);
        },
        onCancel: async () => {
          await model.value.coverFunc([]);
          handleResult(hasFailedFiles);
        }
      });
    }

    handleResult(hasFailedFiles);
  }
</script>
