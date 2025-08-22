// 监听重复数据通知
import { EventsEmit, EventsOff, EventsOn } from '@wailsapp/runtime';
import { UUID } from '@/util';

const messageMap = new Map<string, { onDuplicateFunc: OnDuplicateData; strategy: UserChooseType; onImportResult: OnImportResult }>();

/**
 * 监听重复数据处理
 */
export function setupDuplicateData() {
  EventsOn('exists_duplicate_data', async (data: DuplicateDataInfo) => {
    console.log('重复数据:', data);
    const handler = messageMap.get(data.messageId);
    if (handler) {
      const strategy = await handler.onDuplicateFunc(data);
      handler.strategy = strategy;
      console.log('用户选择的类型:', strategy);
      // 发送确认消息
      EventsEmit('confirm_duplicate_data', {
        messageId: data.messageId,
        strategy
      });
    }
  });

  // 监听导入结果通知
  EventsOn('import_result', (data: ImportResult) => {
    const handler = messageMap.get(data.messageId);
    console.log('导入结果:', data, handler);
    if (handler) {
      handler.onImportResult({ ...data, strategy: handler.strategy });
      messageMap.delete(data.messageId);
    }
  });
}
export function destroyDuplicateData() {
  EventsOff('exists_duplicate_data');
  EventsOff('import_result');
}

/**
 * 用户选择的类型
 */
export enum UserChooseType {
  NONE = 'NONE',
  REPLACE = 'replace',
  SKIP = 'skip',
  CANCEL = 'cancel'
}

/**
 * 重复数据处理
 */
export function useDuplicateData() {
  return {
    /**
     * 启动重复数据处理
     * @param onDuplicateData 重复数据处理
     * @param fn
     */
    start: (fn: StartCallback, onDuplicateData: OnDuplicateData) => {
      const item = {
        onDuplicateFunc: onDuplicateData,
        onImportResult: (_: ImportResult) => {
          console.log('导入结果:', _);
        },
        strategy: UserChooseType.NONE
      };
      const uuid = UUID();
      messageMap.set(uuid, item);
      onUnmounted(() => messageMap.delete(uuid));

      return new Promise<ImportResult>(async resolve => {
        item.onImportResult = (data: ImportResult) => {
          resolve(data);
        };
        const result = await fn(uuid);
        if (result === false) {
          messageMap.delete(uuid);
        }
      });
    }
  };
}

/**
 * 重复数据处理
 */
type OnDuplicateData = (data: DuplicateDataInfo) => Promise<UserChooseType>;
/**
 * 导入结果
 */
type OnImportResult = (data: ImportResult) => void;

/**
 * 启动重复数据处理
 */
type StartCallback = (messageId: string) => any | boolean | Promise<boolean>;

/**
 * 重复数据信息
 */
export interface DuplicateDataInfo {
  messageId: string;
  message: string;
  data: any;
  excelRow?: number; // Excel行号
  fileName?: string; // 文件名
}

/**
 * 导入结果
 */
export interface ImportResult {
  ok: boolean;
  message: string;
  successCount: number;
  replaceCount: number;
  skipCount: number;
  messageId: string;
  error?: string;
  /**
   * 用户选择的类型
   */
  strategy: UserChooseType;
}
