<template>
  <Window>
    <View class="container">
      <a-form
        style="width: 350px"
        :model="formState"
        name="basic"
        :label-col="{ span: 3 }"
        :wrapper-col="{ span: 24 }"
        autocomplete="off"
        @finish="onFinish"
      >
        <a-form-item class="text-tip">请选择区域后校验数据！</a-form-item>

        <a-form-item label="省" name="province" :rules="[{ required: true, message: '请选择省！' }]">
          <a-select
            v-model:value="formState.province"
            show-search
            allow-clear
            placeholder="请选择省"
            :options="provinceOptions"
            :filter-option="filterOption"
            @change="handleProvinceChange"
          ></a-select>
        </a-form-item>
        <a-form-item label="市" name="city" :rules="[{ required: true, message: '请选择市！' }]">
          <a-select
            v-model:value="formState.city"
            show-search
            allow-clear
            placeholder="请选择市"
            :options="cityOptions"
            :filter-option="filterOption"
            @change="handleCityChange"
          ></a-select>
        </a-form-item>

        <a-form-item label="县" name="district">
          <a-select
            v-model:value="formState.district"
            show-search
            allow-clear
            placeholder="请选择县"
            :options="districtOptions"
            :filter-option="filterOption"
          ></a-select>
        </a-form-item>

        <a-form-item class="text-center">
          <a-button type="primary" class="padding-horizontal" ghost html-type="submit">确认</a-button>
        </a-form-item>
      </a-form>
    </View>
  </Window>
</template>

<script setup lang="ts">
  import { useRouter } from 'vue-router';
  import { reactive, ref } from 'vue';
  import type { SelectProps } from 'ant-design-vue';
  import View from '@/components/View.vue';
  import { GetChinaAreaStr } from '@wailsjs/go';
  import { SaveAreaConfig } from '@wailsjs/go';
  const router = useRouter();

  interface FormState {
    province: string | null;
    city: string | null;
    district: string | null;
    firstLogin: boolean;
  }

  const formState = reactive<FormState>({
    province: null,
    city: null,
    district: null,
    firstLogin: true
  });

  let LOCATION_DATA: any[] = [];
  const provinceOptions = ref<SelectProps['options']>([] );

  const cityOptions = ref<SelectProps['options']>([]);
  const districtOptions = ref<SelectProps['options']>([]);

  let selectedProvince: any | null = null;
  let selectedCity: any | null = null;

  onMounted(async () => {
    const res = await GetChinaAreaStr();
    LOCATION_DATA = JSON.parse(res.data)

    provinceOptions.value = LOCATION_DATA.map(item => ({
      value: item.code,
      label: item.name
    }))
  });

  const handleProvinceChange = (value: string) => {
    selectedCity = null;
    cityOptions.value = [];
    districtOptions.value = [];
    formState.city = null;
    formState.district = null;

    if (!value) {
      selectedProvince = null;
      return;
    }

    selectedProvince = LOCATION_DATA.find(item => item.code === value)!;
    cityOptions.value = selectedProvince.children.map((item: any) => ({
      value: item.code,
      label: item.name
    }));
    districtOptions.value = [];
  };

  const handleCityChange = (value: string) => {
    if (!value) {
      selectedCity = null;
      districtOptions.value = [];
      formState.district = null;
      return;
    }
    selectedCity = selectedProvince!.children.find((item: any) => item.code === value)!;
    if (!selectedCity.children) {
      districtOptions.value = [];
      return;
    }
    districtOptions.value = selectedCity.children.map((item: any) => ({
      value: item.code,
      label: item.name
    }));
  };

  const filterOption = (input: string, option: any) => {
    return option.label.indexOf(input) >= 0;
  };

  const onFinish = () => {
    // 获取区域名称
    const provinceName = selectedProvince?.name || '';
    const cityName = selectedCity?.name || '';
    let districtName = '';
    if (selectedCity.children) {
      districtName = selectedCity.children.find((item: any) => item.code === formState.district)?.name || '';
    }

    SaveAreaConfig({
      province_name: provinceName,
      city_name: cityName,
      country_name: districtName
    });
    router.push('/main');
  };
</script>

<style scoped>
  .text-tip :deep(.ant-form-item-control-input-content) {
    font-size: 18px;
    text-align: center;
  }
</style>
