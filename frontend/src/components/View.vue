<template>
  <div class="desk-view" :style="{ padding: padding }">
    <slot></slot>
  </div>
</template>

<script lang="ts">
  export default {
    name: 'View',
    props: { padding: { type: String, default: '12px' } },
    setup(props, context) {
      return {};
    }
  };
</script>

<style scoped lang="less">
  .desk-view {
    position: relative;
    flex: 1;
  }
</style>
