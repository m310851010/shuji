/**
 * 阻止全局拖拽默认行为
 */
export function userGlobalDragAndDrop() {
  console.log('userGlobalDragAndDrop', 'ddd');
  const dragoverListener = (e: DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const addEvent = () => {
    document.addEventListener('dragover', dragoverListener);
    document.addEventListener('drop', dragoverListener);
  };

  const removeEvent = () => {
    document.removeEventListener('dragover', dragoverListener);
    document.removeEventListener('drop', dragoverListener);
  };

  return {
    addEvent,
    removeEvent
  };
}
