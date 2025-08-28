<template>
  <div class="wh-100 flex-vertical">
    <div class="page-header">
      <span class="header-title">DB合并</span>
    </div>
    <div class="page-content text-center">
      <UploadComponent v-model="selectedFiles" v-on:update:model-value="handleUpdateModelValue" :accept="() => true" :validFile="['db']"  filterName="DB文件" filterPattern="*.db" title="选择DB文件">
        <div>只能选择DB文件（.db），支持批量选择(最多4个)</div>
        <div>支持一次性拖一个或多个文件Excel文件，以及整个文件夹</div>
        <div>选择文件后，点击下方按钮开始合并</div>
      </UploadComponent>

        <a-form
            style=" margin-top: 20px;"
            layout="inline"
            :model="formState"
            ref="formRef"
            autocomplete="off"
        >
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
        <a-button  type="primary" @click="handleMerge">合并</a-button>
      </div>
    </div>
  </div>

  <a-modal
    v-model:open="modal.show"
    :bodyStyle="{ paddingTop: 0 }"
    class="full-screen-modal button-middle"
    :title="modal.title"
    :cancel-button-props="{ style: 'display: none' }"
    @ok="modal.handleOk"
    ok-text="确认数据覆盖"
  >
    <div class="wh-100 relative">
      <div class="abs" style="overflow: auto;">
        <DBMergeCoverTable 
          v-for="(item, index) in modal.tableList" 
          :key="index"
          :ref="(el) => { if (el) modal.tableRefs[index] = el }"
          :conflictList="item.conflicts" 
          :dbFileNames="item.fileNames" 
          :tableType="item.tableType" 
        />
      </div>
    </div>
  </a-modal>
</template>

<script setup lang="tsx">
import {message, type SelectProps} from 'ant-design-vue';
import UploadComponent from './components/Upload.vue';
import DBMergeCoverTable from './components/DBMergeCoverTable.vue';
import {reactive, ref} from 'vue';
import {GetChinaAreaStr, MergeDatabase, SaveAreaConfig, MergeConflictData, OpenSaveDialog, Copyfile, Movefile, Removefile} from '@wailsjs/go';
import { openModal } from '@/components/useModal';
import { TableType } from '../constant';
import { main } from '@wailsjs/models';

  const selectedFiles = ref<EnhancedFile[]>([]);

  
  const modal = reactive({
    show: false,
    tableList: [] as any[],
    tableRefs: [] as any[],
    targetDbPath: '',
    title: 'DB合并',
    showModal: async (data: any) => {
      modal.show = true;
      modal.tableList = data;
      modal.tableRefs = [];
    },
    handleOk: async () => {
      try {
        // 收集所有选中的冲突数据
        const allSelectedConflicts: Record<string, any[]> = {};
        
        modal.tableRefs.forEach((tableRef) => {
          if (tableRef && tableRef.getSelectedData) {
            const selectedData = tableRef.getSelectedData();
            // 合并所有表类型的冲突数据
            Object.keys(selectedData).forEach(tableType => {
              if (!allSelectedConflicts[tableType]) {
                allSelectedConflicts[tableType] = [];
              }
              allSelectedConflicts[tableType].push(...selectedData[tableType]);
            });
          }
        });
        
        // 检查是否有选中的数据
        const hasSelectedData = Object.keys(allSelectedConflicts).length > 0 && 
          Object.values(allSelectedConflicts).some(conflicts => conflicts.length > 0);
        
        if (hasSelectedData) {
           // 调用后端接口处理冲突数据
            const result = await MergeConflictData(modal.targetDbPath, allSelectedConflicts);
            if (!result.ok) {
              message.error(result.message);
              Removefile(modal.targetDbPath);
              return;
            }
        }
        
        
        modal.show = false;
          //弹出保存文件对话框选择保存路径把目标合并的db保存到指定位置
        const res = await OpenSaveDialog(new main.FileDialogOptions({
          title: '保存合并后的DB文件',
          defaultFilename: modal.targetDbPath,
          defaultPath: modal.targetDbPath,
        }));

        
        if (res.canceled) {
            Removefile(modal.targetDbPath);
          } else {
            await Movefile(modal.targetDbPath, res.filePaths[0]);
          }
        
      } catch (error) {
        console.error('处理冲突数据失败:', error);
        message.error('处理冲突数据失败');
      }
    }
  });

  const handleMerge = () => {
    if (!selectedFiles.value.length) {
      message.error('请先选择DB文件');
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
         const res = await MergeDatabase(provinceName, cityName, districtName, selectedFiles.value.map(value => value.fullPath));
          console.log('MergeDatabase==', res);
          if (!res.ok) {
            message.error(res.message);
            return;
          }

          // 有重复数据
          if (res.data && res.data.totalConflictCount) {
            // 保存目标数据库路径
            modal.targetDbPath = res.data.targetDbPath;
            
            const {table1Conflicts, table2Conflicts, table3Conflicts, attachment2Conflicts} = res.data;
            const tableTypes = [TableType.table1, TableType.table2, TableType.table3, TableType.attachment2];
            const tableList = [table1Conflicts, table2Conflicts, table3Conflicts, attachment2Conflicts].map((item, index) => ({
              conflicts: item.conflicts,
              fileNames: item.fileNames,
              tableType: tableTypes[index],
            })).filter((item) => item.conflicts.length);

            modal.showModal(tableList);
          }
        }).catch(() => {});
  }

  const handleUpdateModelValue = (value: EnhancedFile[]) => {
    if (value.length) {
      // 根据正则过滤掉非法文件, 文件名规则为: export_20250826150000_xichengqu.db

      const regex = /^export_\d{14}_[a-zA-Z0-9]+\.db$/;
      const validFiles = value.filter((item) => regex.test(item.name));
      if (validFiles.length !== value.length) {
        message.warn('请选择正确的DB文件, 文件名规则为: export_20250826150000_xichengqu.db');
        selectedFiles.value = validFiles;
      }
    } else if (value.length > 4) {
      message.warn('最多选择4个文件');
      selectedFiles.value = value.slice(0, 4);
    } else {
      selectedFiles.value = value;
    }
  }


interface FormState {
  province: string | null;
  city: string | null;
  district: string | null;
}

const formState = reactive<FormState>({
  province: '河北省',
  city: '秦皇岛市',
  district: null,
});

const formRef = ref<any>();

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

</script>

<style scoped>
.form-item-cls {
  width: 200px;
}
</style>
