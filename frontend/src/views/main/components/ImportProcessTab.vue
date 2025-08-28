<template>
  <div class="box-grey">
    <a-flex justify="flex-end">
      <a-button type="primary" @click="handleExportClick">导出清单</a-button>
    </a-flex>
  </div>

  <div class="box-grey flex-main flex-vertical">
    <a-alert type="info" style="margin-bottom: 10px" v-if="titleList.length">
      <template #message>
        <div class="process-message" v-for="(item, index) in titleList">
          <span>{{ item.label }}</span>
          <span class="number-text">{{ item.total }}</span>
          <span>{{ item.unit }}</span>
          <span v-if="index < titleList.length - 1">，</span>
        </div>
      </template>
    </a-alert>
    <div class="flex-main relative">
      <div class="abs" ref="tableBoxRef" style="background-color: #fff">
        <a-table :dataSource="dataSource" :columns="columns" size="small" bordered :pagination="false" :scroll="tableScroll" />
      </div>
    </div>
  </div>
</template>

<script setup lang="tsx">
  import { message, TableColumnType, Tag, Spin } from 'ant-design-vue';
  import { useTableHeight } from '@/hook';
  import { TableType } from '@/views/constant';
  import {
    OpenSaveDialog,
    QueryTable1Process,
    QueryTable2Process,
    QueryTable3Process,
    QueryTableAttachment2Process,
    ExportTable1ProgressToExcel,
    ExportTable2ProgressToExcel,
    ExportTable3ProgressToExcel,
    ExportAttachment2ProgressToExcel,
    GetCachePath,
    Removefile,
    Movefile
  } from '@wailsjs/go';
  import { db, main } from '@wailsjs/models';
  import { newColumns } from '@/util';
  import { Table3Columns } from '@/views/main/components/columns';
  import dayjs from 'dayjs';

  const props = defineProps({
    tableType: {
      type: String,
      default: ''
    }
  });

  const columns = ref<any>([]);
  const titleList = ref<TitleItem[]>([]);

  function getUnit(): string {
    return { [TableType.table1]: '家企业', [TableType.table2]: '家企业', [TableType.table3]: '张表', [TableType.attachment2]: '张表' }[
      props.tableType
    ]!;
  }

  async function handleTable(result: Promise<db.QueryResult>, initColumns: (data: ProcessData) => any[]) {
    const res = await result;
    if (!res.ok) {
      message.error(res.message);
      return;
    }
    const data = res.data as ProcessData;
    if (!data) {
      return;
    }
    const _columns = initColumns(data);
    if (data.years) {
      data.years.sort();
      data.years.forEach(year => {
        _columns.push({
          title: `${year}年数据`,
          dataIndex: year,
          key: year,
          align: 'center',
          customRender: (opt: any) => {
            return <>{opt.record[year] ? <Tag color="success">已导入</Tag> : <Tag>未导入</Tag>}</>;
          }
        });
      });

      //计算标题
      const yearMap = data.years.reduce(
        (acc, year) => {
          acc[year] = 0;
          return acc;
        },
        {} as Record<string, number>
      );
      for (const year of data.years) {
        for (const item of data.list) {
          if (item[year]) {
            yearMap[year]++;
          }
        }
      }

      titleList.value = data.years.map(year => ({
        label: `${year}年已录入`,
        total: yearMap[year],
        unit: getUnit()
      }));
    }
    columns.value = _columns;
    dataSource.value = data.list || [];
  }

  function getAreaName(area_level: number) {
    return { '1': '城市', '2': '区县', '3': '区县' }[area_level]!;
  }

  onMounted(async () => {
    if (props.tableType === TableType.table1) {
      await handleTable(QueryTable1Process(), _ => newColumns({ unit_name: '企业' }));
    } else if (props.tableType === TableType.table2) {
      await handleTable(QueryTable2Process(), data => newColumns({ unit_name: '企业' }));
    } else if (props.tableType === TableType.table3) {
      const res = await QueryTable3Process();
      if (!res.ok) {
        message.error(res.message);
        return;
      }
      const data = res.data as ProcessData;
      if (!data) {
        return;
      }

      dataSource.value = data.list || [];
      if (data.area_level === 3) {
        titleList.value = [{ label: '共', total: data.list.length, unit: `条数据` }];
        columns.value = Table3Columns;
        return;
      }

      titleList.value = [{ label: '已录入', total: data.list.filter(item => item.is_import).length, unit: `张表` }];

      columns.value = newColumns(
        { area_name: getAreaName(data.area_level!) },
        {
          title: '是否导入',
          dataIndex: 'is_import',
          key: 'is_import',
          align: 'center',
          customRender: opt => {
            return <>{opt.record['is_import'] ? <Tag color="success">已导入</Tag> : <Tag>未导入</Tag>}</>;
          }
        }
      );
    } else if (props.tableType === TableType.attachment2) {
      await handleTable(QueryTableAttachment2Process(), (data: ProcessData) => {
        if (data.area_level === 3) {
          titleList.value = [{ label: '共', total: data.list.length, unit: `条数据` }];
          return Table3Columns;
        }
        return newColumns({ area_name: getAreaName(data.area_level!) });
      });
    }
  });

  const tableBoxRef = ref(null);
  const tableScroll = useTableHeight(tableBoxRef);

  const dataSource = ref<any>([]);

  // 处理导出按钮点击
  const handleExportClick = async () => {
    if (!dataSource.value.length) {
      message.error('导入进度没有数据，不能导出');
      return;
    }

    const now = Date.now() + '' + Math.random();
    const retPath = (await GetCachePath('')) + '/导出进度' + now + '.xlsx';

    const fn = {
      [TableType.table1]: () => ExportTable1ProgressToExcel(retPath),
      [TableType.table2]: () => ExportTable2ProgressToExcel(retPath),
      [TableType.table3]: () => ExportTable3ProgressToExcel(retPath),
      [TableType.attachment2]: () => ExportAttachment2ProgressToExcel(retPath)
    }[props.tableType];
    const res = await fn();
    if (res.ok) {
      const ret = await OpenSaveDialog(
        new main.FileDialogOptions({
          title: '选择导出文件路径',
          defaultFilename: `导出进度${dayjs().format('YYYY-MM-DD HH-mm')}.xlsx`
        })
      );
      if (ret.canceled) {
        await Removefile(retPath);
        return;
      }
      const filePath = ret.filePaths[0];
      await Movefile(retPath, filePath);
      message.success('导出成功');
    } else {
      message.error(res.message);
    }
  };

  interface ProcessData {
    list: Record<string, any>[];
    years?: string[];
    area_level?: number;
  }

  interface TitleItem {
    label: string;
    total: number;
    unit: string;
  }
</script>

<style scoped>
  .process-message {
    color: #096aa2;
    font-size: 15px;
    display: inline-block;

    .number-text {
      color: #292ea7;
      font-weight: bold;
      margin: 0 5px;
    }
  }

  .pagination-container {
    position: absolute;
    bottom: 0;
    right: 0;
    padding: 10px;
  }
</style>
