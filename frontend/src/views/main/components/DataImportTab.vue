<template>
  <div class="box-grey">
    <div class="bottom-line">
      <span class="title">导入数据</span>
      <a-button type="primary" @click="handleUploadClick" :loading="model.isImporting" :disabled="!model.selectedFiles?.length">
        点击导入
      </a-button>
    </div>


    <div class="box-grey">
      <a-steps :current="-1" label-placement="vertical" :items="items"></a-steps>
    </div>

    <div v-if="showValidationResult" class="box-grey" style="padding-top: 5px; padding-bottom: 5px">
      校验结果：
      <span class="result-text">{{ validationResult }}</span>
    </div>

    <!-- 文件导入区域 -->
    <UploadComponent v-model="model.selectedFiles" />

  </div>


  <div class="box-grey">
    <div class="bottom-line title">导入记录</div>
    <a-table :dataSource="dataSource" :columns="columns"  bordered :pagination="false" />
  </div>
</template>

<script setup lang="tsx">
  import UploadComponent from './Upload.vue';
  import TodoCoverTable from './TodoCoverTable.vue';
  import ShowImportResult from './ShowImportResult.vue';
  import { TableColumnType } from 'ant-design-vue';
  import {getFileName, newColumns} from '@/util';
  import { openInfoModal, openModal } from '@/components/useModal';
  import { message, notification } from 'ant-design-vue';

  const model = defineModel({
    type: Object,
    default: () => ({
      selectedFiles: [],
      importFunc: null,
      checkFunc: null,
      isImporting: false
    })
  });

  const showValidationResult = ref(false);
  const validationResult = ref('格式异常');
  const confirmCoverList = ref<any[]>([]);
  const todoCoverList = ref<string[]>([]);

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



  // 处理导入按钮点击
  const handleUploadClick = async () => {
    todoCoverList.value = [];
    if (!model.value.selectedFiles?.length) {
      openInfoModal({ content: '请选择文件' });
      return;
    }

    model.value.isImporting = true;

    const checkResultList: Promise<any>[] = [];
    confirmCoverList.value = [];
         // 批量处理文件, 把处理结果放到一个数组中
  
      for (let i = 0; i < model.value.selectedFiles.length; i++) {
        const file = model.value.selectedFiles[i];
        checkResultList.push(model.value.checkFunc(file.fullPath, true).then((result: any) => {
          console.log(result);
          result.fullPath = file.fullPath;
          result.fileName = getFileName(file.fullPath);
          // 需要覆盖的文件
          if (!result.ok && result.data === 'FILE_EXISTS') {
            result.isCover = true;
            confirmCoverList.value.push(result);
          }
          return result;
        }));
      }

      let checkResults = await Promise.all(checkResultList);

      if (confirmCoverList.value.length) {
        return openModal({
          width: 800,
          content: () => (
            <>
              <div>
                以下文件已存在，是否覆盖？
              </div>
              <div style="height: 350px; overflow: auto">
                <TodoCoverTable fileList={confirmCoverList.value} onUpdateFileList={(val: any) => {
                  todoCoverList.value = val
                }} />
              </div>
            </>
          ),
          onOk: async () => {
            if (todoCoverList.value.length) {
              await Promise.all(todoCoverList.value.map(item => {
                return model.value.checkFunc(item, false).then((ret: any) => {
                  checkResults.forEach((it, i) => {
                    if (it.fullPath === item && it.isCover) {
                      ret.fileName = it.fileName;
                      checkResults.splice(i, 1, ret);
                    }
                  });
                });
              }));
            }

            // 清空文件列表
            model.value.selectedFiles = [];
            model.value.isImporting = false;
            checkResults = checkResults.filter((it: any) => !it.isCover);
            if (checkResults.length) {
              showImportResult(checkResults);
            }

          },
          onCancel: async () => {
            checkResults = checkResults.filter((it: any) => !it.isCover);
            // 清空文件列表
            model.value.selectedFiles = [];
            model.value.isImporting = false;
            if (checkResults.length) {
              showImportResult(checkResults);
            }
          }
        });
      }

    // 清空文件列表
    model.value.selectedFiles = [];
    model.value.isImporting = false;
    showImportResult(checkResults);
  };

  function showImportResult(checkResults: any[]) {
    openInfoModal({
      width: 800,
      content: () => <ShowImportResult style={{height: '400px', overflow: 'auto'}} resultList={checkResults} />
    })
  }

  const dataSource = Array.from({ length: 5 }).fill({
    key: '1',
    enterpriseName: '内蒙古伊核公司',
    age: 32
  });

  const columns: TableColumnType[] = newColumns({ stat_date: '文件名', impDate: '导入时间', impStatus: '导入状态', comment: '说明' });
</script>

<style scoped>
  :deep(.ant-steps-item-title) {
    color: #000 !important;
  }
</style>
