import { ref, Ref, onMounted, onUnmounted, watch } from 'vue';
import { fromEvent, Subject, merge, finalize } from 'rxjs';
import { map, switchMap, takeUntil } from 'rxjs/operators';

// 定义事件类型
interface DragEvent {
  delta: { x: number; y: number };
}

/**
 * 拖拽指令
 * @param elementRef 元素引用
 * @returns 拖拽指令
 */
export function useDraggable(elementRef: Ref<HTMLElement | null>) {
  const x = ref(0);
  const y = ref(0);
  const isDragging = ref(false);
  const destroy$ = new Subject<void>();

  onMounted(() => {
    if (!elementRef.value) return;
    const element = elementRef.value;
    const mouseDown$ = fromEvent<MouseEvent>(element, 'mousedown');
    const mouseMove$ = fromEvent<MouseEvent>(document, 'mousemove');
    const mouseUp$ = fromEvent<MouseEvent>(document, 'mouseup');

    const subscription = mouseDown$
      .pipe(
        takeUntil(destroy$), // 组件卸载时取消
        switchMap(startEvent => {
          // 阻止默认行为 (如文本选择)
          startEvent.preventDefault();
          isDragging.value = true;

          const startX = startEvent.clientX;
          const startY = startEvent.clientY;

          return mouseMove$.pipe(
            map(moveEvent => {
              const deltaX = moveEvent.clientX - startX;
              const deltaY = moveEvent.clientY - startY;
              return { x: deltaX, y: deltaY };
            }),
            // 在 mouseup 或 destroy$ 发生时结束本次拖拽流
            takeUntil(merge(mouseUp$, destroy$)),
            finalize(() => {
              isDragging.value = false;
            })
          );
        })
      )
      .subscribe(delta => {
        // 更新响应式数据
        x.value = delta.x;
        y.value = delta.y;
      });
    // 清理订阅
    onUnmounted(() => {
      subscription.unsubscribe();
      destroy$.next();
      destroy$.complete();
    });
  });

  return { x, y, isDragging };
}
