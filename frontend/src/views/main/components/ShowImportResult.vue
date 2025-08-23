<template>
  <a-table :dataSource="resultList" :columns="columns" rowKey="fileName" bordered :pagination="false">
    <template #bodyCell="{ column, record }">
      <template v-if="column.dataIndex === 'ok'">
        <a-tag :color="record.ok ? 'green' : 'red'">
          {{ record.ok ? '成功' : '失败' }}
        </a-tag>
      </template>
      <template v-else-if="column.dataIndex === 'data'">
        <div v-for="(msg, idx) in record.data" :key="idx" style="color: #ff4d4f;">
            {{ msg }}
        </div>
      </template>
      <template v-else>
        {{ record[column.dataIndex] }}
      </template>
    </template>
  </a-table>
</template>

<script setup lang="ts">
import { defineProps, computed } from 'vue'

const props = defineProps<{
  resultList: Array<{
    fileName: string
    ok: boolean
    data: string[]
  }>
}>()

const columns = [
  {
    title: '文件名',
    dataIndex: 'fileName',
    key: 'fileName',
    width: 280,
  },
  {
    title: '导入状态',
    dataIndex: 'ok',
    key: 'ok',
    width: 100,
  },
  {
    title: '备注',
    dataIndex: 'data',
    key: 'data',
  }
]
</script>

