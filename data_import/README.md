# 数据导入模块

本模块负责处理Excel文件的校验和导入功能，支持附表1、附表2、附表3和附件2四种文件类型。

## 文件结构

```
data_import/
├── common.go                    # 公共函数和常量定义
├── validate_table1.go           # 附表1校验功能
├── validate_table2.go           # 附表2校验功能
├── validate_table3.go           # 附表3校验功能
├── validate_attachment2.go      # 附件2校验功能
├── import_table1.go             # 附表1导入功能
├── import_table2.go             # 附表2导入功能
├── import_table3.go             # 附表3导入功能
├── import_attachment2.go        # 附件2导入功能
└── README.md                    # 说明文档
```

## 功能说明

### 1. 校验功能 (Validate*)

每个校验函数都包含以下步骤：
1. 读取Excel文件
2. 解析Excel文件内容
3. 执行强制校验规则
4. 检查数据库中是否已有数据
5. 返回校验结果

### 2. 导入功能 (Import*)

每个导入函数都包含以下步骤：
1. 读取Excel文件
2. 解析Excel文件内容
3. 执行强制校验规则
4. 清空原表数据
5. 导入新数据到数据库
6. 复制缓存文件
7. 记录导入日志

### 3. 公共功能 (common.go)

包含以下可复用的函数和常量：

#### 公共函数
- `insertImportRecord()` - 插入导入记录
- `checkTableHasData()` - 检查表是否有数据
- `clearTableData()` - 清空表数据
- `validateRequiredField()` - 校验单个必填字段
- `validateRequiredFields()` - 批量校验必填字段

#### 常量定义
- 文件类型常量 (FileType*)
- 导入状态常量 (ImportState*)
- 表名常量 (Table*)
- 字段显示名称映射 (*RequiredFields)

## 数据库表对应关系

| 文件类型 | 数据库表 | 说明 |
|---------|---------|------|
| 附表1 | enterprise_coal_consumption_main<br>enterprise_coal_consumption_usage<br>enterprise_coal_consumption_equip | 规上企业煤炭消费信息 |
| 附表2 | critical_coal_equipment_consumption | 重点耗煤装置（设备）煤炭消耗信息 |
| 附表3 | fixed_assets_investment_project | 固定资产投资项目节能审查煤炭消费情况 |
| 附件2 | coal_consumption_report | 煤炭消费状况表 |

## 校验规则

### 强制校验规则
根据 `files/校验提示词_V5_0821.mhtml` 文件中的强制校验规则进行校验：

1. **附表1**: 检查年份、统一社会信用代码、企业名称等必填字段
2. **附表2**: 检查统一社会信用代码、单位名称、数据年份、耗煤类型、编号等必填字段
3. **附表3**: 检查项目名称、项目所在省市区、节能审查批复文号等必填字段
4. **附件2**: 检查数据年份、单位省市区名称、单位名称等必填字段

### 待实现的校验规则
- 企业清单检查逻辑
- 企业名称和统一信用代码对应关系检查
- 数据单位检查逻辑
- 文件格式检查逻辑

## 数据加密

根据 `files/main.sql` 文件描述，以下字段需要使用SM4加密存储：

### 附表1相关表
- 各种煤炭消费量字段
- 企业名称等敏感信息

### 附表2表
- annual_coal_consumption (年耗煤量)
- 其他敏感信息字段

### 附表3表
- 各种煤炭消费量字段
- 项目名称等敏感信息

### 附件2表
- total_coal (煤炭消费总量)
- raw_coal (原煤)
- washed_coal (洗精煤)
- other_coal (其他煤炭)
- power_generation (火力发电)
- heating (供热)
- coal_washing (煤炭洗选)
- coking (炼焦)
- oil_refining (炼油及煤制油)
- gas_production (制气)
- industry (工业)
- raw_materials (用作原材料)
- other_uses (其他用途)
- coke (焦炭)
- is_confirm (是否已确认)
- is_check (是否已校核)

## 待实现功能

1. **Excel解析功能**: 所有 `parse*Excel` 方法需要根据具体的Excel文件格式实现
2. **数据导入功能**: 所有 `import*Data` 方法需要实现具体的数据库插入逻辑
3. **SM4加密**: 敏感数据的加密存储功能
4. **重复数据检查**: 检查文件是否已上传过的逻辑
5. **高级校验规则**: 企业清单检查、数据单位检查等

## 使用示例

```go
// 校验附表1文件
result := app.ValidateTable1File("path/to/table1.xlsx")
if result.Ok {
    // 校验通过，可以导入
    importResult := app.ImportTable1("path/to/table1.xlsx")
    if importResult.Ok {
        fmt.Println("导入成功")
    }
} else {
    fmt.Println("校验失败:", result.Message)
}
```

## 编译状态

✅ **编译成功** - 所有文件已成功编译，无错误
✅ **前端绑定** - 所有函数已正确暴露给前端
✅ **类型安全** - 使用正确的QueryResult类型，避免类型冲突
