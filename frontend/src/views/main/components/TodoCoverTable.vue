


<script setup lang="ts">
import {ref, defineProps, defineEmits, onMounted} from 'vue';

const props = defineProps<{
  fileList: Array<{ fileName: string; filePath: string }>
}>()

onMounted(() => {
  console.log(props.fileList);
})
const emit = defineEmits(['updateFileList'])

// 选中的行
const selectedRows = ref<Array<{ fileName: string; fullPath: string }>>([])

// 返回选中的filePath数组
function getSelectedFilePaths() {
  return selectedRows.value.map(item => item.fullPath)
}

defineExpose({
  getSelectedFilePaths
})


</script>
<template>
<a-table
  :dataSource="props.fileList"
  :rowSelection="{
    selectedRowKeys: selectedRows.map((item: any)  => item.fullPath),
    onChange: (selectedRowKeys: string[], selectedRowsData: any[]) => {
      selectedRows = selectedRowsData;
      emit('updateFileList', selectedRowsData.map(item => item.fullPath));
    }
  }"
  rowKey="fullPath"
  style="width: 100%"
  size="small"
  bordered
  :pagination="false"
>
  <a-table-column title="文件名" dataIndex="fileName" key="fileName" />
</a-table>

</template>

<style scoped>

</style>
