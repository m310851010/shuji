<template>
  <div class="db-merge-cover-table">
    <a-table
      :dataSource="conflictData"
      :columns="columns"
      :rowKey="(record: any) => record.key"
      size="small"
      bordered
      :pagination="false"
      :scroll="{ y: 400 }"
    >
      <!-- 自定义表头 -->
      <template #headerCell="{ column }">
        <template v-if="column.key.startsWith('db')">
          <div class="db-header-cell">
            <div class="db-name">{{ column.title }}</div>
            <a-checkbox
              :checked="allSelectedStates[getDbIndex(column.key)] || false"
              :indeterminate="indeterminateStates[getDbIndex(column.key)] || false"
              :disabled="!hasAnyConflict"
              @change="(checked: boolean) => handleSelectAllDb(getDbIndex(column.key), checked)"
            >
              全选
            </a-checkbox>
          </div>
        </template>
        <template v-else>
          {{ column.title }}
        </template>
      </template>

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
          <span class="detail-label">{{ getFieldLabel(String(key)) }}:</span>
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

interface ConflictSourceInfo {
  filePath: string;
  fileName: string;
  tableType: string;
  obj_ids: string[];
}

interface ConflictDetail {
  credit_code?: string;
  stat_date?: string;
  unit_name?: string;
  project_name?: string;
  project_code?: string;
  document_number?: string;
  province_name?: string;
  city_name?: string;
  country_name?: string;
  conflict: ConflictSourceInfo[];
}

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
  conflictDetail: ConflictDetail; // 冲突详情
}

interface DetailModal {
  visible: boolean;
  title: string;
  data: any;
}

const props = defineProps<{
  conflictList: ConflictDetail[]; // 冲突列表
  dbFileNames: string[]; // 数据库文件名数组
  tableType: TableType;
}>();

const emit = defineEmits<{
  selectionChange: [selectedData: Record<string, ConflictSourceInfo[]>];
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

const allSelectedStates = ref<boolean[]>([]);

// 初始化全选状态
watch(() => props.dbFileNames, (newDbFileNames) => {
  if (newDbFileNames && newDbFileNames.length > 0) {
    allSelectedStates.value = newDbFileNames.map(() => true); // 默认全选
  } else {
    allSelectedStates.value = [];
  }
}, { immediate: true });

// 更新全选状态的computed
const updateAllSelectedStates = () => {
  props.dbFileNames.forEach((_, index) => {
    const conflictItems = conflictData.value.filter(item => item.hasConflict);
    if (conflictItems.length > 0) {
      allSelectedStates.value[index] = conflictItems.every(item => item.selections[`db${index}`]);
    }
  });
};

const indeterminateStates = computed(() => {
  return props.dbFileNames.map((_, index) => {
    const conflictItems = conflictData.value.filter(item => item.hasConflict);
    const selectedCount = conflictItems.filter(item => item.selections[`db${index}`]).length;
    return selectedCount > 0 && selectedCount < conflictItems.length;
  });
});

// 从列key中获取数据库索引
const getDbIndex = (columnKey: string): number => {
  const index = parseInt(columnKey.replace('db', ''));
  // 确保索引在有效范围内
  if (index >= 0 && index < allSelectedStates.value.length) {
    return index;
  }
  return 0; // 默认返回0
};

const processConflictData = (conflicts: ConflictDetail[]) => {
  const processedData: ConflictRecord[] = [];
  
  conflicts.forEach((conflict, index) => {
    if (conflict.conflict && conflict.conflict.length > 0) {
      // 初始化选择状态，默认全部选中
      const selections: Record<string, boolean> = {};
      
      props.dbFileNames.forEach((_, dbIndex) => {
        const dbKey = `db${dbIndex}`;
        selections[dbKey] = true; // 默认选中
      });
      
      // 根据表类型设置不同的字段
      const record: ConflictRecord = {
        key: `conflict_${index}`,
        hasConflict: true,
        selections,
        conflictDetail: conflict
      };

      switch (props.tableType) {
        case 'table1':
        case 'table2':
          record.unitName = conflict.unit_name || '';
          record.creditCode = conflict.credit_code || '';
          record.statDate = conflict.stat_date || '';
          break;
        case 'table3':
          record.projectName = conflict.project_name || '';
          record.projectCode = conflict.project_code || '';
          record.reviewNumber = conflict.document_number || '';
          break;
        case 'attachment2':
          record.provinceName = conflict.province_name || '';
          record.cityName = conflict.city_name || '';
          record.countryName = conflict.country_name || '';
          record.statDate = conflict.stat_date || '';
          break;
      }
      
      processedData.push(record);
    }
  });
  
  return processedData;
};

watch(() => props.conflictList, (newConflicts) => {
  conflictData.value = processConflictData(newConflicts);
  // 更新全选状态
  updateAllSelectedStates();
}, { immediate: true, deep: true });

const handleSelectAllDb = (dbIndex: number, checked: boolean) => {
  const dbKey = `db${dbIndex}`;
  conflictData.value.forEach(item => {
    if (item.hasConflict) {
      item.selections[dbKey] = checked;
    }
  });
  // 更新全选状态
  allSelectedStates.value[dbIndex] = checked;
  emitSelectionChange();
};

const handleDbSelect = (record: ConflictRecord, dbKey: string) => {
  // 更新全选状态
  updateAllSelectedStates();
  emitSelectionChange();
};

const handleViewData = (record: ConflictRecord, dbKey: string) => {
  const dbIndex = parseInt(dbKey.replace('db', ''));
  const fileName = props.dbFileNames[dbIndex];
  
  // 查找对应的冲突源信息
  const conflictSource = record.conflictDetail.conflict.find(
    source => source.fileName === fileName
  );
  
  if (conflictSource) {
    // 根据表类型生成不同的标题
    let title = '';
    switch (props.tableType) {
      case TableType.table1:
      case TableType.table2:
        title = `${record.unitName || '未知企业'} - ${fileName} 数据详情`;
        break;
      case TableType.table3:
        title = `${record.projectName || '未知项目'} - ${fileName} 数据详情`;
        break;
      case TableType.attachment2:
        title = `${record.provinceName || ''}${record.cityName || ''}${record.countryName || ''} - ${fileName} 数据详情`;
        break;
    }
    
    detailModal.value = {
      visible: true,
      title: title,
             data: {
         filePath: conflictSource.filePath,
         fileName: conflictSource.fileName,
         tableType: conflictSource.tableType,
         obj_ids: conflictSource.obj_ids
       }
    };
  } else {
    message.warning('暂无数据可查看');
  }
};

const emitSelectionChange = () => {
  // 构建符合 MergeConflictData 函数参数要求的数据结构
  const selectedConflicts: Record<string, ConflictSourceInfo[]> = {};
  
  // 按表类型分组选中的冲突数据
  props.dbFileNames.forEach((fileName, dbIndex) => {
    const dbKey = `db${dbIndex}`;
    const selectedItems = conflictData.value.filter(item => 
      item.hasConflict && item.selections[dbKey]
    );
    
    if (selectedItems.length > 0) {
      const tableType = props.tableType;
      if (!selectedConflicts[tableType]) {
        selectedConflicts[tableType] = [];
      }
      
      // 收集该文件中所有选中的冲突源信息
      selectedItems.forEach(item => {
        const conflictSource = item.conflictDetail.conflict.find(
          source => source.fileName === fileName
        );
        if (conflictSource) {
          // 检查是否已经存在相同的文件路径
          const existingIndex = selectedConflicts[tableType].findIndex(
            existing => existing.filePath === conflictSource.filePath
          );
          if (existingIndex >= 0) {
            // 合并 obj_ids
            selectedConflicts[tableType][existingIndex].obj_ids.push(...conflictSource.obj_ids);
          } else {
            // 添加新的冲突源信息
            selectedConflicts[tableType].push({
              filePath: conflictSource.filePath,
              fileName: conflictSource.fileName,
              tableType: conflictSource.tableType,
              obj_ids: [...conflictSource.obj_ids]
            });
          }
        }
      });
    }
  });
  
  emit('selectionChange', selectedConflicts);
};

// 字段标签映射
const getFieldLabel = (key: string): string => {
  const labelMap: Record<string, string> = {
    filePath: '文件路径',
    fileName: '文件名',
    tableType: '表类型',
    obj_ids: '对象ID列表',
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
    document_number: '审查意见文号',
    unit_level: '单位等级',
    construction_unit: '建设单位'
  };
  return labelMap[key] || key;
};

// 暴露方法给父组件
defineExpose({
  getSelectedData: (): Record<string, ConflictSourceInfo[]> => {
    // 构建符合 MergeConflictData 函数参数要求的数据结构
    const selectedConflicts: Record<string, ConflictSourceInfo[]> = {};
    
    // 按表类型分组选中的冲突数据
    props.dbFileNames.forEach((fileName, dbIndex) => {
      const dbKey = `db${dbIndex}`;
      const selectedItems = conflictData.value.filter(item => 
        item.hasConflict && item.selections[dbKey]
      );
      
      if (selectedItems.length > 0) {
        const tableType = props.tableType;
        if (!selectedConflicts[tableType]) {
          selectedConflicts[tableType] = [];
        }
        
        // 收集该文件中所有选中的冲突源信息
        selectedItems.forEach(item => {
          const conflictSource = item.conflictDetail.conflict.find(
            source => source.fileName === fileName
          );
          if (conflictSource) {
            // 检查是否已经存在相同的文件路径
            const existingIndex = selectedConflicts[tableType].findIndex(
              existing => existing.filePath === conflictSource.filePath
            );
            if (existingIndex >= 0) {
              // 合并 obj_ids
              selectedConflicts[tableType][existingIndex].obj_ids.push(...conflictSource.obj_ids);
            } else {
              // 添加新的冲突源信息
              selectedConflicts[tableType].push({
                filePath: conflictSource.filePath,
                fileName: conflictSource.fileName,
                tableType: conflictSource.tableType,
                obj_ids: [...conflictSource.obj_ids]
              });
            }
          }
        });
      }
    });
    
    return selectedConflicts;
  },
  clearSelection: () => {
    conflictData.value.forEach(item => {
      Object.keys(item.selections).forEach(key => {
        item.selections[key] = false;
      });
    });
    // 重置全选状态
    allSelectedStates.value = allSelectedStates.value.map(() => false);
    emitSelectionChange();
  }
});
</script>

<style scoped lang="less">

</style>
