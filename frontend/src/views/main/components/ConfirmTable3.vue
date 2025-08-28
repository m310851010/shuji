<template>
  <a-space direction="vertical" size="large">
    <a-table
      v-for="(table, index) in tables"
      :key="index"
      :dataSource="table.dataSource"
      :columns="table.columns"
      bordered
      :pagination="false"
      :scroll="{ x: 'max-content' }"
    />
  </a-space>
</template>

<script setup lang="tsx">
  import { Table3Columns } from '@/views/main/components/columns';

  /**
   * 组件属性定义
   */
  const props = defineProps({
    tableInfoList: {
      type: Array,
      default: () => []
    }
  });

  /**
   * 表格信息接口定义
   */
  interface TableInfo {
    dataSource: any;
    columns: any[];
  }

  const tables = reactive<TableInfo[]>([]);

  /**
   * 格式化时间字符串，将ISO格式转换为友好显示格式
   * @param timeStr 时间字符串，可能包含T分隔符
   * @returns 格式化后的时间字符串，格式为：YYYY年MM月DD日 HH:mm:ss
   */
  const formatDateTime = (timeStr: string) => {
    if (!timeStr) return '';

    // 如果包含T，说明是ISO格式，需要格式化
    if (timeStr.includes('T')) {
      try {
        const date = new Date(timeStr);
        // 格式化为中文日期格式：YYYY年MM月DD日 HH:mm:ss
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

  /**
   * 监听props数据变化，更新表格数据
   */
  watch(
    () => props.tableInfoList,
    newData => {
      if (newData && newData.length > 0) {
        // 清空现有表格
        tables.splice(0, tables.length);
        console.debug(props.tableInfoList, 'props.tableInfoList');

        // 添加新的表格配置
        tables.push({
          dataSource: props.tableInfoList,
          columns: Table3Columns
        });
      }
    },
    { immediate: true }
  );
</script>

<style scoped>
  /* 让表格容器占满屏幕宽度 */
  .ant-space {
    width: 99%;
    margin: 0;
    padding: 16px;
    box-sizing: border-box;
  }

  /* 让表格占满容器宽度 */
  :deep(.ant-table) {
    width: 100%;
    min-width: 100%;
  }

  /* 确保表格容器不会溢出 */
  :deep(.ant-table-container) {
    width: 100%;
    overflow-x: auto;
  }

  /* 设置表格内容区域的最小宽度 */
  :deep(.ant-table-content) {
    width: 100%;
    min-width: 100%;
  }

  /* 让表格头部和主体都占满宽度 */
  :deep(.ant-table-thead),
  :deep(.ant-table-tbody) {
    width: 100%;
  }
</style>
