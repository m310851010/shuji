<template>
  <div class="page-header">&nbsp;</div>

  <div class="page-content flex-vertical">
    <div class="flex-main relative" ref="tableBoxRef">
      <div class="abs">
        <a-table :dataSource="dataSource" :columns="columns" size="small" bordered :pagination="false" :scroll="tableScroll" />
      </div>
    </div>

    <div class="operation-area">
      <a-button type="primary" @click="handleExportClick">导出汇总数据（.db）</a-button>
    </div>
  </div>
</template>

<script setup lang="tsx">
import {message, TableColumnType,} from 'ant-design-vue';
  import { useTableHeight } from '@/hook';
import {ExportDBData, OpenSaveDialog, QueryExportData} from '@wailsjs/go';
  import {main} from '@wailsjs/models';
import dayjs from 'dayjs';
import {TableType, TableTypeName} from '@/views/constant';
  const tableBoxRef = ref(null);
  const tableScroll = useTableHeight(tableBoxRef);

  const dataSource = ref<ExportItem[]>([]);

  function normalizeData(item: ExportItem[], tableTypeName: string) {
    if (!item?.length) {
      return [{tableTypeName, stat_date:'--', count: 0, is_checked_no: 0, is_checked_yes: 0, is_confirm_no:0, is_confirm_yes: 0}];
    }
    return item.map(item => {
      item.tableTypeName = tableTypeName;
      item.stat_date = item.stat_date || '--';
      return item;
    })
  }

function normalizeData2(item: ExportItem[], tableTypeName: string) {
  if (!item?.length) {
    return [{tableTypeName, stat_date:'--', count: 1, is_checked_no: 0, is_checked_yes: 0, is_confirm_no:0, is_confirm_yes: 0}];
  }
  return item.map(item => {
    item.tableTypeName = tableTypeName;
    item.count = 1;
    item.stat_date = item.stat_date || '--';
    item.is_confirm_yes = item.is_confirm_yes > 0 ? 1 : 0;
    item.is_confirm_no = item.is_confirm_no  ===0? 1 : 0;
    item.is_checked_yes = item.is_checked_yes > 0 ? 1 : 0;
    item.is_checked_no = item.is_checked_yes  === 0 ? 1 : 0;
    return item;
  })
}

  onMounted(async () => {
    const result = await QueryExportData();
    console.log(result)
    if (result.ok) {
      let list: ExportItem[] = [];
          list = list.concat(normalizeData(result.data[TableType.table1], TableTypeName.table1));
          list = list.concat(normalizeData(result.data[TableType.table2], TableTypeName.table2));
          list = list.concat(normalizeData(result.data[TableType.table3], TableTypeName.table3));
          list = list.concat(normalizeData2(result.data[TableType.attachment2], TableTypeName.attachment2));
      dataSource.value = list
    }
  })

  const columns: TableColumnType<ExportItem>[] = [
    {
      title: '年份',
      dataIndex: 'stat_date',
      key: 'stat_date',
      align: 'center'

    },
    {
      title: '表格类型',
      dataIndex: 'tableTypeName',
      key: 'tableTypeName',
      align: 'center',
    },
    {
      title: '导入进度',
      align: 'center',
      customRender: ({record}) => {
        return `${record.is_checked_yes}/${record.count}`
      }
    },
    {
      title: '自动校验',
      ellipsis: true,
      align: 'center',
      customRender: ({ record}) => {
        return `${record.is_checked_yes}/${record.count}`
      }
    },
    {
      title: '人工校验',
      align: 'center',
      customRender: ({ record}) => {
        return `${record.is_confirm_yes}/${record.count}`
      }
    }
  ];
  const handleExportClick = async () => {
    const result = await OpenSaveDialog(new main.FileDialogOptions({
      title: '导出汇总数据',
      defaultFilename: `export_导出汇总数据_${dayjs().format('YYYYMMDD')}.db`,
    }));

    console.log(result);
    if (result.canceled) {
      console.log('用户取消保存');
      return
    }

    const exportResult = await ExportDBData(result.filePaths[0])
    console.log(exportResult);
    if (exportResult.ok) {
      message.success('导出成功');
    } else {
      message.error('导出失败:' + exportResult.message);
    }
  }

  interface ExportItem {
    tableTypeName: string;
    stat_date: string;
    is_confirm_no: number;
    is_confirm_yes: number;
    is_checked_no: number;
    is_checked_yes: number;
    count: number;
  }
</script>

<style scoped></style>
