<template>
  <div class="wh-100 flex-vertical">
    <div class="page-header">DB合并</div>
    <div class="page-content flex-main text-center">
      <UploadComponent>
        <div>只能选择DB文件（.db），支持批量选择</div>
        <div>选择文件后，点击下方按钮开始合并</div>
      </UploadComponent>
      <div style="margin-top: 20px">
        <a-button type="primary">合并</a-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="tsx">
  import UploadComponent from './components/Upload.vue';
</script>

<style scoped></style>
