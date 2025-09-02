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
  import { newColumns } from '@/util';

  /**
   * 组件属性定义
   */
  const props = defineProps({
    tableInfoList: {
      type: Array as PropType<Array<Record<string, any>>>,
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

        // 第一个表格：基本信息表格（使用索引0的数据）
        if (newData[0] && Array.isArray(newData[0])) {
          tables.push({
            dataSource: newData[0],
            columns: newColumns<any>(
              { stat_date: '年份' },
              {
                title: () => (
                  <>
                    <span style="margin-right: 5px">单位名称</span>
                  </>
                ),
                key: 'unit_name'
              },
              {
                title: () => (
                  <>
                    <span style="margin-right: 5px">统一社会信用代码</span>
                  </>
                ),
                key: 'credit_code'
              },
              {
                trade_a: '行业门类',
                trade_b: '行业大类',
                trade_c: '行业中类'
              },
              {
                title: '省份',
                dataIndex: 'province_name',
                key: 'province_name'
              },
              {
                title: '城市',
                dataIndex: 'city_name',
                key: 'city_name'
              },
              {
                title: '区县',
                dataIndex: 'country_name',
                key: 'country_name'
              },
              { tel: '联系电话' }
            )
          });
        }

        // 第二个表格：综合能源消费和煤炭消费情况（使用索引0的数据）
        if (newData[0] && Array.isArray(newData[0])) {
          tables.push({
            dataSource: newData[0],
            columns: [
              {
                title: '综合能源消费情况',
                children: newColumns<any>({
                  annual_energy_equivalent_value: '年综合能耗当量值（万吨标准煤，含原料用能）',
                  annual_energy_equivalent_cost: '年综合能耗等价值（万吨标准煤，含原料用能）',
                  annual_raw_material_energy: '年原料用能消费量（万吨标准煤）'
                })
              },
              {
                title: '煤炭消费情况',
                children: newColumns<any>({
                  annual_total_coal_consumption: '耗煤总量（实物量，万吨）',
                  annual_total_coal_products: '耗煤总量（标准量，万吨标准煤）',
                  annual_raw_coal: '原料用煤（实物量，万吨）',
                  annual_raw_coal_consumption: '原煤消费（实物量，万吨）',
                  annual_clean_coal_consumption: '洗精煤消费（实物量，万吨）',
                  annual_other_coal_consumption: '其他煤炭消费（实物量，万吨）',
                  annual_coke_consumption: '焦炭消费（实物量，万吨）'
                })
              }
            ]
          });
        }

        // 第三个表格：煤炭消费主要用途情况（使用索引1的数据）
        if (newData[1] && Array.isArray(newData[1])) {
          tables.push({
            dataSource: newData[1],
            columns: [
              {
                title: '煤炭消费主要用途情况',
                children: newColumns<any>({
                  row_no: '序号',
                  main_usage: '主要用途',
                  specific_usage: '具体用途',
                  input_variety: '投入品种',
                  input_unit: '投入计量单位',
                  input_quantity: '投入量',
                  output_energy_types: '产出品种品类',
                  measurement_unit: '产出计量单位',
                  output_quantity: '产出量',
                  remarks: '备注'
                })
              }
            ]
          });
        }

        // 第四个表格：重点耗煤装置（设备）情况（使用索引2的数据）
        if (newData[2] && Array.isArray(newData[2])) {
          tables.push({
            dataSource: newData[2],
            columns: [
              {
                title: '重点耗煤装置（设备)情况',
                children: newColumns<any>({
                  row_no: '序号',
                  equip_type: '类型',
                  equip_no: '编号',
                  total_runtime: '累计使用时间',
                  design_life: '设计年限',
                  energy_efficiency: '能效水平',
                  capacity_unit: '容量单位',
                  capacity: '容量',
                  coal_type: '耗煤品种',
                  annual_coal_consumption: '年耗煤量（单位：吨）'
                })
              }
            ]
          });
        }
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
