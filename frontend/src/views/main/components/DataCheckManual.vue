<template>
  <div class="wh-100 flex-vertical">
    <a-flex justify="flex-end" style="margin-bottom: 10px">
      <a-button type="primary">批量确认</a-button>
    </a-flex>
    <div class="flex-main relative" ref="tableBoxRef">
      <div class="abs">
        <a-table
          :dataSource="dataSource"
          :row-selection="rowSelection"
          :columns="columns"
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
        <ConfirmTable1 />
      </div>
    </div>
  </a-modal>
</template>

<script setup lang="tsx">
  import { Button, TableColumnType, TableProps } from 'ant-design-vue';
  import { useTableHeight } from '@/hook';
  import { newColumns } from '@/util';
  import ConfirmTable1 from './ConfirmTable1.vue';

  const tableBoxRef = ref(null);
  const tableScroll = useTableHeight(tableBoxRef);

  const dataSource = [
    {
      key: '1',
      unit_name: '企业名称',
      credit_code: '统一社会信用代码',
      import_time: '导入时间',
      check_time: '校核时间'
    }
  ];

  const modal = reactive({
    show: false,
    data: {},
    title: '基本信息',
    showModal: (data: any) => {
      modal.show = true;
      modal.data = data;
      console.log(modal.data);
    },
    handleOk: () => {
      modal.show = false;
    }
  });

  const selectedRowKeys = ref<string[]>([]);
  const rowSelection: TableProps['rowSelection'] = {
    type: 'checkbox',
    selectedRowKeys: unref(selectedRowKeys),
    getCheckboxProps: record => {
      return {
        disabled: record.name === 'Disabled User',
        name: record.name,
        options: {
          label: 'Enterprise Name',
          value: 'Enterprise Name'
        },
        slots: {
          default: () => 'Enterprise Name'
        }
      };
    },
    onChange: (keys, selectedRows) => {
      console.log(`selectedRowKeys: ${selectedRowKeys}`, 'selectedRows: ', keys);
      selectedRowKeys.value = keys as string[];
    }
  };

  function handleCheckClick(row: Record<string, any>) {
    console.log(row);
  }

  const columns: TableColumnType[] = newColumns(
    {
      unit_name: '企业名称',
      credit_code: '统一社会信用代码',
      import_time: '导入时间',
      check_time: '校核时间'
    },
    {
      title: '操作',
      customRender: opt => {
        return (
          <>
            {opt.index === 0 ? (
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
