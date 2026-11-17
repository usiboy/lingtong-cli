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

## 增量列表筛选和排序

`POST /basicdata/record/listIncremental` 和 `GET /basicdata/record/listIncremental` 均支持 CLI 传递筛选和排序条件。POST 使用 JSON 数组，推荐 CLI 优先使用；GET 继续兼容已有客户端。目标服务端须包含“支持增量列表筛选和排序”改动。先查询表 Schema，获取 `schemaId`、基础资料 ID 和字段的**列 ID**；筛选、排序的 `columnId` 应为列 ID 字符串，而不是字段名。租户和用户取当前登录态。


推荐使用 POST：

```bash
lingtong-cli api POST /basicdata/record/listIncremental --data '{"schemaId":100,"basicDataId":2966,"pageNum":1,"pageSize":20,"filterCondition":[{"columnId":"10","comparator":"eq","value":"李莹"}],"sortingCondition":[{"columnId":"36","asc":false}]}' --envelope
```

后续增量查询在同一 JSON 顶层加入 `"lastSyncTime":1720000000000`。`filterCondition` 和 `sortingCondition` 是数组，元素字段分别与下表的索引参数一致；`column`、`asc` 也可作为传统全量查询排序参数，租户及用户从登录态获取。具体筛选值和排序规则与 GET 相同。

GET 调用方式：

```bash
lingtong-cli api GET /basicdata/record/listIncremental --params '{"schemaId":100,"basicDataId":2966,"pageNum":1,"pageSize":20,"filterCondition[0].columnId":"10","filterCondition[0].comparator":"eq","filterCondition[0].value":"李莹","sortingCondition[0].columnId":"36","sortingCondition[0].asc":false}' --envelope
```

| 参数 | 含义 |
| --- | --- |
| `schemaId` | 表结构 ID；增量变更记录按它隔离。 |
| `basicDataId` | 基础资料表 ID；增量查询缺省时尝试通过 `schemaId` 解析。 |
| `lastSyncTime` | 上次同步时间，**毫秒时间戳**；不传时返回全量分页，传入时返回该时间之后至本次同步时间的变更。 |
| `pageNum`、`pageSize` | 全量查询页码和每页数量，默认分别为 1、20；增量变更不是按这两个参数分页。 |
| `viewId` | 可选已有视图 ID，沿用其筛选条件；请求筛选与视图筛选共同生效。 |
| `filterCondition[0].columnId` | 第一个筛选条件的 Schema 列 ID，字符串。 |
| `filterCondition[0].comparator` | 小写运算符：`eq`、`ne`、`neq`、`gt`、`gte`、`lt`、`lte`、`between`、`ct`、`nct`、`in`、`empty`、`notempty`。 |
| `filterCondition[0].value` | 字符串比较值；`empty`、`notempty` 不需要。`between` 可传 JSON 数组字符串如 `"[10,20]"` 或逗号分隔的两个值；`in` 可传 JSON 数组字符串如 `"[\\"甲\\",\\"乙\\"]"`。 |
| `sortingCondition[0].columnId` | 第一个排序条件的 Schema 列 ID，字符串。 |
| `sortingCondition[0].asc` | `true` 升序，`false` 或未传降序。请求排序非空时优先于已有视图排序。 |

多个条件按 `filterCondition[1].columnId`、`filterCondition[1].comparator`、`filterCondition[1].value` 和 `sortingCondition[1].columnId`、`sortingCondition[1].asc` 等递增下标传递。GET 参数必须使用示例中的扁平键，**不要**把 `filterCondition` 或 `sortingCondition` 写成嵌套 JSON 数组：通用 `api` 命令不会将对象数组转换为 Spring 索引参数。无效列 ID 或不支持的运算符会被服务端跳过，使用前应核对 Schema。

响应 `result.syncType` 为 `FULL` 时，记录在 `result.data` 中；为 `INCREMENTAL` 时，变更在 `result.rows` 中。`result.syncTime` 是服务端本次同步时间（毫秒），完成本次处理后可用于下一次 `lastSyncTime`。增量行中 `dataId` 是记录 ID，`operationType` 为 `UPDATE` 或 `DELETE`；传入筛选后，已有行不再匹配时也可能以 `DELETE` 通知客户端移除，`rowData` 不一定存在。增量排序只排列本次 ES 命中的变更行；删除等未命中行排在其后，不能据此当作整个数据集的全局排序。全量分页读取期间要自行处理数据变化及断点，不要仅以单页结果视为完整快照。

**当前服务端限制**：增量流程内部使用默认分页查询 ES，单次最多取前 20 条命中数据。一次同步涉及超过 20 条非删除变更时，后续记录可能被误判为不匹配并返回 `DELETE`，排序也不能覆盖全部变更。在服务端修复并部署前，不要将这种结果直接用于覆盖或删除本地数据；应核对结果并控制或复查变更批次。
