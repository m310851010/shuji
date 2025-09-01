<template>
  <div class="wh-100 flex-vertical">
    <div class="page-header">设置</div>
    <div class="page-content flex-main text-center">
      <a-button type="primary" size="large" style="padding-left: 30px; padding-right: 30px" @click="handleResetPassword">重置密码</a-button>
    </div>
  </div>
</template>

<script setup lang="tsx">
  import { message, Form, Input } from 'ant-design-vue';
  import { openInfoModal, openModal } from '@/components/useModal';
  import { GetAreaConfig, Login, SetUserPassword } from '@wailsjs/go';
  import { reactive, ref } from 'vue';

  interface FormState {
    oldPassword: string;
    newPassword: string;
  }
  const formState = reactive<FormState>({
    oldPassword: '',
    newPassword: ''
  });

  const formRef = ref<any>();

  const handleResetPassword = () => {
    const modalRef = openModal({
      title: '重置密码',
      content: () => (
        <>
          <Form model={formState} ref={formRef}>
            <Form.Item label="旧密码" name="oldPassword" rules={[{ required: true, message: '请输入旧密码' }]}>
              <Input.Password placeholder="请输入旧密码" v-model:value={formState.oldPassword} />
            </Form.Item>
            <Form.Item label="新密码" name="newPassword" rules={[{ required: true, message: '请输入新密码' }]}>
              <Input.Password placeholder="请输入新密码" v-model:value={formState.newPassword} />
            </Form.Item>
          </Form>
        </>
      ),
      onOk: async formData => {
        formRef.value.validate().then(async () => {
          const loginRet = await Login(formState.oldPassword);
          if (!loginRet.ok) {
            message.error('旧密码错误');
            return;
          }

          const ret = await SetUserPassword(formState.newPassword);
          if (!ret.ok) {
            message.error(ret.message);
            return;
          }

          modalRef.close();
          openInfoModal({
            title: '重置密码',
            content: '密码重置成功',
            okText: '确定并退出登录',
            onOk: async () => {
              window.location.href = '#/login';
            },
            onCancel: () => {
              window.location.href = '#/login';
            }
          });
        });
        return false;
      }
    });
  };
</script>

<style scoped></style>
