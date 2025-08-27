<template>
  <div class="db-merge-cover-table">
    <!-- 表头全选控制 -->
    <div class="table-header-controls">
      <div 
        v-for="(fileName, index) in dbFileNames" 
        :key="index"
        class="db-controls"
      >
        <span>{{ fileName }}</span>
        <a-checkbox
          v-model:checked="allSelectedStates[index]"
          :indeterminate="indeterminateStates[index]"
          :disabled="!hasAnyConflict"
          @change="(checked: boolean) => handleSelectAllDb(index, checked)"
        >
          全选
        </a-checkbox>
      </div>
    </div>

    <a-table
      :dataSource="conflictData"
      :columns="columns"
      :rowKey="(record: any) => record.key"
      size="small"
      bordered
      :pagination="false"
      :scroll="{ y: 400 }"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key.startsWith('db')">
          <div class="db-column">
            <a-button 
              type="link" 
              size="small" 
              @click="handleViewData(record, column.key)"
            >
              查看
            </a-button>
            <a-checkbox
              v-model:checked="record.selections[column.key]"
              :disabled="!record.hasConflict"
              @change="() => handleDbSelect(record, column.key)"
            />
          </div>
        </template>
      </template>
    </a-table>

    <!-- 查看数据详情模态框 -->
    <a-modal
      v-model:open="detailModal.visible"
      :title="detailModal.title"
      width="80%"
      :footer="null"
      @cancel="detailModal.visible = false"
    >
      <div v-if="detailModal.data" class="detail-content">
        <div 
          v-for="(value, key) in detailModal.data" 
          :key="key"
          class="detail-item"
        >
          <span class="detail-label">{{ getFieldLabel(key) }}:</span>
          <span class="detail-value">{{ value }}</span>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue';
import {message, TableColumnType} from 'ant-design-vue';
import {TableType} from '@/views/constant';

interface ConflictRecord {
  key: string;
  unitName?: string;
  creditCode?: string;
  statDate?: string;
  projectName?: string;
  projectCode?: string;
  reviewNumber?: string;
  unitLevel?: string;
  provinceName?: string;
  cityName?: string;
  countryName?: string;
  hasConflict: boolean;
  selections: Record<string, boolean>; // 动态选择状态
  dataMap: Record<string, any>; // 动态数据映射
}

interface DetailModal {
  visible: boolean;
  title: string;
  data: any;
}

const props = defineProps<{
  conflictList: any[];
  dbFileNames: string[]; // 数据库文件名数组
  tableType: TableType;
}>();

const emit = defineEmits<{
  selectionChange: [selectedData: any[]];
}>();

const conflictData = ref<ConflictRecord[]>([]);
const detailModal = ref<DetailModal>({
  visible: false,
  title: '',
  data: null
});

const columns = computed((): TableColumnType[] => {
  let baseColumns: TableColumnType[] = [];

  switch (props.tableType) {
    case TableType.table1:
    case TableType.table2:
      baseColumns = [
        {
          title: '企业名称',
          dataIndex: 'unitName',
          key: 'unitName',
          width: '200px',
          ellipsis: true
        },
        {
          title: '企业代码',
          dataIndex: 'creditCode',
          key: 'creditCode',
          width: '150px'
        },
        {
          title: '年份',
          dataIndex: 'statDate',
          key: 'statDate',
          width: '100px'
        }
      ];
      break;
    case TableType.table3:
      baseColumns = [
        {
          title: '项目名称',
          dataIndex: 'projectName',
          key: 'projectName',
          width: '200px',
          ellipsis: true
        },
        {
          title: '项目代码',
          dataIndex: 'projectCode',
          key: 'projectCode',
          width: '150px'
        },
        {
          title: '审查意见文号',
          dataIndex: 'reviewNumber',
          key: 'reviewNumber',
          width: '150px'
        }
      ];
      break;
    case TableType.attachment2:
      baseColumns = [
        {
          title: '省',
          dataIndex: 'provinceName',
          key: 'provinceName',
          width: '150px',
          ellipsis: true
        },
        {
          title: '市',
          dataIndex: 'cityName',
          key: 'cityName',
          width: '150px',
          ellipsis: true
        },
        {
          title: '县',
          dataIndex: 'countryName',
          key: 'countryName',
          width: '150px',
          ellipsis: true
        },
        {
          title: '年份',
          dataIndex: 'statDate',
          key: 'statDate',
          width: '100px'
        }
      ];
      break;
  }

  props.dbFileNames.forEach((fileName, index) => {
    baseColumns.push({
      title: fileName,
      key: `db${index}`,
      width: '200px'
    });
  });

  return baseColumns;
});

const hasAnyConflict = computed(() => {
  return conflictData.value.some(item => item.hasConflict);
});

const allSelectedStates = computed(() => {
  return props.dbFileNames.map((_, index) => {
    const conflictItems = conflictData.value.filter(item => item.hasConflict);
    return conflictItems.length > 0 && conflictItems.every(item => item.selections[`db${index}`]);
  });
});

const indeterminateStates = computed(() => {
  return props.dbFileNames.map((_, index) => {
    const conflictItems = conflictData.value.filter(item => item.hasConflict);
    const selectedCount = conflictItems.filter(item => item.selections[`db${index}`]).length;
    return selectedCount > 0 && selectedCount < conflictItems.length;
  });
});

const processConflictData = (conflicts: any[]) => {
  const processedData: ConflictRecord[] = [];
  
  conflicts.forEach((conflict, index) => {
    if (conflict.SourceData && conflict.SourceData.length > 0) {
      const sourceData = conflict.SourceData[0];
      
      // 初始化选择状态和数据映射
      const selections: Record<string, boolean> = {};
      const dataMap: Record<string, any> = {};
      
      props.dbFileNames.forEach((_, dbIndex) => {
        const dbKey = `db${dbIndex}`;
        selections[dbKey] = false;
        // 这里需要根据实际的冲突数据结构来设置数据
        dataMap[dbKey] = conflict.SourceData?.[dbIndex] || null;
      });
      
      // 根据表类型设置不同的字段
      const record: ConflictRecord = {
        key: `${conflict.TableName}_${index}`,
        hasConflict: true,
        selections,
        dataMap
      };

      switch (props.tableType) {
        case 'table1':
        case 'table2':
          record.unitName = sourceData.unit_name || '';
          record.creditCode = sourceData.credit_code || '';
          record.statDate = sourceData.stat_date || '';
          break;
        case 'table3':
          record.projectName = sourceData.project_name || '';
          record.projectCode = sourceData.project_code || '';
          record.reviewNumber = sourceData.review_number || '';
          break;
        case 'attachment2':
          record.provinceName = sourceData.province_name || '';
          record.cityName = sourceData.city_name || '';
          record.countryName = sourceData.country_name || '';
          record.statDate = sourceData.stat_date || '';
          break;
      }
      
      processedData.push(record);
    }
  });
  
  return processedData;
};

watch(() => props.conflictList, (newConflicts) => {
  conflictData.value = processConflictData(newConflicts);
}, { immediate: true, deep: true });

const handleSelectAllDb = (dbIndex: number, checked: boolean) => {
  const dbKey = `db${dbIndex}`;
  conflictData.value.forEach(item => {
    if (item.hasConflict) {
      item.selections[dbKey] = checked;
    }
  });
  emitSelectionChange();
};

const handleDbSelect = (record: ConflictRecord, dbKey: string) => {
  emitSelectionChange();
};

const handleViewData = (record: ConflictRecord, dbKey: string) => {
  const data = record.dataMap[dbKey];
  if (data) {
    const dbIndex = parseInt(dbKey.replace('db', ''));
    // 根据表类型生成不同的标题
    let title = '';
    switch (props.tableType) {
      case TableType.table1:
      case TableType.table2:
        title = `${record.unitName || '未知企业'} - Db${dbIndex + 1}（${props.dbFileNames[dbIndex]}） 数据详情`;
        break;
      case TableType.table3:
        title = `${record.projectName || '未知项目'} - Db${dbIndex + 1}（${props.dbFileNames[dbIndex]}） 数据详情`;
        break;
      case TableType.attachment2:
        title = `${record.provinceName || ''}${record.cityName || ''}${record.countryName || ''} - Db${dbIndex + 1}（${props.dbFileNames[dbIndex]}） 数据详情`;
        break;
    }
    detailModal.value = {
      visible: true,
      title: title,
      data: data
    };
  } else {
    message.warning('暂无数据可查看');
  }
};

const emitSelectionChange = () => {
  const selectedData = conflictData.value.filter(item => 
    item.hasConflict && Object.values(item.selections).some(selected => selected)
  );
  emit('selectionChange', selectedData);
};

// 字段标签映射
const getFieldLabel = (key: string): string => {
  const labelMap: Record<string, string> = {
    unit_name: '企业名称',
    credit_code: '统一社会信用代码',
    stat_date: '统计日期',
    province_name: '省份',
    city_name: '城市',
    country_name: '区县',
    equip_type: '设备类型',
    coal_type: '煤炭类型',
    energy_efficiency: '能源效率',
    main_usage: '主要用途',
    specific_usage: '具体用途',
    input_variety: '投入品种',
    output_energy_types: '产出能源类型',
    project_name: '项目名称',
    project_code: '项目代码',
    review_number: '审查意见文号',
    unit_level: '单位等级',
    construction_unit: '建设单位'
  };
  return labelMap[key] || key;
};

// 暴露方法给父组件
defineExpose({
  getSelectedData: () => {
    return conflictData.value.filter(item => 
      item.hasConflict && Object.values(item.selections).some(selected => selected)
    );
  },
  clearSelection: () => {
    conflictData.value.forEach(item => {
      Object.keys(item.selections).forEach(key => {
        item.selections[key] = false;
      });
    });
    emitSelectionChange();
  }
});
</script>

<style scoped lang="less">
.db-merge-cover-table {
  .table-header-controls {
    display: flex;
    justify-content: space-between;
    margin-bottom: 16px;
    padding: 12px;
    background-color: #f5f5f5;
    border-radius: 6px;
    flex-wrap: wrap;
    gap: 16px;

    .db-controls {
      display: flex;
      align-items: center;
      gap: 8px;
      font-weight: 500;
      min-width: 200px;
    }
  }

  .db-column {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .detail-content {
    max-height: 400px;
    overflow-y: auto;
    
    .detail-item {
      display: flex;
      padding: 8px 0;
      border-bottom: 1px solid #f0f0f0;
      
      &:last-child {
        border-bottom: none;
      }
      
      .detail-label {
        font-weight: 500;
        min-width: 120px;
        color: #666;
      }
      
      .detail-value {
        flex: 1;
        word-break: break-all;
      }
    }
  }

  :deep(.ant-table-thead > tr > th) {
    background-color: #fafafa;
    font-weight: 600;
  }

  :deep(.ant-checkbox-disabled) {
    opacity: 0.5;
  }
}
</style>
