import { onMounted } from 'vue';
import { fromEvent, Subscription } from 'rxjs';

export function useTableHeight(tableBoxRef: Ref<any>, offset = 40) {
  const tableScroll = ref<{ y: number }>();
  let resize$: Subscription;
  onMounted(() => {
    tableScroll.value = { y: tableBoxRef.value.clientHeight - offset };
    resize$ = fromEvent(window, 'resize').subscribe(() => {
      tableScroll.value = { y: tableBoxRef.value.clientHeight - offset };
    });
  });

  onUnmounted(() => {
    resize$.unsubscribe();
  });

  return tableScroll;
}
