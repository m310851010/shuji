<template>
  <a-space direction="vertical" size="large">
    <a-table
      v-for="(table, index) in tables"
      :key="index"
      :dataSource="table.dataSource"
      :columns="table.columns"
      size="small"
      bordered
      :pagination="false"
    />
  </a-space>
</template>

<script setup lang="tsx">
  import { TableColumnType } from 'ant-design-vue';
  import { CheckCircleFilled } from '@ant-design/icons-vue';
  import { newColumns } from '@/util';

  interface TableInfo {
    dataSource: any;
    columns: any[];
  }

  const tables = reactive<TableInfo[]>([]);

  const dataSource_1 = Array.from({ length: 5 }).fill({
    key: '1',
    enterpriseName: '内蒙古伊核公司',
    age: 32
  });

  tables.push({
    dataSource: dataSource_1,
    columns: newColumns<any>(
      { stat_date: '年份' },
      {
        title: () => (
          <>
            <span style="margin-right: 5px">单位名称</span>
            <CheckCircleFilled style={{ color: 'green' }} />
          </>
        ),
        key: 'unit_name'
      },
      {
        title: () => (
          <>
            <span style="margin-right: 5px">统一社会信用代码</span>

            <CheckCircleFilled style={{ color: 'green' }} />
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
        title: '单位所在省/市/区',
        customRender: opt => {
          const { province_name, city_name, country_name } = opt.record;
          return [province_name, city_name, country_name].filter(Boolean).join('/');
        }
      },
      {
        title: '单位所在地市',
        customRender: opt => {
          const { province_name, city_name, country_name } = opt.record;
          return [province_name, city_name, country_name].filter(Boolean).join('/');
        }
      },
      { tel: '联系电话' }
    )
  });

  tables.push({
    dataSource: [
      {
        annual_energy_equivalent_value: '年综合能耗当量值',
        annual_energy_equivalent_cost: '年综合能耗等价值',
        annual_raw_material_energy: '年原料用能消费量',
        annual_total_coal_consumption: '耗煤总量',
        annual_total_coal_products: '耗煤总量',
        annual_raw_coal: '原料用煤',
        annual_raw_coal_consumption: '原煤消费',
        annual_clean_coal_consumption: '洗精煤消费',
        annual_other_coal_consumption: '其他煤炭消费',
        annual_coke_consumption: '焦炭消费'
      }
    ],
    columns: [
      {
        title: '综合能源消费情况',
        children: newColumns<any>({
          annual_energy_equivalent_value: '年综合能耗当量值',
          annual_energy_equivalent_cost: '年综合能耗等价值',
          annual_raw_material_energy: '年原料用能消费量'
        })
      },
      {
        title: '煤炭消费情况',
        children: newColumns<any>({
          annual_total_coal_consumption: '耗煤总量',
          annual_total_coal_products: '耗煤总量',
          annual_raw_coal: '原料用煤',
          annual_raw_coal_consumption: '原煤消费',
          annual_clean_coal_consumption: '洗精煤消费',
          annual_other_coal_consumption: '其他煤炭消费',
          annual_coke_consumption: '焦炭消费'
        })
      }
    ]
  });

  tables.push({
    dataSource: [
      {
        row_no: '1',
        main_usage: '主要用途',
        specific_usage: '具体用途',
        input_variety: '投入品种',
        input_unit: '投入计量单位',
        input_quantity: '投入量',
        output_energy_types: '产出品种品类',
        measurement_unit: '产出计量单位',
        output_quantity: '产出量',
        remarks: '备注'
      }
    ],
    columns: [
      {
        title: '综合能源消费情况',
        children: newColumns<any>({
          row_no: '序号',
          main_usage: '主要用途',
          specific_usage: '具体用途',
          input_variety: '投入品种',
          input_unit: '投入计量单位',
          input_quantity: '投入量'
        })
      },
      {
        title: '煤炭消费情况',
        children: newColumns<any>({
          output_energy_types: '产出品种品类',
          measurement_unit: '产出计量单位',
          output_quantity: '产出量',
          remarks: '备注'
        })
      }
    ]
  });

  tables.push({
    dataSource: [
      {
        row_no: '1',
        equip_type: '类型',
        equip_no: '编号',
        total_runtime: '累计使用时间',
        design_life: '设计年限',
        energy_efficiency: '能效水平',
        capacity_unit: '容量单位',
        capacity: '容量',
        coal_type: '耗煤品种',
        annual_coal_consumption: '年耗煤量'
      }
    ],
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
          annual_coal_consumption: '年耗煤量'
        })
      }
    ]
  });
</script>
<style scoped></style>
