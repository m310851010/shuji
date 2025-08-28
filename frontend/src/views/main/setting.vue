<template>
  <div class="wh-100 flex-vertical">
    <div class="page-header">设置</div>
    <div class="page-content flex-main text-center">
      <a-button type="primary" size="large" style="padding-left: 30px; padding-right: 30px" @click="handleResetPassword">重置密码</a-button>
    </div>
  </div>
</template>

<script setup lang="tsx">
  import { message } from 'ant-design-vue';
  import { openInfoModal, openModal } from '@/components/useModal';
  import { SetUserPassword } from '@wailsjs/go';
  const handleResetPassword = () => {
    openModal({
      title: '重置密码',
      content: '确定要重置密码吗？',
      onOk: async () => {
        const ret = await SetUserPassword('111111');
        if (ret.ok) {
          openInfoModal({
            title: '重置密码',
            content: '密码已重置,默认密码为111111',
            okText: '确定并退出登录',
            onOk: async () => {
              window.location.href = '#/login';
            },
            onCancel: () => {
              window.location.href = '#/login';
            }
          });
        } else {
          message.error(ret.message);
        }
      }
    });
  };
</script>

<style scoped></style>
