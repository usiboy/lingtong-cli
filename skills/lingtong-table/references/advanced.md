# 表格扩展操作

表格、视图、筛选、同步及下载接口的参数和 DTO 统一维护在 CLI 共享 OpenAPI 规范。先用 `lingtong-cli schema path /路径` 和对应 `service ... --help` 查参数；常规表格、Schema、记录及视图操作仍使用 `table ...` 专用命令。

## 筛选与展示

`/table/filterSort/get` 读取已有筛选排序配置，`/table/filterSort/save` 保存修改。分组使用视图配置，分组展开/收起不修改业务记录；字段显示、列顺序、列宽及颜色沿用 Schema 配置，保留未修改的属性。

## Excel 导入与导出

优先使用 `table import --app-id <id> --file <本地文件>` 新建表格并导入，或 `table data import --basic-data-id <id> --file <本地文件>` 导入已有表格；两者默认等待完成。支持 `.xlsx`、`.xls`、`.csv`，最大 50 MB。`--wait=false` 只提交，`table import-progress --uuid <uuid>` 继续等待。

专用命令自动完成以下步骤：

1. 通过代理 `GET /storage/oss/getStsToken` 获取临时凭据，将文件上传至前端使用的 `erp-tmp-zb` 桶、`report_temporary/` 前缀。平台 Token 不发送到 OSS，临时凭据不写入输出。
2. 新表通过 `POST /basicdata/record/import/columnsSchema` 的请求参数 `fileUrl` 解析列，再保存表格及字段 Schema；已有表格直接复用已有字段。
3. `POST /basicdata/record/import` 接收请求参数 `basicDataId`、`fileUrl`。前端用 multipart，但后端也接受 URL 参数，并未限定 multipart。CLI 将 URL 编码后的参数放在外层 `/gw/ai/proxy` URL 上，后端 POST 包装器会保留原始请求参数；JSON `body` 或 POST 的 `proxy.params` 都不能替代这个绑定方式。
4. 返回任务 UUID 后通过 `/basicdata/record/getImportProgress` 查询；`result:null` 表示尚无进度，状态 `0` 校验、`1` 导入中、`2` 完成、`3` 失败。只有 `2` 才报告 `completed=true`。失败或超时返回非零退出码，并保留 UUID 便于恢复查询。

默认 `--timeout 10m` 限制整个操作，`--poll-interval 2s` 控制轮询；超时/中断不会取消服务端任务。已有表格按主键规则新增或更新，新表部分失败保留已创建表格并报告 ID。先检查任务/表格，避免盲目重新提交。

导入模板通过 `/basicdata/record/getImportTemplate` 获取。导出通过 `/basicdata/record/export` 发起，再由 `/download/center/queryList` 和 `/download/center/queryFileUrl` 查询任务及下载地址；导出任务已创建不等于文件已生成。

## 字典与连接器同步

- 字典：`/connectorDict/list` 查询，`/basicdata/record/dictionary/import` 导入。
- 连接器配置：`/basicdata/connector/get`、`save`；请求参数通过其 `request/get`、`request/save` 维护。
- 自动同步：`/basicdata/connector/sync/get`、`getRequestField`、`save`；先读取并保留现有规则。
- 手动同步：`/basicdata/connector/sync/manualSync`，按用户要求触发数据变更。
- 开启状态：`POST /basicdata/open/update` 的 `id`、`open`、`openTask` 放在 query。

账户、连接器和模型取数分别复用对应业务 Skill。仅调整同步配置时不隐含立即执行同步。
