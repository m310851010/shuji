<template>
  <div class="desk-window">
    <slot></slot>
  </div>
</template>

<script lang="ts">
  import Titlebar from './Titlebar.vue';
  export default {
    name: 'Window',
    // 默认不显示标题栏
    props: { showTitlebar: { type: Boolean, default: false } },
    setup(props, context) {
      return () => {
        const children = context.slots.default?.() || [];
        const nodes = props.showTitlebar ? children : children.filter(child => child.type !== Titlebar);
        return h(
          'div',
          {
            class: 'desk-window'
          },
          nodes
        );
      };
    }
  };
</script>

<style scoped lang="less">
  .desk-window {
    --webkit-user-select: none;
    user-select: none;
    cursor: default;
    display: flex;
    flex-direction: column;
    padding: 0;
    box-sizing: border-box;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    position: fixed;
    overflow: visible !important;
    -webkit-box-orient: vertical;
    -webkit-box-direction: normal;
    -ms-flex-direction: column;
    color: #000;
    -webkit-tap-highlight-color: rgba(0, 0, 0, 0);
  }
</style>
