<template>
  <div class="table-section" v-for="(table, index) in tables" :key="index">
    <div class="info-grid" v-if="table.dataSource && table.dataSource.length > 0">
      <div class="info-item">
        <span class="info-label">单位名称</span>
        <span class="info-value">{{ table.dataSource[0]?.unit_name || '-' }}</span>
      </div>
      <div class="info-item">
        <span class="info-label">统一社会信用代码</span>
        <span class="info-value">{{ table.dataSource[0]?.credit_code || '-' }}</span>
      </div>
      <div class="info-item">
        <span class="info-label">数据年份</span>
        <span class="info-value">{{ table.dataSource[0]?.stat_date || '-' }}</span>
      </div>
      <div class="info-item">
        <span class="info-label">省份</span>
        <span class="info-value">{{ table.dataSource[0]?.province_name || '-' }}</span>
      </div>
      <div class="info-item">
        <span class="info-label">地市</span>
        <span class="info-value">{{ table.dataSource[0]?.city_name || '-' }}</span>
      </div>
      <div class="info-item">
        <span class="info-label">县/区</span>
        <span class="info-value">{{ table.dataSource[0]?.country_name || '-' }}</span>
      </div>
      <div class="info-item">
        <span class="info-label">行业门类</span>
        <span class="info-value">{{ table.dataSource[0]?.trade_a || '-' }}</span>
      </div>
      <div class="info-item">
        <span class="info-label">行业大类</span>
        <span class="info-value">{{ table.dataSource[0]?.trade_b || '-' }}</span>
      </div>
      <div class="info-item">
        <span class="info-label">行业中类</span>
        <span class="info-value">{{ table.dataSource[0]?.trade_c || '-' }}</span>
      </div>
    </div>
    <div class="table-wrapper">
      <a-table
        :dataSource="table.dataSource"
        :columns="table.columns"
        bordered
        size="small"
        :pagination="false"
        class="custom-table"
        :scroll="{ x: 'max-content' }"
      />
    </div>
  </div>
</template>

<script setup lang="tsx">
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
  watch(() => props.tableInfoList, (newData) => {
    if (newData && newData.length > 0) {
      // 清空现有表格
      tables.splice(0, tables.length);
      console.debug(props.tableInfoList, 'props.tableInfoList');
      
      // 添加新的表格配置
      tables.push({
        dataSource: props.tableInfoList,
        columns: [
          {
            title: '序号',
            dataIndex: 'row_no',
            key: 'row_no',
            align: 'center'
          },
          
          {
            title: '类型',
            dataIndex: 'coal_type',
            key: 'coal_type',
            align: 'center'
          },
          {
            title: '编号',
            dataIndex: 'coal_no',
            key: 'coal_no',
            align: 'center'
          },
          {
            title: '累计使用时间',
            dataIndex: 'usage_time',
            key: 'usage_time',
            align: 'center'
          },
          {
            title: '设计年限',
            dataIndex: 'design_life',
            key: 'design_life',
            align: 'center'
          },
          {
            title: '能效水平',
            dataIndex: 'enecrgy_efficienct_bmk',
            key: 'enecrgy_efficienct_bmk',
            align: 'center'
          },
          {
            title: '容量单位',
            dataIndex: 'capacity_unit',
            key: 'capacity_unit',
            align: 'center'
          },
          {
            title: '容量',
            dataIndex: 'capacity',
            key: 'capacity',
            align: 'center'
          },
          {
            title: '用途',
            dataIndex: 'use_info',
            key: 'use_info',
            align: 'center'
          },
          {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            align: 'center'
          },
          {
            title: '年耗煤量（单位：吨）',
            dataIndex: 'annual_coal_consumption',
            key: 'annual_coal_consumption',
            align: 'center'
          }
        ]
      });
    }
  }, { immediate: true });
</script>

<style scoped>
.table-section {
    background: #fff;
    overflow: hidden;
    margin-bottom: 20px;

}

.info-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 0;
    padding: 0;
    border: 1px solid #ddd;
    border-radius: 4px;
    overflow: hidden;
    margin-bottom: 12px;
}

.info-item {
    display: grid;
    grid-template-columns: 1fr 2fr;
    align-items: center;
    padding: 0;
    background: #fff;
    border-right: 1px solid #ddd;
    border-bottom: 1px solid #ddd;
    transition: all 0.2s ease;
    position: relative;
    min-height: 48px;
}

.info-item:nth-child(3n) {
    border-right: none;
}

.info-item:nth-last-child(-n+3) {
    border-bottom: none;
}

.info-label {
    font-size: 12px;
    color: #333;
    font-weight: 600;
    line-height: 1.2;
    height: 100%;
    background: #f5f5f5;
    padding: 12px 16px;
    border-right: 1px solid #ddd;
    display: flex;
    align-items: center;
    margin: 0;
    white-space: nowrap;
    width: 120px;
    min-width: 120px;
    flex-shrink: 0;
}

.info-value {
    font-size: 11px;
    color: #333;
    font-weight: normal;
    word-break: break-all;
    line-height: 1.2;
    text-align: left;
    background: #fff;
    padding: 12px 16px;
    display: flex;
    align-items: center;
    margin: 0;
}

.table-wrapper {
  margin-top: 20px;
}

.custom-table {
  border-radius: 6px;
  overflow: hidden;
}


/* Ant Design Table 单元格样式 */
:deep(.ant-table-tbody > tr > td) {
  text-align: center !important;
}

/* 强制保持三列布局，不使用响应式 */
.info-grid {
  grid-template-columns: repeat(3, 1fr) !important;
}
</style>
