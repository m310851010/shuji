-- 省市县区域表, 当前用户登录时, 获取当前用户所在省市县, 然后查询该表, 用户第一次登录时会设置当前省市县
CREATE TABLE "area_config" (
   "obj_id" varchar(36) NOT NULL,                       -- 主键，表：区域配置表
   "province_name" varchar(100),                        -- 单位省级名称
   "city_name" varchar(100),                            -- 单位市级名称
   "country_name" varchar(100),                         -- 单位县级名称
   PRIMARY KEY ("obj_id")
);

-- XX省（自治区、直辖市）202X年煤炭消费状况表, 对应excel文件附件2
CREATE TABLE "coal_consumption_report" (
  "obj_id" varchar(36) NOT NULL,                       -- 主键，表：XX省（自治区、直辖市）202X年煤炭消费状况表
  "stat_date" varchar(10) NOT NULL,                    -- 数据年份
	"sg_code" varchar(2),                                -- 所属省公司编码
	"unit_id" varchar(16) NOT NULL,                      -- 单位id
	"unit_name" varchar(100),                            -- 单位名称
	"unit_level" varchar(2) NOT NULL,                    -- 单位等级：01 国家 02-省 03-市 04-县
	"province_name" varchar(100),                        -- 单位省级名称
  "city_name" varchar(100),                            -- 单位市级名称
  "country_name" varchar(100),                         -- 单位县级名称
	"total_coal" varchar(100),                           -- 煤炭消费总量',，2位小数，加密
	"raw_coal" varchar(100),                             -- 原煤，2位小数，加密
	"washed_coal" varchar(100),                          -- 洗精煤，2位小数，加密
	"other_coal" varchar(100),                           -- 其他煤炭，2位小数，加密
	"power_generation" varchar(100),                     -- 火力发电'，2位小数，加密
	"heating" varchar(100),                              -- 供热，2位小数，加密
	"coal_washing" varchar(100),                         -- 煤炭洗选，2位小数，加密
	"coking" varchar(100),                               -- 炼焦，2位小数，加密
	"oil_refining" varchar(100),                         -- 炼油及煤制油，2位小数，加密
	"gas_production" varchar(100),                       -- 制气，2位小数，加密
	"industry" varchar(100),                             -- 工业，2位小数，加密
  "raw_materials" varchar(100),                        -- 用作原材料，2位小数，加密
  "other_uses" varchar(100),                           -- 其他用途，2位小数，加密
  "coke" varchar(100),                                 -- 焦炭，2位小数，加密
  "create_user" varchar(100),                          -- 上传用户
	"create_time" datetime NOT NULL,                     -- 创建时间
	"is_confirm" varchar(100),                           -- 是否已确认，0未确认，1已确认，加密
	"is_check" varchar(100),                             -- 是否已校核，0未校核，1已校核，2校核未通过，加密
  PRIMARY KEY ("obj_id")
);


-- 重点耗煤装置（设备）煤炭消耗信息表, 对应excel文件附表2
CREATE TABLE "critical_coal_equipment_consumption" (
  "obj_id" varchar(36) NOT NULL,                       -- 主键，表：重点耗煤装置（设备）煤炭消耗信息表
  "stat_date" varchar(10) NOT NULL,                    -- 数据年份
	"create_time" datetime NOT NULL,                     -- 创建时间
	"sg_code" varchar(2),                                -- 所属省公司编码
	"unit_name" varchar(100),                            -- 单位名称
	"credit_code" varchar(20) NOT NULL,                  -- 统一社会信用代码
	"trade_a" varchar(36),                               -- 行业门类
  "trade_b" varchar(36),                               -- 行业大类
  "trade_c" varchar(36),                               -- 行业中类
	"trade_d" varchar(64),                               -- 行业小类
  "province_code" varchar(10),                         -- 单位省级编码
  "province_name" varchar(100),                        -- 单位省级名称
  "city_code" varchar(10),                             -- 单位市级编码
  "city_name" varchar(100),                            -- 单位市级名称
  "country_code" varchar(10),                          -- 单位县级编码
  "country_name" varchar(100),                         -- 单位县级名称
  "unit_addr" varchar(156),                            -- 单位地址
  "coal_type" varchar(10),                             -- 耗煤类型
  "coal_no" varchar(36),                               -- 编号
  "usage_time" varchar(20),                            -- 累计使用时间
  "design_life" varchar(20),                           -- 设计年限
  "enecrgy_efficienct_bmk" varchar(20),                -- 能效对标 优于先进水平  先进水平至节能水平之间  节能水平至准入水平之间  无能效标准
  "capacity_unit" varchar(10),                         -- 容量单位
  "capacity" varchar(50),                              -- 容量
  "use_info" varchar(10),                              -- 用途
  "status" varchar(10),                                -- 状态
  "annual_coal_consumption" varchar(100),              -- 年耗煤量，2位小数，加密
  "row_no" varchar(36),                                -- 行数
  "create_user" varchar(100),                          -- 上传用户
	"is_confirm" varchar(100),                           -- 是否已确认，0未确认，1已确认，加密
	"is_check" varchar(100),                             -- 是否已校核，0未校核，1已校核，2校核未通过，加密
  PRIMARY KEY ("obj_id")
);


-- 导入excel记录表
CREATE TABLE "data_import_record" (
  "obj_id" varchar(36) NOT NULL,                       -- 主键，表：数据导入-上传记录表
  "file_name" varchar(100) NOT NULL,                   -- 上传文件名
  "file_type" varchar(100) NOT NULL,                   -- 文件类型
  "import_time" datetime NOT NULL,                     -- 上传时间
  "import_state" varchar(20) NOT NULL,                 -- 上传状态，上传成功，上传失败
  "describe" varchar(500),                             -- 说明
  "create_user" varchar(100),                          -- 上传用户
  PRIMARY KEY ("obj_id")
);

-- 规上企业煤炭消费信息-重点耗煤装置情况, 对应excel文件附表1中的 重点耗煤装置情况 表格
CREATE TABLE "enterprise_coal_consumption_equip" (
  "obj_id" varchar(36) NOT NULL,                       -- 主键，表：规上企业煤炭消费信息-重点耗煤装置情况
  "fk_id" varchar(20) NOT NULL,                        -- 关联主表的credit_code，统一社会信用代码
	"stat_date" varchar(10) NOT NULL,                    -- 数据日期
  "create_time" datetime NOT NULL,                     -- 创建时间
  "equip_type" varchar(20),                            -- 设备类型
  "equip_no" varchar(30),                              -- 设备编号
  "total_runtime" varchar(10),                         -- 累计使用时间
  "design_life" varchar(30),                           -- 设计年限
  "energy_efficiency" varchar(50),                     -- 能效水平
	"capacity_unit" varchar(10),                         -- 容量单位
  "capacity" varchar(100),                             -- 容量，2位小数，加密
  "coal_type" varchar(10),                             -- 耗煤品种
  "annual_coal_consumption" varchar(100),              -- 年耗煤量，2位小数，加密
  "row_no" varchar(36),                                -- 行数
	"is_confirm" varchar(100),                           -- 是否已确认，0未确认，1已确认，加密
	"is_check" varchar(100),                             -- 是否已校核，0未校核，1已校核，2校核未通过，加密
  PRIMARY KEY ("obj_id")
);


-- 规上企业煤炭消费信息主表, 对应excel文件附表1中的 综合能源消费情况 表格 和202X年规模以上企业煤炭消费信息表 表格
CREATE TABLE "enterprise_coal_consumption_main" (
  "obj_id" varchar(36) NOT NULL,                       -- 主键，表：规上企业煤炭消费信息主表
  "unit_name" varchar(100),                            -- 单位名称
  "stat_date" varchar(10) NOT NULL,                    -- 数据日期
  "sg_code" varchar(2),                                -- 所属省公司编码
  "tel" varchar(15),                                   -- 联系电话
  "credit_code" varchar(20) NOT NULL,                  -- 统一社会信用代码
  "create_time" datetime NOT NULL,                     -- 创建时间
  "trade_a" varchar(36),                               -- 行业门类
  "trade_b" varchar(36),                               -- 行业大类
  "trade_c" varchar(36),                               -- 行业中类
  "province_code" varchar(10),                         -- 单位省级编码
  "province_name" varchar(100),                        -- 单位省级名称
  "city_code" varchar(10),                             -- 单位市级编码
  "city_name" varchar(100),                            -- 单位市级名称
  "country_code" varchar(10),                          -- 单位县级编码
  "country_name" varchar(100),                         -- 单位县级名称
  "annual_energy_equivalent_value" varchar(100),       -- 年综合能耗当量值，2位小数，加密
  "annual_energy_equivalent_cost" varchar(100),        -- 年综合能耗等价值，2位小数，加密
  "annual_raw_material_energy" varchar(100),           -- 年原料用能消费量，2位小数，加密
  "annual_total_coal_consumption" varchar(100),        -- 年耗煤总量-实物量，2位小数，加密
  "annual_total_coal_products" varchar(100),           -- 年耗煤总量-标准量，2位小数，加密
  "annual_raw_coal" varchar(100),                      -- 年原料用煤，2位小数，加密
  "annual_raw_coal_consumption" varchar(100),          -- 年原煤消费，2位小数，加密
  "annual_clean_coal_consumption" varchar(100),        -- 年洗精煤消费，2位小数，加密
  "annual_other_coal_consumption" varchar(100),        -- 年其他煤炭消费，2位小数，加密
  "annual_coke_consumption" varchar(100),              -- 年焦炭消费，2位小数，加密
  "create_user" varchar(100),                          -- 上传用户
	"is_confirm" varchar(100),                           -- 是否已确认，0未确认，1已确认，加密
	"is_check" varchar(100),                             -- 是否已校核，0未校核，1已校核，2校核未通过，加密
  PRIMARY KEY ("obj_id")
);

-- 规上企业煤炭消费信息-主要用途情况, 对应excel文件附表1中的 煤炭消费主要用途情况 表格
CREATE TABLE "enterprise_coal_consumption_usage" (
  "obj_id" varchar(36) NOT NULL,                       -- 主键，表：规上企业煤炭消费信息-主要用途情况
  "fk_id" varchar(20) NOT NULL,                        -- 关联主表的credit_code，统一社会信用代码
	"stat_date" varchar(10) NOT NULL,                    -- 数据日期
  "create_time" datetime NOT NULL,                     -- 创建时间
  "main_usage" varchar(10),                            -- 主要用途
  "specific_usage" varchar(10),                        -- 具体用途
  "input_variety" varchar(50),                         -- 投入品种
  "input_unit" varchar(30),                            -- 投入计量单位
  "input_quantity" varchar(100),                       -- 投入量，2位小数，加密
	"output_energy_types" varchar(10),                   -- 产出能源品种品类
  "output_quantity" varchar(100),                      -- 产出量，2位小数，加密
  "measurement_unit" varchar(10),                      -- 产出计量单位
  "remarks" varchar(256),                              -- 备注
  "row_no" varchar(36),                                -- 行数
	"is_confirm" varchar(100),                           -- 是否已确认，0未确认，1已确认，加密
	"is_check" varchar(100),                             -- 是否已校核，0未校核，1已校核，2校核未通过，加密
  PRIMARY KEY ("obj_id")
);

-- 规上企业清单, 对应excel文件【终版】0811企业清单表.xlsx
CREATE TABLE "enterprise_list" (
  "obj_id" varchar(36) NOT NULL,                       -- 主键，表：规上企业清单
	"province_name" varchar(100),                        -- 单位省级名称
	"city_name" varchar(100),                            -- 单位市级名称
	"country_name" varchar(100),                         -- 单位县级名称
	"unit_name" varchar(100),                            -- 单位详细名称
	"credit_code" varchar(20) NOT NULL,                  -- 统一社会信用代码
  PRIMARY KEY ("obj_id")
);

-- 固定资产投资项目节能审查煤炭消费情况汇总表,对应excel文件附表3
CREATE TABLE "fixed_assets_investment_project" (
  "obj_id" varchar(36) NOT NULL,                       -- 主键，表：固定资产投资项目节能审查煤炭消费情况汇总表
	"stat_date" varchar(10) NOT NULL,                    -- 数据年份
	"sg_code" varchar(2),                                -- 所属省公司编码
	"project_name" varchar(100),                         -- 项目名称
	"project_code" varchar(36),                          -- 项目代码
	"construction_unit" varchar(156),                    -- 建设单位
	"main_construction_content" varchar(256),            -- 主要建设内容
	"unit_id" varchar(20),                               -- 单位编码
	"province_name" varchar(100),                        -- 单位省级名称
	"city_name" varchar(100),                            -- 单位市级名称
	"country_name" varchar(100),                         -- 单位县级名称
	"trade_a" varchar(20),                               -- 行业大类
  "trade_c" varchar(20),                               -- 行业小类
  "examination_approval_time" varchar(20),             -- 节能审查批复时间
  "scheduled_time" varchar(20),                        -- 拟投产时间
	"actual_time" varchar(20),                           -- 实际投产时间
	"examination_authority" varchar(100),                -- 节能审查机关
	"document_number" varchar(30),                       -- 审查意见文号
  "equivalent_value" varchar(100),                     -- 当量值，2位小数，加密
  "equivalent_cost" varchar(100),                      -- 等价值，2位小数，加密
  "pq_total_coal_consumption" varchar(100),            -- 煤品消费总量-实物量，2位小数，加密
  "pq_coal_consumption" varchar(100),                  -- 煤炭消费量-实物量，2位小数，加密
  "pq_coke_consumption" varchar(100),                  -- 焦炭消费量-实物量，2位小数，加密
  "pq_blue_coke_consumption" varchar(100),             -- 兰炭消费量-实物量，2位小数，加密
  "sce_total_coal_consumption" varchar(100),           -- 煤品消费总量-折标量，2位小数，加密
  "sce_coal_consumption" varchar(100),                 -- 煤炭消费量-折标量，2位小数，加密
  "sce_coke_consumption" varchar(100),                 -- 焦炭消费量-折标量，2位小数，加密
  "sce_blue_coke_consumption" varchar(100),            -- 兰炭消费量-折标量，2位小数，加密
	"is_substitution" varchar(2),                        -- 是否煤炭消费替代
	"substitution_source" varchar(36),                   -- 煤炭消费替代来源
	"substitution_quantity" varchar(100),                -- 煤炭消费替代量，2位小数，加密
	"pq_annual_coal_quantity" varchar(100),              -- 年原料用煤量-实物量，2位小数，加密
	"sce_annual_coal_quantity" varchar(100),             -- 年原料用煤量-折标量，2位小数，加密
  "create_user" varchar(100),                          -- 上传用户
	"create_time" datetime NOT NULL,                     -- 创建时间
	"is_confirm" varchar(100),                           -- 是否已确认，0未确认，1已确认，加密
	"is_check" varchar(100),                             -- 是否已校核，0未校核，1已校核，2校核未通过，加密
  PRIMARY KEY ("obj_id")
);

-- 重点装置清单, 对应excel文件【终版】0811装置清单表.xlsx
CREATE TABLE "key_equipment_list" (
  "obj_id" varchar(36) NOT NULL,                       -- 主键，表：重点装置清单
	"province_name" varchar(100),                        -- 单位省级名称
	"city_name" varchar(100),                            -- 单位市级名称
	"country_name" varchar(100),                         -- 单位县级名称
	"unit_name" varchar(100),                            -- 使用单位名称
	"credit_code" varchar(20) NOT NULL,                  -- 使用单位统一社会信用代码
	"equip_type" varchar(20),                            -- 设备类型
	"equip_model_number" varchar(20),                    -- 设备型号
  "equip_no" varchar(30),                              -- 设备编号
  PRIMARY KEY ("obj_id")
);


-- 用户和管理员密码表, 只有一条数据, 为空时说明用户未使用该软件, 管理员密码为初始化数据
CREATE TABLE "pws_info" (
  "obj_id" varchar(36) NOT NULL,                       -- 主键，表：密码表
	"admin_pws" varchar(100),                            -- 管理员密码，加密
	"user_pws" varchar(100),                             -- 用户密码，加密，默认为空，设置密码后才有密码
  PRIMARY KEY ("obj_id")
);
