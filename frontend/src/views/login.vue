<template>
  <Window>
    <View class="container">
      <a-form class="form" :model="formState" name="basic" autocomplete="off" @finish="onFinish">
        <a-form-item class="text-tip">
          {{ formState.firstLogin ? '初次进入请设置密码！' : '请输入密码解锁！' }}
        </a-form-item>

        <a-form-item name="password" :rules="[{ required: true, message: '密码不能为空！' }]">
          <a-input-password v-model:value="formState.password" placeholder="请输入密码" />
        </a-form-item>

        <a-form-item class="text-center">
          <a-button type="primary" class="padding-horizontal" ghost html-type="submit" :disabled="!success">
            {{ formState.firstLogin ? '确认' : '解锁' }}
          </a-button>
        </a-form-item>
      </a-form>
    </View>
  </Window>
</template>

<script setup lang="ts">
  import { useRouter } from 'vue-router';
  import { reactive, ref } from 'vue';
  import {
    GetPasswordInfo,
    SetUserPassword,
    Login,
    GetAreaConfig
  } from '@wailsjs/go';
  import { main } from '@wailsjs/models';
  import { openInfoModal } from '@/components/useModal';
  interface FormState {
    password: string;
    firstLogin: boolean;
  }

  const success = ref(false);

  GetPasswordInfo().then(ret => {
    if (!ret.ok) {
      openInfoModal({ content: ret.message });
      return;
    }

    if (ret.data?.length === 1) {
      const [{ user_pws }] = ret.data;
      formState.firstLogin = !user_pws;
      success.value = true;
    } else {
      openInfoModal({ content: '数据异常，请联系管理员！' });
    }
  });

  const router = useRouter();
  const formState = reactive<FormState>({
    password: '',
    firstLogin: false
  });

  const onFinish = () => {

    if (formState.firstLogin) {
      SetUserPassword(formState.password).then(ret => {
        if (!ret.ok) {
          openInfoModal({ content: ret.message });
          return;
        }

        router.push('/select-address');
      });
      return;
    }

    // 非首次登录，验证密码
    Login(formState.password).then(async ret => {
      if (!ret.ok) {
        openInfoModal({ content: ret.message });
      } else {
        const areaResult = await GetAreaConfig()
        if (areaResult.ok && areaResult.data) {
          await router.push('/main');
        } else {
          await router.push('/select-address');
        }
      }
    });
  };
</script>

<style scoped>
  .text-tip :deep(.ant-form-item-control-input-content) {
    font-size: 18px;
    text-align: center;
  }

  .ant-btn-primary.ant-btn-background-ghost {
    width: 120px;
    background-color: #1A5284;
    color: #fff;
  }
  .ant-btn-primary.ant-btn-background-ghost:not(:disabled):hover {
    background-color: #1a5384d6;
    color: #fff;
  }


</style>
