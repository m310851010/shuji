// useModal.ts
import { createVNode, render } from 'vue';
import ModalComponent from './Modal'; // 引入我们自定义的 Modal 组件
/**
 * // 在你的组件中使用
 * import { useModal, openModal } from '@/components/useModal';
 *
 * // 方式一：使用 useModal (获取控制权)
 * const myModal = useModal({
 *   title: <span>可拖拽标题</span>,
 *   content: <div>这里是内容</div>,
 *   onOk: () => {
 *     console.log('OK clicked');
 *   },
 *   onCancel: () => {
 *     console.log('Cancel clicked');
 *   },
 * });
 *
 * // 打开
 * myModal.open();
 *
 * // 关闭
 * // myModal.close();
 *
 * // 方式二：使用 openModal (直接打开)
 * const modalRef = openModal({
 *   title: '快速打开',
 *   content: '这个 Modal 可以拖拽！',
 *   okText: '确认',
 *   cancelText: '取消',
 * });
 *
 * // 之后可以关闭
 * // modalRef.close();
 * // modalRef.destroy();
 */
import { ModalFuncProps } from 'ant-design-vue/es/modal/Modal';
import { ButtonProps } from 'ant-design-vue';
interface UseModalProps extends ModalFuncProps {
  // title?: string | JSX.Element | (() => JSX.Element);
  // content?: string | JSX.Element | (() => JSX.Element);
  buttons?: (ButtonProps & { text: string })[];
}

interface UseModalReturn {
  open: () => void;
  close: () => void;
  destroy: () => void;
  visible: boolean;
}

export function useModal(config: UseModalProps): UseModalReturn {
  let container: HTMLDivElement | null = null;
  let closeFn: (() => void) | null = null;
  let vnode: any = null;

  const visible = ref(false); // 状态

  const open = () => {
    if (container) return; // 防止重复打开

    container = document.createElement('div');
    document.body.appendChild(container);
    visible.value = true;

    vnode = createVNode(ModalComponent, {
      ...config,
      visibleState: visible,
      'onUpdate:visible': (val: boolean) => {
        visible.value = val;
        if (!val && closeFn) {
          closeFn();
        }
      }
    });

    render(vnode, container);
  };

  const close = () => {
    if (visible.value) {
      visible.value = false;
    }
  };

  const destroy = () => {
    if (container && vnode) {
      render(null, container);
      if (container.parentNode) {
        container.parentNode.removeChild(container);
      }
      container = null;
      vnode = null;
      closeFn = null;
    }
  };

  closeFn = destroy;

  return {
    open,
    close,
    destroy,
    visible: visible.value // 注意：这里返回的是值，如果需要响应式，返回 visible
  };
}

export function openModal(config: UseModalProps) {
  const modal = useModal(config);
  modal.open();
  return modal;
}

export function openInfoModal(config: UseModalProps) {
  if (!config.cancelText) {
    config.cancelText = null;
  }
  config.width ??= 450;
  return openModal(config);
}
