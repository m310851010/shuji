<template>
  <div class="db-merge-cover-table">
    <a-table
      :dataSource="conflictData"
      :columns="columns"
      :rowKey="(record: any) => record.key"
      size="small"
      bordered
      :pagination="false"
      :scroll="{ x: 'max-content', y: 400 }"
    >
      <!-- 自定义表头 -->
      <template #headerCell="{ column }">
        <template v-if="column.key.startsWith('db')">
          <div class="db-header-cell">
            <div class="db-name">{{ column.title }}</div>
            <span style="margin: 0 8px">全选</span>
            <a-checkbox
              :checked="getAllSelectedState(column.key)"
              :indeterminate="getIndeterminateState(column.key)"
              :disabled="!hasAnyConflict"
              @change="(e: any) => handleSelectAllDb(getDbIndex(column.key), e.target.checked)"
            ></a-checkbox>
          </div>
        </template>
        <template v-else>
          {{ column.title }}
        </template>
      </template>

      <template #bodyCell="{ column, record }">
        <template v-if="column.key.startsWith('db')">
          <div class="db-column" v-if="record[column.title]">
            <a-button type="link" size="small" @click="handleViewDetailData(record, column.key)">查看</a-button>
            <a-checkbox
              v-model:checked="record.selections[column.key]"
              :disabled="!record.hasConflict"
              @change="() => handleDbSelect(record, column.key)"
            />
          </div>
          <div class="db-column" v-else>无</div>
        </template>
      </template>
    </a-table>

    <!-- 查看详细数据模态框 -->
    <a-modal
      v-model:open="confirmModal.visible"
      :bodyStyle="{ paddingTop: 0 }"
      class="full-screen-modal button-middle"
      :title="confirmModal.title"
      :cancel-button-props="{ style: 'display: none' }"
      @ok="confirmModal.handleOk"
      ok-text="关闭"
    >
      <div class="wh-100 relative">
        <div class="abs" style="overflow: auto">
          <ConfirmTable1 v-if="confirmModal.tableType == 'table1'" :tableInfoList="confirmModal.tableInfoList" />
          <ConfirmTable2 v-if="confirmModal.tableType == 'table2'" :tableInfoList="confirmModal.tableInfoList" />
          <ConfirmTable3 v-if="confirmModal.tableType == 'table3'" :tableInfoList="confirmModal.tableInfoList" />
          <ConfirmTable4 v-if="confirmModal.tableType == 'attachment2'" :tableInfoList="confirmModal.tableInfoList" />
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { message, TableColumnType } from 'ant-design-vue';
  import { TableType, TableTypeName } from '@/views/constant';
  import ConfirmTable1 from './ConfirmTable1.vue';
  import ConfirmTable2 from './ConfirmTable2.vue';
  import ConfirmTable3 from './ConfirmTable3.vue';
  import ConfirmTable4 from './ConfirmTable4.vue';
  import {
    QueryDataDetailTable1ByDBFile,
    QueryDataDetailTable2ByDBFile,
    QueryDataDetailTable3ByDBFile,
    QueryDataDetailAttachment2ByDBFile
  } from '@wailsjs/go';

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

  // 新的冲突数据结构
  interface Condition {
    credit_code?: string; // 统一信用代码（表1、表2）
    stat_date?: string; // 年份（表1、表2、附件2）
    project_code?: string; // 项目代码（表3）
    document_number?: string; // 审查意见文号（表3）
    province_name?: string; // 省（附件2）
    city_name?: string; // 市（附件2）
    country_name?: string; // 县（附件2）
  }

  interface ConflictData {
    filePath: string;
    tableType: string;
    conditions: Condition[];
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
    [key: string]: any;
  }

  interface ConfirmModal {
    visible: boolean;
    title: string;
    tableType: string;
    tableInfoList: Array<Record<string, any>>;
    handleOk: () => void;
  }

  const props = defineProps<{
    conflictList: ConflictDetail[]; // 冲突列表
    dbFileNames: string[]; // 数据库文件名数组
    tableType: TableType;
  }>();

  const emit = defineEmits<{
    selectionChange: [selectedData: ConflictData[]];
  }>();

  const conflictData = ref<ConflictRecord[]>([]);

  const confirmModal = ref<ConfirmModal>({
    visible: false,
    title: '',
    tableType: '',
    tableInfoList: [],
    handleOk: () => {
      confirmModal.value.visible = false;
    }
  });

  // 计算有冲突的文件名
  const conflictFileNames = computed(() => {
    const fileNamesWithConflict = new Set<string>();
    
    // 遍历所有冲突记录，收集有冲突的文件名
    props.conflictList.forEach(conflict => {
      if (conflict.conflict && conflict.conflict.length > 0) {
        conflict.conflict.forEach(conflictSource => {
          fileNamesWithConflict.add(conflictSource.fileName);
        });
      }
    });
    
    // 返回有冲突的文件名数组，保持原始顺序
    return props.dbFileNames.filter(fileName => fileNamesWithConflict.has(fileName));
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
            width: '150px'
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
            width: '125px',
            ellipsis: true
          },
          {
            title: '市',
            dataIndex: 'cityName',
            key: 'cityName',
            width: '125px',
            ellipsis: true
          },
          {
            title: '县',
            dataIndex: 'countryName',
            key: 'countryName',
            width: '125px',
            ellipsis: true
          },
          {
            title: '年份',
            dataIndex: 'statDate',
            key: 'statDate',
            width: '125px'
          }
        ];
        break;
    }

    // 只添加有冲突的文件列
    conflictFileNames.value.forEach((fileName, index) => {
      baseColumns.push({
        title: fileName,
        key: `db${index}`,
        width: '300px'
      });
    });

    return baseColumns;
  });

  const hasAnyConflict = computed(() => {
    return conflictData.value.some(item => item.hasConflict);
  });

  const allSelectedStates = ref<boolean[]>([]);

  // 创建文件名到列索引的映射
  const fileNameToColumnIndex = computed(() => {
    const mapping = new Map<string, number>();
    conflictFileNames.value.forEach((fileName, index) => {
      mapping.set(fileName, index);
    });
    return mapping;
  });

  // 初始化全选状态
  watch(
    () => conflictFileNames.value,
    newConflictFileNames => {
      if (newConflictFileNames && newConflictFileNames.length > 0) {
        allSelectedStates.value = newConflictFileNames.map((_, index) => index === 0); // 默认只选中第一个
      } else {
        allSelectedStates.value = [];
      }
    },
    { immediate: true }
  );

  // 更新全选状态的computed
  const updateAllSelectedStates = () => {
    conflictFileNames.value.forEach((fileName, index) => {
      // 只考虑有冲突且有数据的行
      const conflictItemsWithData = conflictData.value.filter(item => 
        item.hasConflict && item[fileName]
      );
      if (conflictItemsWithData.length > 0) {
        // 检查当前列是否在所有有数据的冲突记录中都被选中
        allSelectedStates.value[index] = conflictItemsWithData.every(item => item.selections[`db${index}`]);
      } else {
        allSelectedStates.value[index] = false;
      }
    });
  };

  const indeterminateStates = computed(() => {
    return conflictFileNames.value.map((fileName, index) => {
      // 只考虑有冲突且有数据的行
      const conflictItemsWithData = conflictData.value.filter(item => 
        item.hasConflict && item[fileName]
      );
      const selectedCount = conflictItemsWithData.filter(item => item.selections[`db${index}`]).length;
      // 由于每条记录只能选择一个文件，所以indeterminate状态应该始终为false
      return false;
    });
  });

  // 从列key中获取数据库索引
  const getDbIndex = (columnKey: string): number => {
    const index = parseInt(columnKey.replace('db', ''));
    // 确保索引在有效范围内
    if (index >= 0 && index < conflictFileNames.value.length) {
      return index;
    }
    return 0; // 默认返回0
  };

  // 安全地获取全选状态
  const getAllSelectedState = (columnKey: string): boolean => {
    const index = getDbIndex(columnKey);
    return allSelectedStates.value[index] || false;
  };

  // 安全地获取indeterminate状态
  const getIndeterminateState = (columnKey: string): boolean => {
    const index = getDbIndex(columnKey);
    return indeterminateStates.value[index] || false;
  };

  const processConflictData = (conflicts: ConflictDetail[]) => {
    const processedData: ConflictRecord[] = [];

    // 首先收集所有有冲突的文件名
    const fileNamesWithConflict = new Set<string>();
    conflicts.forEach(conflict => {
      if (conflict.conflict && conflict.conflict.length > 0) {
        conflict.conflict.forEach(conflictSource => {
          fileNamesWithConflict.add(conflictSource.fileName);
        });
      }
    });

    // 获取有冲突的文件名数组，保持原始顺序
    const conflictFileNamesList = props.dbFileNames.filter(fileName => fileNamesWithConflict.has(fileName));

    conflicts.forEach((conflict, index) => {
      if (conflict.conflict && conflict.conflict.length > 0) {
        // 初始化选择状态，确保每一行都有一列被选中
        const selections: Record<string, boolean> = {};

        // 找到第一个有数据的列并选中
        let selectedColumnIndex = -1;
        for (let i = 0; i < conflictFileNamesList.length; i++) {
          const fileName = conflictFileNamesList[i];
          const hasData = conflict.conflict.some(source => source.fileName === fileName);
          if (hasData) {
            selections[`db${i}`] = true;
            selectedColumnIndex = i;
            break;
          }
        }

        const key = `conflict_${index}`;
        // 根据表类型设置不同的字段
        const record: ConflictRecord = {
          key: `conflict_${index}`,
          hasConflict: true,
          selections,
          conflictDetail: conflict
        };

        // 只为有冲突的文件设置数据
        for (const dbFileName of conflictFileNamesList) {
          record[dbFileName] = conflict.conflict.find(source => source.fileName === dbFileName);
        }

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

  watch(
    () => props.conflictList,
    newConflicts => {
      conflictData.value = processConflictData(newConflicts);
      updateAllSelectedStates();
    },
    { immediate: true, deep: true }
  );

  const handleSelectAllDb = (dbIndex: number, checked: boolean) => {
    const dbKey = `db${dbIndex}`;
    const fileName = conflictFileNames.value[dbIndex];

    if (checked) {
      // 如果选中当前列，只处理当前列有数据的行
      conflictData.value.forEach(item => {
        if (item.hasConflict) {
          // 只有当该行在当前列有数据时才处理
          if (item[fileName]) {
            // 先取消该行的所有选择
            Object.keys(item.selections).forEach(key => {
              item.selections[key] = false;
            });
            // 然后选中当前列
            item.selections[dbKey] = true;
          }
          // 如果该行在当前列没有数据，不做任何处理，保持原有选择状态
        }
      });

      // 检查当前列是否所有有数据的行都处于选中状态
      const allRowsInCurrentColumn = conflictData.value.filter(item => 
        item.hasConflict && item[fileName]
      );
      const allRowsSelected = allRowsInCurrentColumn.length > 0 && 
        allRowsInCurrentColumn.every(item => item.selections[dbKey]);

      // 如果当前列全选成功，检查是否需要取消其他列的全选
      if (allRowsSelected) {
        // 获取当前列有数据的行索引
        const currentColumnRowIndices = new Set(
          conflictData.value
            .map((item, index) => item.hasConflict && item[fileName] ? index : -1)
            .filter(index => index !== -1)
        );

        conflictFileNames.value.forEach((otherFileName, otherIndex) => {
          if (otherIndex !== dbIndex) {
            // 获取其他列有数据的行索引
            const otherColumnRowIndices = new Set(
              conflictData.value
                .map((item, index) => item.hasConflict && item[otherFileName] ? index : -1)
                .filter(index => index !== -1)
            );

            // 检查是否有重叠的行
            const hasOverlap = [...currentColumnRowIndices].some(index => 
              otherColumnRowIndices.has(index)
            );

            // 如果有重叠的行，则取消其他列的全选状态
            if (hasOverlap) {
              allSelectedStates.value[otherIndex] = false;
            }
          }
        });
      }
    } else {
      // 如果用户试图取消选中当前列，则重新选中当前列
      conflictData.value.forEach(item => {
        if (item.hasConflict && item[fileName]) {
          // 重新选中当前列（只对有数据的行）
          item.selections[dbKey] = true;
        }
      });
      
      // 保持当前列的全选状态为true
      allSelectedStates.value[dbIndex] = true;
      emitSelectionChange();
      return;
    }

    allSelectedStates.value[dbIndex] = checked;
    emitSelectionChange();
  };

  const handleDbSelect = (record: ConflictRecord, dbKey: string) => {
    const dbIndex = parseInt(dbKey.replace('db', ''));
    const fileName = conflictFileNames.value[dbIndex];
    
    // 如果该行在当前列没有数据，不允许选中
    if (!record[fileName]) {
      return;
    }

    // 如果当前复选框被选中，则取消其他所有复选框的选中状态
    if (record.selections[dbKey]) {
      Object.keys(record.selections).forEach(key => {
        if (key !== dbKey) {
          record.selections[key] = false;
        }
      });
    } else {
      // 如果用户试图取消选中，则重新选中当前复选框
      record.selections[dbKey] = true;
    }

    // 更新全选状态
    updateAllSelectedStates();
    emitSelectionChange();
  };


  const handleViewDetailData = async (record: ConflictRecord, dbKey: string) => {
    const dbIndex = parseInt(dbKey.replace('db', ''));
    const fileName = conflictFileNames.value[dbIndex];

    // 根据文件名查找对应的冲突源信息
    const conflictSource = record.conflictDetail.conflict.find(source => source.fileName === fileName);

    if (conflictSource) {
      try {
        let tableInfoList: Array<Record<string, any>> = [];

        // 根据表类型调用不同的API获取详细数据
        switch (props.tableType) {
          case TableType.table1:
            const resDetail = await QueryDataDetailTable1ByDBFile(conflictSource.obj_ids, conflictSource.filePath);
            if (resDetail.data) {
              // 表1返回的是数组，需要处理每个元素
              const allData = resDetail.data as Array<Record<string, any>>;
              if (allData.length > 0) {
                // 取第一个数据作为示例（因为表1有复杂的嵌套结构）
                const firstData = allData[0];
                const { main, usage, equip } = firstData;
                tableInfoList = [[main], usage, equip] as Array<Record<string, any>>;
              }
            }
            break;
          case TableType.table2:
            const table2Data = await QueryDataDetailTable2ByDBFile(conflictSource.obj_ids, conflictSource.filePath);
            console.log('表2详细数据:', table2Data);
            if (table2Data.data) {
              // 表2返回的是数组
              tableInfoList = table2Data.data as Array<Record<string, any>>;
              console.log('表2处理后的数据:', tableInfoList);
            }
            break;
          case TableType.table3:
            const table3Data = await QueryDataDetailTable3ByDBFile(conflictSource.obj_ids, conflictSource.filePath);
            if (table3Data.data) {
              // 表3返回的是数组
              tableInfoList = table3Data.data as Array<Record<string, any>>;
            }
            break;
          case TableType.attachment2:
            const attachment2Data = await QueryDataDetailAttachment2ByDBFile(conflictSource.obj_ids, conflictSource.filePath);
            if (attachment2Data.data) {
              // 附件2返回的是数组
              tableInfoList = attachment2Data.data as Array<Record<string, any>>;
            }
            break;
        }

        confirmModal.value = {
          visible: true,
          title: '基本信息',
          tableType: props.tableType,
          tableInfoList: tableInfoList,
          handleOk: () => {
            confirmModal.value.visible = false;
          }
        };
      } catch (error) {
        console.error('获取详细数据失败:', error);
        message.error('获取详细数据失败，请重试');
      }
    } else {
      message.warning('暂无数据可查看');
    }
  };

  const emitSelectionChange = () => {
    // 构建符合新的 MergeConflictData 函数参数要求的数据结构
    const selectedConflictData: ConflictData[] = [];

    // 按文件路径分组收集所有选中的冲突数据
    const fileDataMap = new Map<string, Condition[]>();

    // 遍历所有冲突行，获取每行选中的数据
    conflictData.value.forEach(item => {
      if (item.hasConflict) {
        // 找到该行被选中的列
        const selectedColumnKey = Object.keys(item.selections).find(key => item.selections[key]);
        if (selectedColumnKey) {
          const dbIndex = parseInt(selectedColumnKey.replace('db', ''));
          const fileName = conflictFileNames.value[dbIndex];
          
          // 确保该行在该列有数据（不是"无"）
          if (item[fileName]) {
            const conflictSource = item.conflictDetail.conflict.find(source => source.fileName === fileName);
            if (conflictSource) {
              // 创建条件对象
              const condition: Condition = {};
              switch (props.tableType) {
                case TableType.table1:
                case TableType.table2:
                  condition.credit_code = item.creditCode;
                  condition.stat_date = item.statDate;
                  break;
                case TableType.table3:
                  condition.project_code = item.projectCode;
                  condition.document_number = item.reviewNumber;
                  break;
                case TableType.attachment2:
                  condition.province_name = item.provinceName;
                  condition.city_name = item.cityName;
                  condition.country_name = item.countryName;
                  condition.stat_date = item.statDate;
                  break;
              }

              // 按文件路径分组
              if (!fileDataMap.has(conflictSource.filePath)) {
                fileDataMap.set(conflictSource.filePath, []);
              }
              fileDataMap.get(conflictSource.filePath)!.push(condition);
            }
          }
        }
      }
    });

    // 构建最终结果：每种tableType最多只有一个ConflictData
    fileDataMap.forEach((conditions, filePath) => {
      selectedConflictData.push({
        filePath: filePath,
        tableType: props.tableType,
        conditions: conditions
      });
    });

    emit('selectionChange', selectedConflictData);
  };

  // 暴露方法给父组件
  defineExpose({
    getSelectedData: (): ConflictData[] => {
      // 构建符合新的 MergeConflictData 函数参数要求的数据结构
      const selectedConflictData: ConflictData[] = [];

      // 按文件路径分组收集所有选中的冲突数据
      const fileDataMap = new Map<string, Condition[]>();

      // 遍历所有冲突行，获取每行选中的数据
      conflictData.value.forEach(item => {
        if (item.hasConflict) {
          // 找到该行被选中的列
          const selectedColumnKey = Object.keys(item.selections).find(key => item.selections[key]);
          if (selectedColumnKey) {
            const dbIndex = parseInt(selectedColumnKey.replace('db', ''));
            const fileName = conflictFileNames.value[dbIndex];
            
            // 确保该行在该列有数据（不是"无"）
            if (item[fileName]) {
              const conflictSource = item.conflictDetail.conflict.find(source => source.fileName === fileName);
              if (conflictSource) {
                // 创建条件对象
                const condition: Condition = {};
                switch (props.tableType) {
                  case TableType.table1:
                  case TableType.table2:
                    condition.credit_code = item.creditCode;
                    condition.stat_date = item.statDate;
                    break;
                  case TableType.table3:
                    condition.project_code = item.projectCode;
                    condition.document_number = item.reviewNumber;
                    break;
                  case TableType.attachment2:
                    condition.province_name = item.provinceName;
                    condition.city_name = item.cityName;
                    condition.country_name = item.countryName;
                    condition.stat_date = item.statDate;
                    break;
                }

                // 按文件路径分组
                if (!fileDataMap.has(conflictSource.filePath)) {
                  fileDataMap.set(conflictSource.filePath, []);
                }
                fileDataMap.get(conflictSource.filePath)!.push(condition);
              }
            }
          }
        }
      });

      // 构建最终结果：每种tableType最多只有一个ConflictData
      fileDataMap.forEach((conditions, filePath) => {
        selectedConflictData.push({
          filePath: filePath,
          tableType: props.tableType,
          conditions: conditions
        });
      });

      return selectedConflictData;
    },
    clearSelection: () => {
      // 不能完全清空选择，每行必须至少有一个选择
      // 这里重置为默认选择第一个有数据的数据库
      conflictData.value.forEach(item => {
        Object.keys(item.selections).forEach(key => {
          item.selections[key] = false;
        });
        
        // 找到第一个有数据的列并选中
        for (let i = 0; i < conflictFileNames.value.length; i++) {
          const fileName = conflictFileNames.value[i];
          if (item[fileName]) {
            item.selections[`db${i}`] = true;
            break;
          }
        }
      });
      
      // 重置全选状态，只选中第一个
      allSelectedStates.value = allSelectedStates.value.map((_, index) => index === 0);
      emitSelectionChange();
    }
  });
</script>

<style scoped lang="less"></style>
