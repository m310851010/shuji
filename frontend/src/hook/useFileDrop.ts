import { OnFileDrop } from '@wailsapp/runtime';
import { GetFileInfo, Readdir } from '@wailsjs/go';
import { main } from '@wailsjs/models';

// 当前文件拖放处理函数
let currentFileDropHandler: ((files: EnhancedFile[], x: number, y: number) => void)[] = [];

/**
 * 使用文件拖拽处理函数
 */
export function useFileDrop() {
  OnFileDrop(async (x, y, paths) => {
    const files: EnhancedFile[] = [];
    for (let i = 0; i < paths.length; i++) {
      const fullPath = paths[i];
      const fileInfo = await GetFileInfo(fullPath);
      if (fileInfo.isDirectory) {
        const _fileInfo: EnhancedFile = await getFilesDir(fileInfo);
        files.push(_fileInfo);
      } else {
        files.push(fileInfo as unknown as EnhancedFile);
      }
    }

    if (currentFileDropHandler.length) {
      currentFileDropHandler.forEach(fn => fn(files, x, y));
    }
  }, true);
}

/**
 * 获取文件夹信息
 * @param fileInfo 文件信息
 * @returns 文件夹信息
 */
async function getFilesDir(fileInfo: main.FileInfo): Promise<EnhancedFile> {
  const dirResult = await Readdir(fileInfo.fullPath);
  const _fileInfo = fileInfo as unknown as EnhancedFile;
  if (!dirResult.ok) {
    return _fileInfo;
  }
  const files: EnhancedFile[] = [];
  for (const filePath of dirResult.data) {
    const _f = await GetFileInfo(filePath);
    files.push(_f as unknown as EnhancedFile);
  }
  _fileInfo.files = files;
  return _fileInfo;
}

/**
 * 设置文件拖拽处理函数
 * @param handler 文件拖拽处理函数
 */
export function setFileDropHandler(...handler: ((files: EnhancedFile[], x: number, y: number) => void)[]) {
  currentFileDropHandler = [...handler];
}

/**
 * 清除文件拖拽处理函数
 */
export const clearFileDropHandler = () => {
  currentFileDropHandler.length = 0;
};
