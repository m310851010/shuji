<template>
  <a-flex justify="flex-end" style="margin-bottom: 10px">
    <a-button type="primary">校验</a-button>
  </a-flex>

  <div class="box-grey no-bg" style="height: 340px">
    <a-flex align="center" justify="space-between" :vertical="true" class="h-100">
      <div></div>
      <div style="font-size: 24px" :style="{ color: passed ? '#52c41a' : '#ff4d4f' }">数据{{ passed ? '已' : '未' }}通过模型校验</div>
      <div>
        <a-button type="primary" v-if="!passed">下载模型报告</a-button>
      </div>
    </a-flex>
  </div>
</template>

<script setup lang="tsx">
  const passed = ref(false);
</script>
