import { onMounted } from 'vue';
import { fromEvent, Subscription } from 'rxjs';

export function useTableHeight(tableBoxRef: Ref<any>, scrollOptions?: { x?: any; y?: any; offset?: number }): Ref<any> {
  const tableScroll = ref<{ y?: any; x?: any }>({});
  let resize$: Subscription;
  scrollOptions ??= {};
  const offset = scrollOptions.offset ?? 40;
  onMounted(() => {
    setTimeout(() => {
      tableScroll.value = { ...scrollOptions, y: tableBoxRef.value.clientHeight - offset };
      console.log(tableScroll.value);
    }, 200);
   
    resize$ = fromEvent(window, 'resize').subscribe(() => {
      tableScroll.value = { ...scrollOptions, y: tableBoxRef.value.clientHeight - offset };
    });
  });

  onUnmounted(() => {
    resize$.unsubscribe();
  });

  return tableScroll;
}
