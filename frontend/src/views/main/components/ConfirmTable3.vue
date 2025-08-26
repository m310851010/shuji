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
            title: '项目名称',
            dataIndex: 'project_name',
            key: 'project_name',
            align: 'center'
          },
          {
            title: '项目代码',
            dataIndex: 'project_code',
            key: 'project_code',
            align: 'center'
          },
          {
            title: '建设单位',
            dataIndex: 'construction_unit',
            key: 'construction_unit',
            align: 'center'
          },
          {
            title: '主要建设内容',
            dataIndex: 'main_construction_content',
            key: 'main_construction_content',
            align: 'center'
          },
          {
            title: '项目所在省',
            dataIndex: 'province_name',
            key: 'province_name',
            align: 'center'
          },
          {
            title: '项目所在地市',
            dataIndex: 'city_name',
            key: 'city_name',
            align: 'center'
          },
          {
            title: '项目所在区县',
            dataIndex: 'country_name',
            key: 'country_name',
            align: 'center'
          },
          {
            title: '所属行业大类',
            dataIndex: 'trade_a',
            key: 'trade_a',
            align: 'center'
          },
          {
            title: '所属行业小类',
            dataIndex: 'trade_c',
            key: 'trade_c',
            align: 'center'
          },
          {
            title: '节能审查批复时间',
            dataIndex: 'examination_approval_time',
            key: 'examination_approval_time',
            align: 'center'
          },
          {
            title: '拟投产时间',
            dataIndex: 'scheduled_time',
            key: 'scheduled_time',
            align: 'center'
          },
          {
            title: '实际投产时间',
            dataIndex: 'actual_time',
            key: 'actual_time',
            align: 'center'
          },
          {
            title: '节能审查机关',
            dataIndex: 'examination_authority',
            key: 'examination_authority',
            align: 'center'
          },
          {
            title: '审查意见文号',
            dataIndex: 'document_number',
            key: 'document_number',
            align: 'center'
          },
          {
            title: '年综合能源消费量（万吨标准煤）',
            align: 'center',
            children: [
              {
                title: '当量值',
                dataIndex: 'equivalent_value',
                key: 'equivalent_value',
                align: 'center'
              },
              {
                title: '等价值',
                dataIndex: 'equivalent_cost',
                key: 'equivalent_cost',
                align: 'center'
              }
            ]
          },
          {
            title: '年煤品消费量（万吨，实物量）',
            align: 'center',
            children: [
              {
                title: '煤品消费总量',
                dataIndex: 'pq_total_coal_consumption',
                key: 'pq_total_coal_consumption',
                align: 'center'
              },
              {
                title: '煤炭消费量',
                dataIndex: 'pq_coal_consumption',
                key: 'pq_coal_consumption',
                align: 'center'
              },
              {
                title: '焦炭消费量',
                dataIndex: 'pq_coke_consumption',
                key: 'pq_coke_consumption',
                align: 'center'
              },
              {
                title: '兰炭消费量',
                dataIndex: 'pq_blue_coke_consumption',
                key: 'pq_blue_coke_consumption',
                align: 'center'
              }
            ]
          },
          {
            title: '年煤品消费量（万吨标准煤，折标量）',
            align: 'center',
            children: [
              {
                title: '煤品消费总量',
                dataIndex: 'sce_total_coal_consumption',
                key: 'sce_total_coal_consumption',
                align: 'center'
              },
              {
                title: '煤炭消费量',
                dataIndex: 'sce_coal_consumption',
                key: 'sce_coal_consumption',
                align: 'center'
              },
              {
                title: '焦炭消费量',
                dataIndex: 'sce_coke_consumption',
                key: 'sce_coke_consumption',
                align: 'center'
              },
              {
                title: '兰炭消费量',
                dataIndex: 'sce_blue_coke_consumption',
                key: 'sce_blue_coke_consumption',
                align: 'center'
              }
            ]
          },
          {
            title: '煤炭消费替代情况',
            align: 'center',
            children: [
              {
                title: '是否煤炭消费替代',
                dataIndex: 'is_substitution',
                key: 'is_substitution',
                align: 'center'
              },
              {
                title: '煤炭消费替代来源',
                dataIndex: 'substitution_source',
                key: 'substitution_source',
                align: 'center'
              },
              {
                title: '煤炭消费替代量',
                dataIndex: 'substitution_quantity',
                key: 'substitution_quantity',
                align: 'center'
              }
            ]
          },
          {
            title: '原料用煤情况',
            align: 'center',
            children: [
              {
                title: '年原料用煤量（万吨，实物量）',
                dataIndex: 'pq_annual_coal_quantity',
                key: 'pq_annual_coal_quantity',
                align: 'center'
              },
              {
                title: '年原料用煤量（万吨标准煤，折标量）',
                dataIndex: 'sce_annual_coal_quantity',
                key: 'sce_annual_coal_quantity',
                align: 'center'
              }
            ]
          }
        ]
      });
    }
  }, { immediate: true });
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
