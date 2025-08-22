<template>
  <div class="desk-footer">
    <slot></slot>
  </div>
</template>

<script lang="ts">
  export default {
    name: 'Footer',
    setup(props, context) {
      return {};
    }
  };
</script>

<style scoped lang="less">
  .desk-footer {
    position: relative;
    padding: 10px 16px;
    text-align: right;
    border-top: 1px solid rgba(0, 0, 0, 0.06);
  }
</style>
