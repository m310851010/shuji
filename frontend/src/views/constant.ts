import type { SelectProps } from 'ant-design-vue';

/**
 * 导入的文件类型
 */
export enum TableType {
  table1 = 'table1',
  table2 = 'table2',
  table3 = 'table3',
  attachment2 = 'attachment2'
}

/**
 * 导入的文件类型名称
 */
export const TableTypeName: Record<TableType, string>  = {
  [TableType.table1] : '规上企业',
  [TableType.table2] : '其他单位',
  [TableType.table3] : '新上项目',
  [TableType.attachment2] : '区域综合'
}


/**
 * 导入的文件类型选项
 */
export const TableOptions: SelectProps['options'] = [
  { label: TableTypeName[TableType.table1], value: TableType.table1 },
  { label: TableTypeName[TableType.table2], value: TableType.table2 },
  { label: TableTypeName[TableType.table3], value: TableType.table3 },
  { label: TableTypeName[TableType.attachment2], value: TableType.attachment2 }
];

/**
 * 清单类型
 */
export enum ManifestType {
  /**
   * 企业
   */
  enterprise = 'enterprise',
  /**
   * 设备
   */
  equipment = 'equipment'
}

/**
 * 清单类型选项
 */
export const ManifestTypeOptions: SelectProps['options'] = [
  { label: '企业', value: ManifestType.enterprise },
  { label: '装置', value: ManifestType.equipment }
];

/**
 * 校验类型
 */
export enum CheckType {
  /**
   * 模型校验
   */
  model = 'model',
  /**
   * 人工校验
   */
  manual = 'manual'
}

/**
 * 校验类型选项
 */
export const CheckTypeOptions: SelectProps['options'] = [
  { label: '模型校验', value: CheckType.model },
  { label: '人工校验', value: CheckType.manual }
];

export const EXCEL_TYPES = ['application/vnd.ms-excel', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'];
