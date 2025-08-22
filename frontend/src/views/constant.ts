import type { SelectProps } from 'ant-design-vue';

/**
 * 导入的文件类型
 */
export enum TableType {
  table1 = '附表1',
  table2 = '附表2',
  table3 = '附表3',
  attachment2 = '附件2'
}

/**
 * 导入的文件类型选项
 */
export const TableOptions: SelectProps['options'] = [
  { label: '表1', value: TableType.table1 },
  { label: '表2', value: TableType.table2 },
  { label: '表3', value: TableType.table3 },
  { label: '附件2', value: TableType.attachment2 }
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
