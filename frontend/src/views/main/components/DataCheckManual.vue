<template>
  <div class="wh-100 flex-vertical">
    <a-flex justify="flex-end" style="margin-bottom: 10px">
      <a-button type="primary" @click="handleBatchConfirm">批量确认</a-button>
    </a-flex>
    <div class="flex-main relative" ref="tableBoxRef">
      <div class="abs">
        <a-table
          :dataSource="dataSource"
          :row-selection="rowSelection"
          :columns="columns"
          :rowKey="(record: {obj_id: string}) => record.obj_id"
          size="small"
          bordered
          :pagination="false"
          :scroll="tableScroll"
        />
      </div>
    </div>
  </div>

  <a-modal
    v-model:open="modal.show"
    :bodyStyle="{ paddingTop: 0 }"
    class="full-screen-modal button-middle"
    :title="modal.title"
    :cancel-button-props="{ style: 'display: none' }"
    @ok="modal.handleOk"
    ok-text="确认当前表格"
  >
    <div class="wh-100 relative">
      <div class="abs">
        <ConfirmTable1  v-if="props.tableType == 'table1'"  :tableInfoList="tableInfoList"  />
        <ConfirmTable2  v-if="props.tableType == 'table2'"   :tableInfoList="tableInfoList"  />
        <ConfirmTable3  v-if="props.tableType == 'table3'"  :tableInfoList="tableInfoList"  />
        <ConfirmTable4  v-if="props.tableType == 'attachment2'"  :tableInfoList="tableInfoList" />
      </div>
    </div>
  </a-modal>
</template>

<script setup lang="tsx">
  import { Button, TableColumnType, TableProps, message } from 'ant-design-vue';
  import { useTableHeight } from '@/hook';
  import { newColumns } from '@/util';
  import ConfirmTable1 from './ConfirmTable1.vue';
  import ConfirmTable2 from './ConfirmTable2.vue';
  import ConfirmTable3 from './ConfirmTable3.vue';
  import ConfirmTable4 from './ConfirmTable4.vue';

  import {
  QueryDataAttachment2, QueryDataDetailTable1,
  QueryDataTable1,
  QueryDataTable2,
  QueryDataTable3,

  ConfirmDataAttachment2,
  ConfirmDataTable1,
  ConfirmDataTable2,
  ConfirmDataTable3,
} from '@wailsjs/go';

  const tableBoxRef = ref(null);
  const tableScroll = useTableHeight(tableBoxRef);
  const tableInfoList = ref<Array<Record<string, any>>>([])
  const objId = ref<string[]>([])
  const props = defineProps({
    tableType: {
      type: String,
      default: ''
    }
  });

  const dataSource = ref([]);

  const queryDataByTableType = (tableType: string) => {
      switch (tableType) {
      case 'table1':
        queryTable1Data();
        break;
      case 'table2':
        queryTable2Data();
        break;
      case 'table3':
        queryTable3Data();
        break;
      case 'attachment2':
        queryAttachment2Data();
        break;
      default:
        console.warn('未知的表格类型:', tableType);
        dataSource.value = [];
    }
  };

  const queryTable1Data = async () => {
    dataSource.value = [];
    const resDetail = await QueryDataTable1();
    if (resDetail.data) {
      dataSource.value = resDetail.data;

      console.debug(dataSource.value,'dataSource.value')
    }
  };

  const queryTable2Data = async () => {
    dataSource.value = [];
    const resDetail = await QueryDataTable2();
    if (resDetail.data) {
      dataSource.value = resDetail.data;
    }
  };

  const queryTable3Data = async () => {
    dataSource.value = [];
    const resDetail = await QueryDataTable3();
    if (resDetail.data) {
      dataSource.value = resDetail.data;
    }
  };

  const queryAttachment2Data = async () => {
    dataSource.value = [];
    const resDetail = await QueryDataAttachment2();
    if (resDetail.data && resDetail.data.length > 0) {
      // 按年份和省市县分组, 再按年份倒序
      const groupedData = resDetail.data.reduce((acc: Record<string, any>, item: Record<string, any>) => {
        const key = `${item.stat_date}-${item.province_name}-${item.city_name}-${item.country_name}`;
        if (!acc[key]) {
          acc[key] = {
            obj_id: key,
            create_time: item.create_time,
            is_confirm: item.is_confirm,
            stat_date: item.stat_date,
            province_name: item.province_name,
            city_name: item.city_name,
            country_name: item.country_name,
            data: []
          };
        } 
        acc[key].data.push(item);
        return acc;
      }, {});

      const list = Object.values(groupedData).sort((a: any, b: any) => {
        return b.stat_date.localeCompare(a.stat_date);
      }) as any;;

      dataSource.value = list;
    }
  };

  // 监听tableType变化
  watch(
    () => props.tableType,
    (newTableType) => {
      if (newTableType) {
        queryDataByTableType(newTableType);
      }
    },
    { immediate: true }
  );

  const modal = reactive({
    show: false,
    data: {} as Record<string, any>,
    title: '基本信息',
    showModal: async (data: any) => {
      modal.show = true;
      modal.data = data;
      objId.value = [modal.data.obj_id]
      switch (props.tableType) {
        case 'table1':
          try {
            const resDetail = await QueryDataDetailTable1(modal.data.obj_id);
            const { main, usage, equip } = resDetail.data;
            tableInfoList.value = [[main], usage, equip] as Array<Record<string, any>>;
          } catch (error) {
            console.error('获取表1详细数据失败:', error);
            tableInfoList.value = [modal.data as Record<string, any>];
          }
          break;
        case 'table2':
          tableInfoList.value = [modal.data as Record<string, any>];
          break;
        case 'table3':
          tableInfoList.value = [modal.data as Record<string, any>];
          break;
        case 'attachment2':
          objId.value = modal.data.data.map((item: any) => item.obj_id);
          tableInfoList.value = modal.data.data as Record<string, any>[];
          break;
      }
    },
    handleOk: async () => {
      try {
     console.log(objId.value,'objId.value');
     
        await executeConfirm(props.tableType, objId.value);
        queryDataByTableType(props.tableType);
      } catch (error) {
        console.error('确认数据失败:', error);
        message.error('确认数据失败，请重试');
      }
      modal.show = false;
    }
  });

  const selectedRowKeys = ref<string[]>([]);
  const selectedRows = ref<Record<string, any>[]>([]);
  
  const rowSelection = computed<TableProps['rowSelection']>(() => ({
    type: 'checkbox',
    selectedRowKeys: selectedRowKeys.value,
    getCheckboxProps: record => ({
      disabled: record.is_confirm == 1 || record.is_confirm == 2,
      name: record.name
    }),
    onChange: (keys, rows) => {
      selectedRowKeys.value = keys as string[];
      selectedRows.value = rows as Record<string, any>[];
      console.debug(selectedRows,'selectedRows')
    }
  }));



  const executeConfirm = async (tableType: string, objIds: string[]) => {
    let result;
    let dataTypeName = '';
    
    objIds.forEach((item) => { console.debug(item,'item') })
    switch (tableType) {
      case 'table1':
        console.debug(objIds.map((item) => item),'objIds')
        result = await ConfirmDataTable1(objIds);
        console.debug(result, 'result')
        dataTypeName = '规上企业数据';
        break;
      case 'table2':
        result = await ConfirmDataTable2(objIds);
        dataTypeName = '其他单位数据';
        break;
      case 'table3':
        result = await ConfirmDataTable3(objIds);
        dataTypeName = '新上项目数据';
        break;
      case 'attachment2':
        result = await ConfirmDataAttachment2(objIds);
        dataTypeName = '区域综合数据';
        break;
      default:
        throw new Error(`未知的表格类型: ${tableType}`);
    }
    
    if (result.message) {
      message.success(result.message);
    } else {
      const count = objIds.length;
      const action = count > 1 ? '批量确认' : '确认';
      message.success(`已${action} ${count} 条${dataTypeName}`);
    }
    return result;
  };

  const handleBatchConfirm = async () => {
    if (selectedRowKeys.value.length === 0) {
      message.warning('请先选择要确认的数据');
      return;
    }

    if (props.tableType === 'attachment2') {
      selectedRowKeys.value = selectedRows.value.map((item) => item.data.map((item: any) => item.obj_id)).flat();
    }

    try {
      await executeConfirm(props.tableType, selectedRowKeys.value);
      selectedRowKeys.value = [];
      queryDataByTableType(props.tableType);
    } catch (error) {
      console.error('批量确认失败:', error);
      message.error('批量确认失败，请重试');
    }
  };



  const formatDateTime = (timeStr: string) => {
    if (!timeStr) return '';
    
    if (timeStr.includes('T')) {
      try {
        const date = new Date(timeStr);
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const hour = String(date.getHours()).padStart(2, '0');
        const minute = String(date.getMinutes()).padStart(2, '0');
        const second = String(date.getSeconds()).padStart(2, '0');
        return `${year}年${month}月${day}日 ${hour}:${minute}:${second}`;
      } catch (error) {
        console.warn('时间格式化失败:', timeStr, error);
        return timeStr;
      }
    }
    
    return timeStr;
  };

  const columns: TableColumnType[] = newColumns(
    props.tableType === 'table1' ? {
      unit_name:  '企业名称',
      credit_code: '统一社会信用代码',
      stat_date : '年份'
    } : 
    props.tableType === 'table2' ? {
      unit_name: '单位名称',
      credit_code: '统一社会信用代码',
      stat_date : '年份'
    } : 
    props.tableType === 'table3' ? {
      project_name : '项目名称',
      project_code : '项目代码',
      actual_time : '年份'
    } : 
    props.tableType === 'attachment2' ? {
      city_name: '区域名称',
      stat_date : '年份'
    } : {
      unit_name: '企业名称',
      credit_code: '统一社会信用代码'
    },

    {
      title: '校核时间',
      dataIndex: 'create_time',
      key: 'create_time',
      align: 'center',
      ellipsis: true,
      customRender: ({ text }) => {
        return formatDateTime(text);
      }
    },
    {
      title: '操作',
      customRender: opt => {
        return (
          <>
            {opt.value.is_confirm == 0 ? (
              <Button type="primary" size="small" onClick={() => modal.showModal(opt.record)}>
                校核
              </Button>
            ) : (
              <Button type="primary" size="small" class="ant-btn-loading">
                已校核
              </Button>
            )}
          </>
        );
      }
    }
  );
</script>


<style>
/* 全局样式确保模态框按钮颜色生效 */
.ant-modal-footer .ant-btn-primary {
  background-color: #1A5284 !important;
  border-color: #1A5284 !important;
}

.ant-modal-footer .ant-btn-primary:hover {
  background-color: #0f3a5f !important;
  border-color: #0f3a5f !important;
}

.ant-modal-footer .ant-btn-primary:focus {
  background-color: #1A5284 !important;
  border-color: #1A5284 !important;
}

.overflow-auto {
  overflow: auto;
}
</style>