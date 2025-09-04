<template>
  <div class="page-header">
    <span class="header-title">数据导出</span>
  </div>

  <div class="page-content flex-vertical">
    <div class="flex-main relative" ref="tableBoxRef">
      <div class="abs">
        <a-table :dataSource="dataSource" :columns="columns" size="small" bordered :pagination="false" :scroll="tableScroll" />
      </div>
    </div>

    <div class="operation-area">
      <div>
        <div class="result-text">需要自动校验、人工校验都通过才能导出</div>
        <a-button type="primary" style="margin: 10px auto 0" @click="handleExportClick">导出汇总数据（.db）</a-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="tsx">
  import { message, TableColumnType } from 'ant-design-vue';
  import { useTableHeight } from '@/hook';
  import { ExportDBData, OpenSaveDialog, QueryExportData, GetEnhancedAreaConfig, GetAreaConfig } from '@wailsjs/go';
  import { main } from '@wailsjs/models';
  import dayjs from 'dayjs';
  import { TableType, TableTypeName } from '@/views/constant';
  const tableBoxRef = ref(null);
  const tableScroll = useTableHeight(tableBoxRef);

  const dataSource = ref<ExportItem[]>([]);

  function normalizeData(item: ExportItem[], tableTypeName: string) {
    if (!item?.length) {
      return [{ tableTypeName, stat_date: '--', count: 0, is_checked_no: 0, is_checked_yes: 0, is_confirm_no: 0, is_confirm_yes: 0 }];
    }
    return item
      .map(item => {
        item.tableTypeName = tableTypeName;
        item.stat_date = item.stat_date || '--';
        return item;
      })
      .sort((a, b) => a.stat_date.localeCompare(b.stat_date));
  }

  onMounted(async () => {
    const result = await QueryExportData();
    console.log(result);
    if (result.ok) {
      let list: ExportItem[] = [];
      list = list.concat(normalizeData(result.data[TableType.table1], TableTypeName.table1));
      list = list.concat(normalizeData(result.data[TableType.table2], TableTypeName.table2));
      list = list.concat(normalizeData(result.data[TableType.table3], TableTypeName.table3));
      list = list.concat(normalizeData(result.data[TableType.attachment2], TableTypeName.attachment2));
      dataSource.value = list;
    }
  });

  const columns: TableColumnType<ExportItem>[] = reactive([
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
      align: 'center'
    },
    {
      title: '导入进度',
      align: 'center',
      customRender: ({ record }) => {
        return `${record.is_checked_yes}`;
        // /${record.count}
      }
    },
    {
      title: '自动校验',
      ellipsis: true,
      align: 'center',
      customRender: ({ record }) => {
        return `${record.is_checked_yes}`;
        // /${record.count}
      }
    },
    {
      title: '人工校验',
      align: 'center',
      customRender: ({ record }) => {
        return `${record.is_confirm_yes}`;
        // ${record.count}
      }
    }
  ]);

  onMounted(async () => {
    const areaResult = await GetAreaConfig();
    if (areaResult.ok && areaResult.data?.country_name) {
      columns.splice(2, 1);
    }
  });

  const handleExportClick = async () => {
    let allPass = true;
    for (const item of dataSource.value) {
      if (item.is_confirm_no > 0 || item.is_checked_no !== item.is_confirm_no) {
        allPass = false;
        break;
      }
    }

    if (!allPass) {
      message.error('存在自动校验、人工校验未通过的数据，不能导出');
      return;
    }

    const areaResult = await GetEnhancedAreaConfig();
    const areaConfig = areaResult.data;
    const areaCode = areaConfig.country_code || areaConfig.city_code;
    const areaName = areaConfig.country_name || areaConfig.city_name;

    const result = await OpenSaveDialog(
      new main.FileDialogOptions({
        title: '导出汇总数据',
        defaultFilename: `export_${dayjs().format('YYYYMMDDHHmmss')}${areaCode}_${areaName}.db`
      })
    );

    console.log(result);
    if (result.canceled) {
      console.log('用户取消保存');
      return;
    }

    const exportResult = await ExportDBData(result.filePaths[0]);
    console.log(exportResult);
    if (exportResult.ok) {
      message.success('导出成功');
    } else {
      message.error('导出失败:' + exportResult.message);
    }
  };

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
