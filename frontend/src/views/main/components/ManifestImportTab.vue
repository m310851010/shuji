<template>
  <div class="box-grey">

    <div class="bottom-line">
      <span class="title">导入清单</span>

      <a-button type="primary" @click="handleUploadClick" :loading="model.isImporting">导入清单</a-button>
    
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
</template>

<script setup lang="tsx">
  import UploadComponent from './Upload.vue';
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
    if (!model.value.selectedFiles?.length) {
      openInfoModal({ content: '请选择文件' });
      return;
    }

    model.value.isImporting = true;

    try {
      // 批量处理文件
      for (let i = 0; i < model.value.selectedFiles.length; i++) {
        const file = model.value.selectedFiles[i];

        const fn = async () => {
          const importResult = await model.value.importFunc(file.fullPath);
          if (importResult.ok) {
            notification.success({
              placement: 'top',
              message: '导入成功',
              description: importResult.message,
              duration: 5
            });
          } else {
            notification.info({
              placement: 'top',
              message: '导入失败',
              description: importResult.message,
              duration: 5
            });
          }
        };

        const checkResult = await model.value.checkFunc(file.fullPath);
        if (checkResult.ok) {
          validationResult.value = '校验通过';
          //  已存在数据
          if (checkResult.data) {
            await new Promise(resolve => {
              openModal({
                content: `${model.value.name}数据已存在，是否替换？`,
                onOk: async () => {
                  await fn();
                  resolve(true);
                },
                onCancel: async () => {
                  resolve(true);
                }
              });
            });
          } else {
            await fn();
          }
        } else {
          notification.error({
            placement: 'top',
            message: '校验失败',
            description: checkResult.message,
            duration: 5
          });
        }
      }

      console.log('批量导入完成');
      model.value.isImporting = false;

      // 清空文件列表
      model.value.selectedFiles = [];
    } catch (error) {
      console.error('批量导入失败:', error);
      message.error('批量导入失败: ' + (error as Error).message);
    } finally {
      model.value.isImporting = false;
    }
  };
</script>

<style scoped>
  .result-text {
    font-weight: bold;
    color: #1890ff;
  }
  :deep(.ant-steps-item-title) {
    color: #000 !important;
  }
</style>
