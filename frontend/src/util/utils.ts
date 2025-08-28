import { TableColumnType } from 'ant-design-vue';

/**
 * 新建列
 * @param columns
 */
export function newColumns<T = any>(...columns: TableColType[]) {
  const newColumns: TableColumnType<T>[] = [];
  for (const column of columns) {
    if (column.dataIndex || column.customRender || column.key || column.title) {
      const key = column.key || (column.dataIndex as string);
      column.key = key;
      column.dataIndex = key;
      column.ellipsis = true;
      column.align = 'center';
      newColumns.push(column);
      continue;
    }

    for (const key in column) {
      newColumns.push({
        title: (column as Record<string, string>)[key],
        dataIndex: key,
        key: key,
        align: 'center',
        ellipsis: true
      });
    }
  }
  return newColumns;
}
export type TableColType = Record<string, string> | TableColumnType;

/**
 * 生成uuid
 * @returns
 */
export function UUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    const r = (Math.random() * 16) | 0;
    const v = c == 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

/**
 * omit
 * @param obj
 * @param fields
 * @returns
 */
export function omit<T>(obj: T, fields: (keyof T)[]) {
  const shallowCopy = Object.assign({}, obj);
  for (let i = 0; i < fields.length; i += 1) {
    const key = fields[i];
    delete shallowCopy[key];
  }
  return shallowCopy;
}

/**
 * 获取文件名
 * @param filePath 文件路径
 * @returns 文件名
 */
export function getFileName(filePath: string) {
  //  兼容windows和linux
  filePath = filePath.replace(/\\/g, '/');
  return filePath.split('/').pop() || '';
}

/**
 * 获取文件扩展名
 * @param filePath
 */
export function getFileExtension(filePath: string) {
  return filePath.substring(filePath.lastIndexOf('.') + 1).toLowerCase();
}
