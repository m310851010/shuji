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
        <template v-if="column.key.startsWith('db') && record[column.title]">
          <div class="db-column">
            <a-button type="link" size="small" @click="handleViewDetailData(record, column.key)">查看</a-button>
            <a-checkbox
              v-model:checked="record.selections[column.key]"
              :disabled="!record.hasConflict || isCheckboxDisabled(record, column.key)"
              @change="() => handleDbSelect(record, column.key)"
            />
          </div>
        </template>
        <template v-else>无</template>
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

    props.dbFileNames.forEach((fileName, index) => {
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

  // 初始化全选状态
  watch(
    () => props.dbFileNames,
    newDbFileNames => {
      if (newDbFileNames && newDbFileNames.length > 0) {
        allSelectedStates.value = newDbFileNames.map((_, index) => index === 0); // 默认只选中第一个
      } else {
        allSelectedStates.value = [];
      }
    },
    { immediate: true }
  );

  // 更新全选状态的computed
  const updateAllSelectedStates = () => {
    props.dbFileNames.forEach((_, index) => {
      const conflictItems = conflictData.value.filter(item => item.hasConflict);
      if (conflictItems.length > 0) {
        // 检查当前列是否在所有冲突记录中都被选中
        allSelectedStates.value[index] = conflictItems.every(item => item.selections[`db${index}`]);
      }
    });
  };

  const indeterminateStates = computed(() => {
    return props.dbFileNames.map((_, index) => {
      const conflictItems = conflictData.value.filter(item => item.hasConflict);
      const selectedCount = conflictItems.filter(item => item.selections[`db${index}`]).length;
      // 由于每条记录只能选择一个文件，所以indeterminate状态应该始终为false
      return false;
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

    conflicts.forEach((conflict, index) => {
      if (conflict.conflict && conflict.conflict.length > 0) {
        // 初始化选择状态，默认只选中第一个文件
        const selections: Record<string, boolean> = {};

        // 设置默认选中第一个数据库
        if (props.dbFileNames.length > 0) {
          selections['db0'] = true;
        }

        const key = `conflict_${index}`;
        // 根据表类型设置不同的字段
        const record: ConflictRecord = {
          key: `conflict_${index}`,
          hasConflict: true,
          selections,
          conflictDetail: conflict
        };

        for (const dbFileName of props.dbFileNames) {
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

    if (checked) {
      // 如果选中当前列，需要先取消其他列的选择
      conflictData.value.forEach(item => {
        if (item.hasConflict) {
          // 先取消所有选择
          Object.keys(item.selections).forEach(key => {
            item.selections[key] = false;
          });
          // 然后选中当前列
          item.selections[dbKey] = true;
        }
      });

      props.dbFileNames.forEach((_, index) => {
        if (index !== dbIndex) {
          allSelectedStates.value[index] = false;
        }
      });
    } else {
      // 如果取消选中当前列，取消该列的所有选择
      conflictData.value.forEach(item => {
        if (item.hasConflict) {
          item.selections[dbKey] = false;
        }
      });
    }

    allSelectedStates.value[dbIndex] = checked;
    emitSelectionChange();
  };

  const handleDbSelect = (record: ConflictRecord, dbKey: string) => {
    // 如果当前复选框被选中，则取消其他所有复选框的选中状态
    if (record.selections[dbKey]) {
      Object.keys(record.selections).forEach(key => {
        if (key !== dbKey) {
          record.selections[key] = false;
        }
      });
    }

    // 更新全选状态
    updateAllSelectedStates();
    emitSelectionChange();
  };

  // 判断复选框是否应该被禁用
  const isCheckboxDisabled = (record: ConflictRecord, dbKey: string): boolean => {
    // 如果当前复选框已选中，则不禁用
    if (record.selections[dbKey]) {
      return false;
    }

    // 检查是否有其他复选框被选中
    const hasOtherSelected = Object.keys(record.selections).some(key => key !== dbKey && record.selections[key]);

    // 如果有其他复选框被选中，则禁用当前复选框
    return hasOtherSelected;
  };

  const handleViewDetailData = async (record: ConflictRecord, dbKey: string) => {
    const dbIndex = parseInt(dbKey.replace('db', ''));
    const fileName = props.dbFileNames[dbIndex];

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

    // 按表类型分组选中的冲突数据
    props.dbFileNames.forEach((fileName, dbIndex) => {
      const dbKey = `db${dbIndex}`;
      const selectedItems = conflictData.value.filter(item => item.hasConflict && item.selections[dbKey]);

      if (selectedItems.length > 0) {
        // 收集该文件中所有选中的冲突源信息
        selectedItems.forEach(item => {
          const conflictSource = item.conflictDetail.conflict.find(source => source.fileName === fileName);
          if (conflictSource) {
            // 检查是否已经存在相同的文件路径
            const existingIndex = selectedConflictData.findIndex(existing => existing.filePath === conflictSource.filePath);
            if (existingIndex >= 0) {
              // 添加新的冲突条件
              const condition: Condition = {};

              // 根据表类型设置相应的条件字段
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

              selectedConflictData[existingIndex].conditions.push(condition);
            } else {
              // 创建新的冲突数据
              const condition: Condition = {};

              // 根据表类型设置相应的条件字段
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

              selectedConflictData.push({
                filePath: conflictSource.filePath,
                tableType: props.tableType,
                conditions: [condition]
              });
            }
          }
        });
      }
    });

    emit('selectionChange', selectedConflictData);
  };

  // 暴露方法给父组件
  defineExpose({
    getSelectedData: (): ConflictData[] => {
      // 构建符合新的 MergeConflictData 函数参数要求的数据结构
      const selectedConflictData: ConflictData[] = [];

      // 按表类型分组选中的冲突数据
      props.dbFileNames.forEach((fileName, dbIndex) => {
        const dbKey = `db${dbIndex}`;
        const selectedItems = conflictData.value.filter(item => item.hasConflict && item.selections[dbKey]);

        if (selectedItems.length > 0) {
          // 收集该文件中所有选中的冲突源信息
          selectedItems.forEach(item => {
            const conflictSource = item.conflictDetail.conflict.find(source => source.fileName === fileName);
            if (conflictSource) {
              // 检查是否已经存在相同的文件路径
              const existingIndex = selectedConflictData.findIndex(existing => existing.filePath === conflictSource.filePath);
              if (existingIndex >= 0) {
                // 添加新的冲突条件
                const condition: Condition = {};

                // 根据表类型设置相应的条件字段
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

                selectedConflictData[existingIndex].conditions.push(condition);
              } else {
                // 创建新的冲突数据
                const condition: Condition = {};

                // 根据表类型设置相应的条件字段
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

                selectedConflictData.push({
                  filePath: conflictSource.filePath,
                  tableType: props.tableType,
                  conditions: [condition]
                });
              }
            }
          });
        }
      });

      return selectedConflictData;
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

<style scoped lang="less"></style>
