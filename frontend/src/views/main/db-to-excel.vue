<template>
  <div class="wh-100 flex-vertical">
    <div class="page-header">
      <span class="header-title">DB文件转Excel</span>
    </div>
    <div class="page-content text-center">
      <UploadComponent
        v-model="selectedFiles"
        v-on:update:model-value="handleUpdateModelValue"
        :accept="() => true"
        :validFile="['db']"
        filterName="DB文件"
        filterPattern="*.db"
        title="选择DB文件"
      >
        <div>只能选择DB文件（.db），支持批量选择</div>
        <div>支持一次性拖一个或多个DB文件，以及整个文件夹</div>
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
        <a-table 
          :dataSource="convertResults" 
          :columns="resultColumns" 
          :pagination="false"
          size="small"
          bordered
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'status'">
              <a-tag :color="record.success ? 'green' : 'red'">
                {{ record.success ? '成功' : '失败' }}
              </a-tag>
            </template>
            <template v-if="column.key === 'action'">
              <a-button 
                v-if="record.success" 
                type="link" 
                @click="openFile(record.outputPath)"
              >
                打开文件
              </a-button>
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
  // 这里需要导入后端转换函数，假设为 ConvertDbToExcel
  // import { ConvertDbToExcel, OpenFileInExplorer } from '@wailsjs/go';

  // 选中的文件
  const selectedFiles = ref<EnhancedFile[]>([]);
  
  // 转换状态
  const isConverting = ref(false);
  const convertProgress = ref(0);
  const convertStatus = ref<'normal' | 'active' | 'success' | 'exception'>('normal');
  const progressText = ref('');
  
  // 转换结果
  const convertResults = ref<Array<{
    fileName: string;
    inputPath: string;
    outputPath: string;
    success: boolean;
    message: string;
  }>>([]);

  // 结果表格列定义
  const resultColumns = [
    {
      title: '文件名',
      dataIndex: 'fileName',
      key: 'fileName',
    },
    {
      title: '输入路径',
      dataIndex: 'inputPath',
      key: 'inputPath',
      ellipsis: true,
    },
    {
      title: '输出路径',
      dataIndex: 'outputPath',
      key: 'outputPath',
      ellipsis: true,
    },
    {
      title: '状态',
      key: 'status',
    },
    {
      title: '消息',
      dataIndex: 'message',
      key: 'message',
    },
    {
      title: '操作',
      key: 'action',
    },
  ];

  /**
   * 处理文件选择更新
   * @param files 选中的文件列表
   */
  const handleUpdateModelValue = (files: EnhancedFile[]) => {
    selectedFiles.value = files;
    // 清空之前的转换结果
    convertResults.value = [];
  };

  /**
   * 处理DB文件转Excel
   */
  const handleConvert = async () => {
    if (!selectedFiles.value || selectedFiles.value.length === 0) {
      message.warning('请先选择DB文件');
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
          // const result = await ConvertDbToExcel(file.fullPath);
          
          // 模拟转换过程
          await new Promise(resolve => setTimeout(resolve, 1000));
          
          // 模拟转换结果
          const result = {
            ok: true,
            data: {
              outputPath: file.fullPath.replace('.db', '.xlsx'),
              message: '转换成功'
            }
          };
          
          if (result.ok) {
            convertResults.value.push({
              fileName: file.name,
              inputPath: file.fullPath,
              outputPath: result.data.outputPath,
              success: true,
              message: result.data.message
            });
          } else {
            convertResults.value.push({
              fileName: file.name,
              inputPath: file.fullPath,
              outputPath: '',
              success: false,
              message: result.data?.message || '转换失败'
            });
          }
        } catch (error) {
          console.error('转换文件失败:', error);
          convertResults.value.push({
            fileName: file.name,
            inputPath: file.fullPath,
            outputPath: '',
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
   * 打开文件所在位置
   * @param filePath 文件路径
   */
  const openFile = async (filePath: string) => {
    try {
      // 这里调用后端函数打开文件
      // await OpenFileInExplorer(filePath);
      message.success('文件已在资源管理器中打开');
    } catch (error) {
      console.error('打开文件失败:', error);
      message.error('打开文件失败');
    }
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