<template>
  <div class="wh-100 flex-vertical">
    <div class="page-header">
      <span class="header-title">DB合并</span>
    </div>
    <div class="page-content text-center">
      <UploadComponent>
        <div>只能选择DB文件（.db），支持批量选择</div>
        <div>选择文件后，点击下方按钮开始合并</div>
      </UploadComponent>
      <div class="operation-area">
        <a-button  type="primary">合并</a-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="tsx">
  import UploadComponent from './components/Upload.vue';
</script>

<style scoped>
</style>
