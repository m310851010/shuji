<template>
  <div
    class="box-grey drop-area"
    :class="{ 'drop-active': isDragging, 'drop-invalid': !hasValidFile }"
    @drop="handleDrop"
    @dragover="handleDragEnter"
    @dragleave="handleDragLeave"
    style="--wails-drop-target: drop"
  >
    <CloudUploadOutlined style="font-size: 40px; color: #9ca3af" />
    <p class="drop-text">
      <span v-if="env.os === 'windows'">将文件拖到此处，或</span>
      <a @click="onOpenFileDialog">点击选择文件</a>
      或
      <a @click="onOpenFolderDialog">点击选择文件夹</a>
    </p>
    <a-alert type="warning" style="color: #944310">
      <template #message>
        <slot></slot>
        <div v-if="!$slots.default">
          <div>只能选择Excel文件（.xlsx/.xls），支持批量选择</div>
          <div v-if="env.os === 'windows'">支持一次性拖多个Excel文件，以及整个文件夹</div>
          <div>选择文件后，点击上方按钮开始导入</div>
        </div>
      </template>
    </a-alert>
  </div>

  <div v-if="selectedFiles && selectedFiles.length" class="selected-files">
    <p class="files-title">已选择 {{ selectedFiles.length }} 个文件：</p>
    <a-row type="flex" justify="space-between" class="box-grey small no-bg" v-for="(file, index) in selectedFiles" :key="index">
      <div
        style="line-height: 24px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; cursor: pointer"
        @click="handleOpenFile(file)"
      >
        <PaperClipOutlined />
        <span style="margin-left: 5px">{{ file.name }}</span>
      </div>
      <a-button type="text" danger @click.stop="removeFile(index)" size="small">移除</a-button>
    </a-row>
  </div>
</template>

<script setup lang="ts">
  import { CloudUploadOutlined, PaperClipOutlined } from '@ant-design/icons-vue';
  import {OpenFileDialog, OpenExternal, GetFileInfo, Readdir} from '@wailsjs/go';
  import { main } from '@wailsjs/models';
  import { EXCEL_TYPES } from '@/views/constant';
  import { getFileExtension } from '@/util/utils';
  import { setFileDropHandler } from '@/hook/useFileDrop';
  import { useEnv } from "@/hook/useEnv";

  const props = defineProps({
    // 验证文件是否有效,默认是excel文件
    title: {
      type: String,
      required: false,
      default: '选择文件'
    },
    filterName: {
      type: String,
      required: false,
      default: 'Excel文件'
    },
    filterPattern: {
      type: String,
      required: false,
      default: '*.xlsx;*.xls'
    },
    validFile: {
      type: Object,
      required: false,
      default: () => EXCEL_TYPES
    },

    // 接受的文件类型,是否允许拖拽,默认是excel文件
    accept: {
      type: [Array, Function],
      required: false,
      default: () => EXCEL_TYPES
    }
  });

  const selectedFiles = defineModel<EnhancedFile[]>({ type: Array, default: () => [] });

  // 定义 emits
  const emit = defineEmits<{
    (e: 'file-change', files: EnhancedFile[]): void;
    (e: 'upload-success'): void;
  }>();

  const isDragging = ref(false);
  const hasValidFile = ref(true);
  const env = useEnv();

  const allFiles: { file: File; valid: boolean }[] = [];

  onMounted(async () => {
    setFileDropHandler((dropFiles, x, y) => {
      console.log(allFiles);
      const files: EnhancedFile[] = [];
      const validFile = props.validFile;
      let valid = false;
      const isFunction = typeof validFile === 'function';

      for (let i = 0; i < allFiles.length; i++) {
        const item = allFiles[i];
        if (!item.valid) {
          continue;
        }

        const fileInfo = dropFiles[i];
        if (!fileInfo.isDirectory) {
          files.push(fileInfo);
          continue;
        }

        for (const fileInDir of fileInfo.files!) {
          if (isFunction) {
            valid = validFile(fileInDir, null);
          } else {
            valid = validFile.indexOf(fileInDir.ext) >= 0;
          }
          if (valid) {
            files.push(fileInDir);
          }
        }
      }

      if (files.length) {
        selectedFiles.value = files;
        emit('file-change', selectedFiles.value);
      }
    });
  });

  const handleDragEnter = (e: DragEvent) => {
    // e.preventDefault();
    // e.stopPropagation();

    hasValidFile.value = true;
    if (!e.dataTransfer) {
      return;
    }

    const items = e.dataTransfer.items;
    if (!e.dataTransfer.items) {
      e.dataTransfer.dropEffect = 'none';
      return;
    }

    const accept = props.accept;
    let valid = false;
    const isFunction = typeof accept === 'function';
    for (let i = 0; i < items.length; i++) {
      const item = items[i];
      if (item.kind !== 'file') {
        continue;
      }

      if (!item.type) {
        valid = true;
        break;
      }

      if (isFunction) {
        valid = accept(item, 'dragEnter');
      } else if (accept && accept.length && accept.indexOf(item.type) !== -1) {
        // 判断文件类型是否在accept中
        valid = true;
        break;
      }

      if (valid) {
        break;
      }
    }

    if (!valid) {
      hasValidFile.value = false;
      e.dataTransfer.dropEffect = 'none';
      return;
    }
    isDragging.value = true;
  };

  const handleDragLeave = (e: DragEvent) => {
    hasValidFile.value = true;
    isDragging.value = false;
  };

  const handleDrop = (e: DragEvent) => {
    hasValidFile.value = true;
    isDragging.value = false;
    allFiles.length = 0; // 存储合法的文件

    if (!e.dataTransfer) {
      return;
    }
    const files = e.dataTransfer.files;
    if (!files.length) {
      return;
    }

    const validFile = props.validFile;
    let valid = false;
    const isFunction = typeof validFile === 'function';

    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      if (file.type === '' && file.size % 4096 === 0) {
        allFiles.push({ file, valid: true });
        continue;
      }

      if (isFunction) {
        valid = validFile(file, e);
      } else {
        const fileExt = getFileExtension(file.name);
        valid = validFile.indexOf(file.type) >= 0 || validFile.indexOf(fileExt) >= 0;
      }
      allFiles.push({ file, valid });
    }
  };

  const onOpenFileDialog = async () => {
    const result = await OpenFileDialog(
      new main.FileDialogOptions({
        title: props.title,
        multiSelections: true,
        filters: [{ name: props.filterName, pattern: props.filterPattern }]
      })
    );

    if (result.canceled) {
      return;
    }

    const files: EnhancedFile[] = [];
    for (const fullPath of result.filePaths) {
      const fileInfo = await GetFileInfo(fullPath);
      files.push(fileInfo as unknown as EnhancedFile);
    }

    if (files.length) {
      selectedFiles.value = files;
      emit('file-change', selectedFiles.value);
    }
  };

  const onOpenFolderDialog = async () => {
    const result = await OpenFileDialog(
      new main.FileDialogOptions({
        title: props.title,
        multiSelections: false,
        openDirectory: true,
        createDirectory: true
      })
    );

    if (result.canceled) {
      return;
    }

    const files: EnhancedFile[] = await getFilesDir(result.filePaths[0]);
    if (files.length) {
      selectedFiles.value = files;
      emit('file-change', selectedFiles.value);
    }
  };

  async function getFilesDir(dirPath: string) {
    const dirResult = await Readdir(dirPath);
    if (!dirResult.ok) {
      return [];
    }

    const files: EnhancedFile[] = [];
    const validFile = props.validFile;
    let valid = false;
    const isFunction = typeof validFile === 'function';

    for (const filePath of dirResult.data) {
      const fileInfo = await GetFileInfo(filePath);
      if (isFunction) {
        valid = validFile(fileInfo, null);
      } else {
        const fileExt = getFileExtension(fileInfo.name);
        valid = validFile.indexOf(fileExt) >= 0;
      }

      if (valid) {
        files.push(fileInfo as unknown as EnhancedFile);
      }
    }

    return files;
  }

  // 移除文件
  const removeFile = (index: number) => {
    if (selectedFiles.value) {
      selectedFiles.value.splice(index, 1);
      emit('file-change', selectedFiles.value);
    }
  };

  const handleOpenFile = (file: EnhancedFile) => {
    OpenExternal(file.fullPath);
  };
</script>

<style scoped>
  .drop-area {
    border: 2px dashed #e8e8e8;
    border-radius: 4px;
    padding: 20px;
    text-align: center;
    transition: border-color 0.3s;
  }

  .drop-area:hover,
  .drop-active {
    border-color: #1890ff;
  }
  .drop-invalid {
    border-color: #ff4d4f;
    background-color: rgba(255, 77, 79, 0.11);
  }

  .drop-text {
    margin-top: 30px;
    margin-bottom: 16px;
    font-size: 16px;
  }

  .file-info {
    color: #faad14;
    font-size: 14px;
    line-height: 1.5;
    background-color: #fffbe6;
    padding: 12px;
    border-radius: 4px;
    display: inline-block;
  }

  .selected-files {
    margin-top: 20px;
  }

  .files-title {
    margin-bottom: 16px;
    font-weight: 500;
  }
</style>
