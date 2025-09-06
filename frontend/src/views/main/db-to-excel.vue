<template>
  <div class="wh-100 flex-vertical">
    <div class="page-header">
      <span class="header-title">数据文件转Excel</span>
    </div>
    <div class="page-content text-center">
      <UploadComponent
        v-model="selectedFiles"
        v-on:update:model-value="handleUpdateModelValue"
        :accept="() => true"
        :validFile="['db']"
        filterName="数据文件"
        filterPattern="*.db"
        title="选择数据文件"
      >
        <div>只能选择数据文件（.db），支持批量选择</div>
        <div>支持一次性拖一个或多个数据文件，以及整个文件夹</div>
        <div>选择文件后，点击下方按钮开始转换为Excel</div>
      </UploadComponent>

      <div class="operation-area">
        <a-button type="primary" @click="handleConvert" :loading="isConverting">转换为Excel</a-button>
      </div>

      <!-- 转换进度显示 -->
      <div v-if="isConverting" class="progress-section">
        <a-progress :percent="convertProgress" :status="convertStatus" />
        <div class="progress-text">{{ progressText }}</div>
      </div>

      <!-- 转换结果显示 -->
      <div v-if="convertResults.length > 0" class="result-section">
        <h3>转换结果</h3>
        <a-table :dataSource="convertResults" :columns="resultColumns" :pagination="false" size="small" bordered>
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'status'">
              <a-tag :color="record.success ? 'green' : 'red'">
                {{ record.success ? '成功' : '失败' }}
              </a-tag>
            </template>
            <template v-if="column.key === 'action'">
              <a-button v-if="record.success" type="link" @click="saveAsFile(record)">保存文件</a-button>
            </template>
          </template>
        </a-table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive } from 'vue';
  import { message } from 'ant-design-vue';
  import UploadComponent from './components/Upload.vue';
  import { DBTranformExcel, Movefile, OpenSaveDialog } from '@wailsjs/go';
  import { main } from '@wailsjs/models';

  // 选中的文件
  const selectedFiles = ref<EnhancedFile[]>([]);

  // 转换状态
  const isConverting = ref(false);
  const convertProgress = ref(0);
  const convertStatus = ref<'normal' | 'active' | 'success' | 'exception'>('normal');
  const progressText = ref('');

  interface ResultItem {
    fileName: string;
    inputPath: string;
    outputPath: string;
    outputFileName: string;
    success: boolean;
    message: string;
  }

  // 转换结果
  const convertResults = ref<Array<ResultItem>>([]);

  // 结果表格列定义
  const resultColumns = [
    {
      title: '文件名',
      dataIndex: 'fileName',
      key: 'fileName',
      ellipsis: true,
      width: 300
    },
    {
      title: '输出路径',
      dataIndex: 'outputPath',
      key: 'outputPath'
    },
    {
      title: '状态',
      key: 'status',
      width: 70
    }
   
  ];

  /**
   * 处理文件选择更新
   * @param value 选中的文件列表
   */
  const handleUpdateModelValue = (value: EnhancedFile[]) => {
    convertResults.value = [];
    if (value.length) {
      // 根据正则过滤掉非法文件, 文件名规则为: export_20250826152020150000_西城区.db

      const regex = /^export_\d{18,20}_[\u4e00-\u9fa5]{2,}\.db$/;
      const validFiles = value.filter(item => regex.test(item.name));
      if (validFiles.length !== value.length) {
        message.warn('请选择正确的数据文件, 文件名规则示例: export_20250826152020150000_西城区.db');
        selectedFiles.value = validFiles;
        return;
      }
    }

    if (value.length > 4) {
      message.warn('最多选择4个文件');
      selectedFiles.value = value.slice(0, 4);
    } else {
      selectedFiles.value = value;
    }
  };

  /**
   * 处理数据文件转Excel
   */
  const handleConvert = async () => {
    if (!selectedFiles.value || selectedFiles.value.length === 0) {
      message.warning('请先选择数据文件');
      return;
    }

    isConverting.value = true;
    convertProgress.value = 0;
    convertStatus.value = 'active';
    convertResults.value = [];

    try {
      const totalFiles = selectedFiles.value.length;

      for (let i = 0; i < totalFiles; i++) {
        const file = selectedFiles.value[i];
        progressText.value = `正在转换: ${file.name} (${i + 1}/${totalFiles})`;

        try {
          // 这里调用后端转换函数
          const result = await DBTranformExcel(file.fullPath);
          if (result.ok) {
            convertResults.value.push({
              fileName: file.name,
              inputPath: file.fullPath,
              outputPath: result.data.outputPath,
              outputFileName: result.data.fileName,
              success: true,
              message: result.message
            });
          } else {
            convertResults.value.push({
              fileName: file.name,
              inputPath: file.fullPath,
              outputPath: '',
              outputFileName: '',
              success: false,
              message: result.message || '转换失败'
            });
          }
        } catch (error) {
          console.error('转换文件失败:', error);
          convertResults.value.push({
            fileName: file.name,
            inputPath: file.fullPath,
            outputPath: '',
            outputFileName: '',
            success: false,
            message: '转换过程中发生错误'
          });
        }

        // 更新进度
        convertProgress.value = Math.round(((i + 1) / totalFiles) * 100);
      }

      convertStatus.value = 'success';
      progressText.value = '转换完成';

      const successCount = convertResults.value.filter(r => r.success).length;
      const failCount = convertResults.value.length - successCount;

      if (failCount === 0) {
        message.success(`所有文件转换成功！共 ${successCount} 个文件`);
      } else {
        message.warning(`转换完成！成功 ${successCount} 个，失败 ${failCount} 个`);
      }
    } catch (error) {
      console.error('转换过程发生错误:', error);
      convertStatus.value = 'exception';
      progressText.value = '转换失败';
      message.error('转换过程发生错误');
    } finally {
      isConverting.value = false;
    }
  };

  /**
   * 文件另存为
   * @param record 结果项
   */
  const saveAsFile = async (record: ResultItem) => {
    const ret = await OpenSaveDialog(
      new main.FileDialogOptions({
        title: '选择导出文件路径',
        defaultFilename: record.outputFileName
      })
    );
    if (ret.canceled) {
      return;
    }
    const filePath = ret.filePaths[0];
    await Movefile(record.outputFileName, filePath);
    message.success('保存成功');
  };
</script>

<style scoped>
  .progress-section {
    margin-top: 30px;
    padding: 20px;
    background-color: #f9f9f9;
    border-radius: 6px;
  }

  .progress-text {
    margin-top: 10px;
    color: #666;
    font-size: 14px;
  }

  .result-section {
    margin-top: 30px;
    text-align: left;
  }

  .result-section h3 {
    margin-bottom: 16px;
    color: #1a5284;
  }

  .operation-area {
    margin-top: 30px;
  }
</style>
