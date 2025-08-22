<template>
  <div class="page-header">&nbsp;</div>

  <div class="page-content flex-vertical">
    <div class="flex-main relative" ref="tableBoxRef">
      <div class="abs">
        <a-table :dataSource="dataSource" :columns="columns" size="small" bordered :pagination="false" :scroll="tableScroll" />
      </div>
    </div>

    <div style="text-align: center; margin-top: 20px">
      <a-button type="primary">导出汇总数据（.db）</a-button>
    </div>
  </div>
</template>

<script setup lang="tsx">
  import { TableColumnType, TableProps } from 'ant-design-vue';
  import { useTableHeight } from '@/hook';
  const tableBoxRef = ref(null);
  const tableScroll = useTableHeight(tableBoxRef);

  const dataSource = ref<Record<string, any>>([]);

  setTimeout(() => {
    dataSource.value = Array.from({ length: 100 })
      .fill(1)
      .map((_, i) => {
        return {
          key: '1' + i,
          tableType: '表1',
          year: '2024',
          importProgress: 32,
          completeRate: 32,
          manualCheck: '已完成',
          modelCheck: '已完成'
        };
      });
  });

  const columns: TableColumnType[] = [
    {
      title: '年份',
      dataIndex: 'year',
      key: 'year',
      align: 'center'
      // customCell: (_, index) => {
      //   if (index === 2) {
      //     return { rowSpan: 2 };
      //   }
      // }
    },
    {
      title: '',
      dataIndex: 'tableType',
      key: 'tableType',
      align: 'center'
    },
    {
      title: '导入进度',
      dataIndex: 'importProgress',
      key: 'importProgress',
      align: 'center'
    },
    {
      title: '完整率',
      dataIndex: 'completeRate',
      key: 'completeRate',
      ellipsis: true,
      align: 'center'
    },
    {
      title: '人工校验',
      dataIndex: 'manualCheck',
      key: 'manualCheck',
      ellipsis: true,
      align: 'center'
    },
    {
      title: '模型校验',
      dataIndex: 'modelCheck',
      key: 'modelCheck',
      ellipsis: true,
      align: 'center'
    }
  ];
</script>

<style scoped></style>
