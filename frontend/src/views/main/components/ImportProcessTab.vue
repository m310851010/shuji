<template>
  <div class="box-grey">
    <a-flex justify="flex-end">
      <a-button type="primary" @click="handleExportClick">导出清单</a-button>
    </a-flex>
  </div>

  <div class="box-grey flex-main flex-vertical">
    <a-alert type="info" style="margin-bottom: 10px">
      <template #message>
        <div class="process-message">
          <span>共导入</span>
          <span class="number-text">{{ totalEnterprise }}</span>
          <span>家企业</span>

          <!-- 项目和区域时展示 -->
          <!-- <span class="number-text">{{ totalTable }}</span>
          <span>个表格</span> -->
        </div>
      </template>
    </a-alert>
    <div class="flex-main relative">
      <div class="abs" ref="tableBoxRef" style="background-color: #fff">
        <a-table :dataSource="dataSource" :columns="columns" size="small" bordered :pagination="false" :scroll="tableScroll" />
      
         <!-- 分页组件 -->
        <div class="pagination-container">
          <a-pagination
            v-model:current="currentPage"
            v-model:page-size="pageSize"
            :total="total"
            :show-size-changer="true"
            :show-quick-jumper="true"
            :show-total="(total: number, range: number[]) => `共 ${total} 条记录，当前显示 ${range[0]}-${range[1]} 条`"
            size="small"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="tsx">
  import { TableColumnType, Tag } from 'ant-design-vue';
  import { useTableHeight } from '@/hook';

  const totalEnterprise = ref(100);
  const totalTable = ref(100);

  const tableBoxRef = ref(null);
  const tableScroll = useTableHeight(tableBoxRef);

  // 分页相关变量
  const currentPage = ref(1); // 当前页码
  const pageSize = ref(10); // 每页显示条数
  const total = ref(5); // 总记录数

  const dataSource = Array.from({ length: 5 }).fill({
    key: '1',
    enterpriseName: '内蒙古伊核公司',
    age: 32
  });

  /**
   * 处理分页变化
   * @param page 当前页码
   * @param size 每页显示条数
   */
  const handlePageChange = (page: number, size: number) => {
    currentPage.value = page;
    pageSize.value = size;
    // 这里可以添加重新获取数据的逻辑
  };

  /**
   * 监听分页变化
   */
  watch([currentPage, pageSize], ([newPage, newSize]) => {
    handlePageChange(newPage, newSize);
  });

  const columns: TableColumnType[] = [
    {
      title: '企业名称',
      dataIndex: 'enterpriseName',
      key: 'enterpriseName',
      align: 'center'
    },
    {
      title: '2023年数据',
      dataIndex: 'data_2023',
      key: 'data_2023',
      align: 'center',
      customRender: opt => {
        return <>{opt.index === 0 ? <Tag>未导入</Tag> : <Tag color="success">已导入</Tag>}</>;
      }
    },
    {
      title: '2024年数据',
      dataIndex: 'data_2024',
      key: 'data_2024',
      align: 'center',
      customRender: opt => {
        return <>{opt.index === 0 ? <Tag>未导入</Tag> : <Tag color="success">已导入</Tag>}</>;
      }
    }
  ];

  // 处理导出按钮点击
  const handleExportClick = () => {
    // 导出当前表格数据
  };
</script>

<style scoped>
  .process-message {
    color: #096aa2;
    font-size: 15px;
    .number-text {
      color: #292ea7;
      font-weight: bold;
      margin: 0 10px;
    }
  }

  .pagination-container {
    position: absolute;
    bottom: 0;
    right: 0;
    padding: 10px;
  }
</style>
