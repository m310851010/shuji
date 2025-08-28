export const Table3Columns: any = [
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
];

export const TableAttachment2Columns = [
  {
    title: '省（市、区）',
    dataIndex: 'province_name',
    key: 'province_name',
    width: 120,
    align: 'center'
  },
  {
    title: '地市（州）',
    dataIndex: 'city_name',
    key: 'city_name',
    width: 120,
    align: 'center'
  },
  {
    title: '县（区）',
    dataIndex: 'country_name',
    key: 'country_name',
    width: 120,
    align: 'center'
  },
  {
    title: '年份',
    dataIndex: 'stat_date',
    key: 'stat_date',
    width: 80,
    align: 'center'
  },
  {
    title: '分品种煤炭消费摸底',
    align: 'center',
    children: [
      {
        title: '煤合计',
        dataIndex: 'total_coal',
        key: 'total_coal',
        width: 100,
        align: 'center'
      },
      {
        title: '原煤',
        dataIndex: 'raw_coal',
        key: 'raw_coal',
        width: 100,
        align: 'center'
      },
      {
        title: '洗精煤',
        dataIndex: 'washed_coal',
        key: 'washed_coal',
        width: 100,
        align: 'center'
      },
      {
        title: '其他煤炭',
        dataIndex: 'other_coal',
        key: 'other_coal',
        width: 100,
        align: 'center'
      }
    ]
  },
  {
    title: '分用途煤炭消费摸底',
    align: 'center',
    children: [
      {
        title: '能源加工转换',
        align: 'center',
        children: [
          {
            title: '1.火力发电',
            dataIndex: 'power_generation',
            key: 'power_generation',
            width: 100,
            align: 'center'
          },
          {
            title: '2.供热',
            dataIndex: 'heating',
            key: 'heating',
            width: 100,
            align: 'center'
          },
          {
            title: '3.煤炭洗选',
            dataIndex: 'coal_washing',
            key: 'coal_washing',
            width: 100,
            align: 'center'
          },
          {
            title: '4.炼焦',
            dataIndex: 'coking',
            key: 'coking',
            width: 100,
            align: 'center'
          },
          {
            title: '5.炼油及煤制油',
            dataIndex: 'oil_refining',
            key: 'oil_refining',
            width: 120,
            align: 'center'
          },
          {
            title: '6.制气',
            dataIndex: 'gas_production',
            key: 'gas_production',
            width: 100,
            align: 'center'
          }
        ]
      },
      {
        title: '终端消费',
        align: 'center',
        children: [
          {
            title: '1.工业',
            dataIndex: 'industry',
            key: 'industry',
            width: 100,
            align: 'center'
          },
          {
            title: '#用作原料、材料',
            dataIndex: 'raw_materials',
            key: 'raw_materials',
            width: 120,
            align: 'center'
          },
          {
            title: '2.其他用途',
            dataIndex: 'other_uses',
            key: 'other_uses',
            width: 100,
            align: 'center'
          }
        ]
      }
    ]
  },
  {
    title: '焦炭消费摸底',
    align: 'center',
    children: [
      {
        title: '焦炭',
        dataIndex: 'coke',
        key: 'coke',
        width: 100,
        align: 'center'
      }
    ]
  }
];
