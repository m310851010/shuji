<template>
  <div class="wh-100 flex-vertical">
    <div class="page-header">
      <span class="header-title">数据文件合并</span>
    </div>
    <div class="page-content">
      <UploadComponent
        v-model="selectedFiles"
        v-on:update:model-value="handleUpdateModelValue"
        :accept="() => true"
        :validFile="['db']"
        filterName="数据文件"
        filterPattern="*.db"
        title="选择数据文件"
      >
        <div>只能选择数据文件（.db），支持批量选择(最多4个)</div>
        <div>支持一次性拖一个或多个文件Excel文件，以及整个文件夹</div>
        <div>选择文件后，点击下方按钮开始合并</div>
      </UploadComponent>

      <a-form style="margin-top: 20px" layout="inline" :model="formState" ref="formRef" autocomplete="off">
        <a-form-item class="text-tip" label="选择合并区域"></a-form-item>

        <a-form-item label="省" name="province" :rules="[{ required: true, message: '请选择省！' }]" class="form-item-cls">
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
        <a-form-item label="市" name="city" :rules="[{ required: true, message: '请选择市！' }]" class="form-item-cls">
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

        <a-form-item label="县" name="district" class="form-item-cls">
          <a-select
            v-model:value="formState.district"
            show-search
            allow-clear
            placeholder="请选择县"
            :options="districtOptions"
            :filter-option="filterOption"
          ></a-select>
        </a-form-item>
      </a-form>

      <div class="operation-area">
        <a-button type="primary" @click="handleMerge">合并</a-button>
      </div>
    </div>
  </div>

  <a-modal
    v-model:open="modal.show"
    :bodyStyle="{ paddingTop: 0 }"
    class="full-screen-modal button-middle"
    :title="modal.title"
    @cancel="modal.handleCancel"
    @ok="modal.handleOk"
    ok-text="确认数据覆盖"
  >
    <div class="wh-100 relative">
      <div class="abs" style="overflow: auto">
        <div v-for="(item, index) in modal.tableList" :key="index" class="table-section">
          <div class="table-title">
            {{ getTableTypeTitle(item.tableType) }}
          </div>
          <DBMergeCoverTable
            :ref="
              el => {
                if (el) modal.tableRefs[index] = el;
              }
            "
            :conflictList="item.conflicts"
            :dbFileNames="item.fileNames"
            :tableType="item.tableType"
          />
        </div>
      </div>
    </div>
  </a-modal>
</template>

<script setup lang="tsx">
  import { message, type SelectProps } from 'ant-design-vue';
  import UploadComponent from './components/Upload.vue';
  import DBMergeCoverTable from './components/DBMergeCoverTable.vue';
  import { reactive, ref } from 'vue';
  import { GetChinaAreaStr, MergeDatabase, MergeConflictData, OpenSaveDialog, Movefile, Removefile } from '@wailsjs/go';
  import { TableType, TableTypeName } from '../constant';
  import { main } from '@wailsjs/models';
  import dayjs from 'dayjs';

  const selectedFiles = ref<EnhancedFile[]>([]);

  //保存合并后的数据文件到指定位置
  async function saveMergeDB(targetDbPath: string) {
    message.success('合并成功, 正在保存到指定位置');
    //弹出保存文件对话框选择保存路径把目标合并的db保存到指定位置

    const areaCode = formState.district || formState.city;
    // 获取区域名称
    const cityName = selectedCity?.name || '';
    let districtName = '';
    if (selectedCity.children) {
      districtName = selectedCity.children.find((item: any) => item.code === formState.district)?.name || '';
    }
    const areaName = districtName || cityName;

    const newName = `export_${dayjs().format('YYYYMMDDHHmmss')}${areaCode}_${areaName}`;

    //弹出保存文件对话框选择保存路径把目标合并的db保存到指定位置
    const res2 = await OpenSaveDialog(
      new main.FileDialogOptions({
        title: '保存合并后的数据文件',
        defaultFilename: `${newName}.db`
      })
    );

    if (res2.canceled) {
      await Removefile(targetDbPath);
    } else {
      await Movefile(targetDbPath, res2.filePaths[0]);
    }
  }

  const modal = reactive({
    show: false,
    tableList: [] as any[],
    tableRefs: [] as any[],
    targetDbPath: '',
    title: '数据文件合并',
    showModal: async (data: any) => {
      modal.show = true;
      modal.tableList = data;
      modal.tableRefs = [];
    },
    handleCancel: async () => {
      modal.show = false;
      modal.tableList = [];
      modal.tableRefs = [];
      message.success('取消合并');
      await Removefile(modal.targetDbPath);
    },
    handleOk: async () => {
      try {
        const allSelectedConflicts: main.ConflictData[] = [];

        modal.tableRefs.forEach(tableRef => {
          if (tableRef && tableRef.getSelectedData) {
            const selectedData = tableRef.getSelectedData();
            // 合并所有表类型的冲突数据
            allSelectedConflicts.push(...selectedData);
          }
        });

        if (allSelectedConflicts.length > 0) {
          const result = await MergeConflictData(modal.targetDbPath, allSelectedConflicts);
        }

        await saveMergeDB(modal.targetDbPath);
        modal.show = false;
      } catch (error) {
        console.error('覆盖冲突数据失败:', error);
        message.error('覆盖冲突数据失败');
      }
    }
  });

  const handleMerge = () => {
    if (!selectedFiles.value.length) {
      message.error('请先选择数据文件');
      return;
    }

    if (selectedFiles.value.length < 2) {
      message.error('请先选择至少2个数据文件');
      return;
    }

    formRef.value
      .validate()
      .then(async () => {
        if (!selectedFiles.value.length) {
          return;
        }

        // 获取区域名称
        const provinceName = selectedProvince?.name || '';
        const cityName = selectedCity?.name || '';
        let districtName = '';
        if (selectedCity.children) {
          districtName = selectedCity.children.find((item: any) => item.code === formState.district)?.name || '';
        }

        // 合并数据库
        const res = await MergeDatabase(
          provinceName,
          cityName,
          districtName,
          selectedFiles.value.map(value => value.fullPath)
        );
        if (!res.ok) {
          message.error(res.message);
          return;
        }

        // 有重复数据
        if (res.data && res.data.totalConflictCount) {
          // 保存目标数据库路径
          modal.targetDbPath = res.data.targetDbPath;

          const { table1Conflicts, table2Conflicts, table3Conflicts, attachment2Conflicts } = res.data;
          const tableTypes = [TableType.table1, TableType.table2, TableType.table3, TableType.attachment2];
          const tableList = [table1Conflicts, table2Conflicts, table3Conflicts, attachment2Conflicts]
            .map((item, index) => ({
              conflicts: item.conflicts,
              fileNames: item.fileNames,
              tableType: tableTypes[index]
            }))
            .filter(item => item.conflicts?.length);

          await modal.showModal(tableList);
          return;
        }

        await saveMergeDB(res.data.targetDbPath);
      })
      .catch(() => {});
  };

  const handleUpdateModelValue = (value: EnhancedFile[]) => {
    if (value.length) {
      // 根据正则过滤掉非法文件, 文件名规则为: export_20250826152020150000_西城区.db

      const regex = /^export_\d{18,20}_[\u4e00-\u9fa5]{2,}\.db$/;
      const validFiles = value.filter(item => regex.test(item.name));
      if (validFiles.length !== value.length) {
        message.warn('请选择正确的数据文件, 文件名规则示例: export_20250826152020150000_西城区.db');
        selectedFiles.value = validFiles;
        return;
      }
    }

    if (value.length > 4) {
      message.warn('最多选择4个文件');
      selectedFiles.value = value.slice(0, 4);
    } else {
      selectedFiles.value = value;
    }
  };

  interface FormState {
    province: string | null;
    city: string | null;
    district: string | null;
  }

  const formState = reactive<FormState>({
    province: null,
    city: null,
    district: null
  });

  const formRef = ref<any>();

  let LOCATION_DATA: any[] = [];
  const provinceOptions = ref<SelectProps['options']>([]);

  const cityOptions = ref<SelectProps['options']>([]);
  const districtOptions = ref<SelectProps['options']>([]);

  let selectedProvince: any | null = null;
  let selectedCity: any | null = null;

  onMounted(async () => {
    const res = await GetChinaAreaStr();
    LOCATION_DATA = JSON.parse(res.data);

    provinceOptions.value = LOCATION_DATA.map(item => ({
      value: item.code,
      label: item.name
    }));
  });

  // 省选择
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

  // 市选择
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

  const getTableTypeTitle = (tableType: string) => {
    switch (tableType) {
      case TableType.table1:
        return TableTypeName.table1;
      case TableType.table2:
        return TableTypeName.table2;
      case TableType.table3:
        return TableTypeName.table3;
      case TableType.attachment2:
        return TableTypeName.attachment2;
      default:
        return '未知表格';
    }
  };
</script>

<style scoped>
  .form-item-cls {
    width: 200px;
  }

  .table-section {
    margin-bottom: 20px;
  }

  .table-title {
    font-size: 16px;
    font-weight: bold;
    color: #1a5284;
    margin-bottom: 10px;
    padding: 8px 12px;
    background-color: #f5f5f5;
    border-left: 4px solid #1a5284;
  }
</style>
