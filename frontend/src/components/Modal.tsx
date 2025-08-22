import { defineComponent, ref, PropType, CSSProperties, watch, watchEffect } from 'vue';
import { Modal as AntModal, Button, ButtonProps } from 'ant-design-vue';
import { useDraggable } from '@/hook/useDraggable';
import { initDefaultProps } from 'ant-design-vue/es/_util/props-util';
import { modalProps } from 'ant-design-vue/es/modal/Modal';
import { omit } from '@/util';

export default defineComponent({
  name: 'Modal',
  props: {
    ...initDefaultProps(modalProps(), {
      width: 520,
      confirmLoading: false,
      okType: 'primary',
      okText: '确定',
      cancelText: '取消',
      maskClosable: false,
      keyboard: false,
      mask: true,
      closable: true
    }),
    title: { type: [String, Function], required: false, default: '提示' },
    content: { type: [String, Function], required: false },
    buttons: { type: [Array, Function], required: false },
    openState: { type: Object as PropType<Ref<boolean>> }
  },

  emits: ['update:open', 'ok', 'cancel'],
  setup(props, { emit, slots, attrs }) {
    // 拖拽标题栏的 Ref
    const modalTitleRef = ref<HTMLElement>(null!);

    const { x, y, isDragging } = useDraggable(modalTitleRef);

    const startX = ref<number>(0);
    const startY = ref<number>(0);
    const transformX = ref(0);
    const transformY = ref(0);
    watch([x, y], () => {
      transformX.value = startX.value + x.value;
      transformY.value = startY.value + y.value;
    });

    watch(isDragging, newVal => {
      if (newVal) {
        startX.value = transformX.value;
        startY.value = transformY.value;
      }
    });

    const transformStyle = computed<CSSProperties>(() => {
      return {
        transform: `translate(${transformX.value}px, ${transformY.value}px)`
      };
    });

    const handleOk = async () => {
      if (props.onOk) {
        const result = props.onOk();
        if (result && typeof (result as Promise<void>).then === 'function') {
          try {
            await result;
            // 关闭对话框并同步状态
            open.value = false;
          } catch (error) {
            console.error('Modal onOk error:', error);
            return; // 出错时不关闭对话框
          }
        }
        return;
      }

      // 关闭对话框并同步状态
      open.value = false;
      emit('update:open', false);
      emit('ok');
    };

    const handleCancel = () => {
      if (props.onCancel) {
        props.onCancel();
      }
      open.value = false;
      emit('update:open', false);
      emit('cancel');
    };

    const cancelButtonProps = { ...props.cancelButtonProps };
    if (props.cancelText == null) {
      // @ts-ignore
      cancelButtonProps.style = {
        display: 'none'
      };
    }

    const okButtonProps = { ...props.okButtonProps };
    if (props.okText == null) {
      // @ts-ignore
      okButtonProps.style = {
        display: 'none'
      };
    }

    const newProps = computed(() =>
      omit(props, [
        'title',
        'open',
        'openState',
        'onUpdate:open',
        'onUpdate:visible',
        'visible',
        'onOk',
        'onCancel',
        'cancelButtonProps',
        'modalRender',
        'okButtonProps',
        'wrapClassName',
        'content',
        'footer',
        'buttons'
      ])
    );

    const open = toRef(props.openState);

    const footer = computed(() => {
      if (props.footer) {
        if (typeof props.footer === 'function') {
          return props.footer();
        }
        return props.footer;
      }

      if (props.buttons) {
        if (typeof props.buttons === 'function') {
          return props.buttons();
        }

        return props.buttons.map((button: any, i) => {
          return (
            <Button key={i} {...button}>
              {button.text}
            </Button>
          );
        });
      }
      return undefined;
    });

    return () => (
      <AntModal
        {...newProps.value}
        v-model:open={open.value}
        onOk={handleOk}
        onCancel={handleCancel}
        title={
          <div ref={modalTitleRef} style={{ width: '100%', cursor: 'move', userSelect: 'none' }}>
            {typeof props.title === 'function' ? props.title() : props.title}
          </div>
        }
        cancelButtonProps={cancelButtonProps}
        okButtonProps={okButtonProps}
        modalRender={({ originVNode }) => <div style={transformStyle.value}>{originVNode}</div>}
        wrapClassName={(props.wrapClassName || '') + ' mdc-modal-wrap'}
        footer={footer.value}
      >
        {typeof props.content === 'function' ? props.content() : props.content}
      </AntModal>
    );
  }
});
