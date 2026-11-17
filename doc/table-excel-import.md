# 本地 Excel / CSV 导入

```bash
# 与表格管理页“导入表格 → 导入 Excel”一致：解析表头、新建表格、导入
lingtong-cli table import --app-id 165 --file ./订单.xlsx --name 订单

# 导入已有表格：请使用该表格导入模板的列名
lingtong-cli table data import --basic-data-id 1568 --file ./订单.xlsx
# 等价入口
lingtong-cli table import --basic-data-id 1568 --file ./订单.xlsx

# 本地校验和预览；不要求登录，不发网络请求
lingtong-cli table import --app-id 165 --file ./订单.csv --dry-run

# 大文件可增加时限（包含上传、建表、导入、等待）
lingtong-cli table import --basic-data-id 1568 --file ./订单.xls --timeout 20m

# 仅提交，稍后用同一 UUID 查询或继续等待
lingtong-cli table import --basic-data-id 1568 --file ./订单.xlsx --wait=false
lingtong-cli table import-progress --uuid <uuid> --wait=false
lingtong-cli table import-progress --uuid <uuid> --timeout 20m
```

支持 `.xlsx`、`.xls`、`.csv`（扩展名不区分大小写），文件必须非空且不超过 50 MB。`--app-id` 和 `--basic-data-id` 二选一；新表名默认是去除扩展名的文件名，`--name` 仅用于新建表格。新表使用手工维护、常规资料模式，高性能模式和连接器同步均关闭，与页面入口一致。字段由服务端推断，CLI 保留返回的字段 ID、类型和其他属性。

默认 `--wait=true`、`--timeout=10m`、`--poll-interval=2s`。进度写到 stderr，最终结果写到 stdout，可结合 `--envelope`、`--jq`。等待成功的结果例如：

```json
{
  "success": true,
  "result": {
    "basicDataId": 1568,
    "created": false,
    "uuid": "任务 UUID",
    "completed": true,
    "status": 2,
    "current": 100,
    "total": 100
  }
}
```

`--wait=false` 只保证提交成功，返回 `completed:false` 和 UUID，不把提交成功算作导入完成。进度状态：`0` 校验、`1` 导入中、`2` 完成、`3` 失败。初次查询可能返回 `result:null`，命令会继续等到状态出现或超时。

导入失败返回 API 错误（退出码 1）；本地参数/文件错误为 2，未登录为 3，等待超时或网络问题为 4。失败输出包括任务 UUID（如果已获得）和已创建表格 ID。超时或 Ctrl+C 只停止客户端等待，服务端任务仍可能继续。使用 `table import-progress` 查询已有 UUID，不重复提交。新表创建后的部分失败会保留表格，需要先检查 Schema 和数据；已有表格按后端主键规则新增或更新，不保证只追加数据。

## 接口兼容依据

实现对齐 `lingtong-app-view` 的 `excelImportWhole.vue`、`dataInfo.vue`、`excelImport.vue`、`excelImportResult.vue`、`uploadOss.js` 和 `oss-client.js`。后端依据 `lingtong-app` development 的 `BasicDataFormRecordController`、`UniversalProxyHandler`、`HttpMethodRequestWrapper`、`BasicDataImportServiceImpl`。

1. 使用已有 Bearer 认证代理取得 `/storage/oss/getStsToken`。兼容字符串 JSON、`result/data`、`credentials/Credentials` 包装及字段大小写；拒绝不完整或即将过期的临时凭据。
2. 流式 PUT 至 `erp-tmp-zb.oss-cn-zhangjiakou.aliyuncs.com/report_temporary/` 下的独立对象，使用 Content-MD5 和 OSS 签名；平台 Token 不发往 OSS，OSS 错误不回显签名响应体。403 时刷新 STS 并重试相同对象，最多两次。
3. 新表先解析列，再 `/basicdata/save`、`/basicdata/schema/save`；已有表格跳过这三步。
4. POST `/basicdata/record/import` 的 `basicDataId` 和 `fileUrl` 是 Spring 请求参数，接口未限制 multipart。前端的 FormData 与 URL 参数均可绑定。当前代理 POST 包装器忽略 `proxy.params`，CLI 因此把参数放在外层代理 URL 上，由包装器保留原始请求参数。无需修改后端，也不尝试把文件路径/文件内容塞进 JSON body。
5. GET `/basicdata/record/getImportProgress` 轮询任务，所有请求受取消和总时限约束。建表、Schema 保存及导入提交不自动重试，防止重复写入。

## 验证

本地 HTTP 模拟测试覆盖代理参数绑定、特殊文件名编码、OSS 签名与文件内容、STS 刷新、表头建表顺序、已有表导入、异步提交、空进度、任务失败、超时取消和部分建表失败保留 ID。线上验证需要有效的 CLI 登录和明确的目标表格/应用。
