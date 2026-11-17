# CLI 调用与验证

## 发现和读取

复用当前 profile/授权；查看命令与共享规范：

```bash
lingtong-cli auth status
lingtong-cli schema module dashboard
lingtong-cli schema path /dashboard/save
lingtong-cli service dashboard --help
lingtong-cli service dashboard page --appId 942 --pageNum 1 --pageSize 100 --envelope
lingtong-cli service dashboard get --id 3001 --envelope --omit-null=false
```

替换示例 ID，分页读完需要的列表。字段与视图按表格技能查询，生成 validation.md 所述元数据文件。CLI schema 摘要不等于全部嵌套 DTO，组件详细语义看本技能，公共路径在 internal/openapi/spec/openapi.json。

## 保存

接口逻辑路径为 /dashboard/save；主应用 Web 路径为 /admin/ltapp/dashboard/save。CLI 使用逻辑路径，不重复添加 Web 前缀。

保存 payload.json 前先离线校验。CLI 的 --body/--data 接收 JSON 文本，不支持 @file；推荐从文件读取后作为独立 argv 传递，避免复杂 JSON 的 shell 转义：

```python
import json, subprocess
from pathlib import Path
payload = json.loads(Path('payload.json').read_text())
result = subprocess.run([
    'lingtong-cli', 'service', 'dashboard', 'save',
    '--body', json.dumps(payload, ensure_ascii=False),
    '--envelope', '--omit-null=false',
], check=True, capture_output=True, text=True)
print(result.stdout)
```

新增省略 id；更新保留 appId/id。改布局/组件合并最新 config；改名仅 {id,appId,name}。用户要求复制仪表盘时复用合适的真实数据源，重新处理组件 ID 和视图绑定，不能把源仪表盘专属 tableViewId 当成新视图。

保存直接生效。执行用户已授权的变更后同时检查进程退出码、信封 ok 与业务 success；通常业务对象位于 `.data`、实体位于 `.data.result`，以实际返回为准。不确定时回读，避免重复新建。服务端返回的 tableViewId 和规范化 sources 作为下一步依据。

## 查询及两阶段验证

```bash
lingtong-cli service dashboard get --id 3001 --envelope --omit-null=false
lingtong-cli service dashboard query --body '{"id":3001}' --envelope
# 较旧 CLI spec 仍将 query 标为 GET 时，显式沿用当前前端 POST 协议：
lingtong-cli api POST /dashboard/query --data '{"id":3001}' --envelope
```

本技能随附规范将查询同步为前端使用的 POST + JSON {id}；它基于当前主应用调用点，实际部署版本仍需联调，不声称完成线上兼容验证。若 schema 仍显示旧 GET，检查 LINGTONG_OPENAPI 或用户目录覆盖规范是否过期，不擅自覆盖用户自定义规范。

图表和指标响应按组件 id 匹配；多指标还按 sourceId 检查每项 status/errorMessage。表格和透视用各自视图接口取数；配置了全局过滤/切片器时，还要在带 dashboardId 的实际表格上下文验证，不把单独的透视查询结果当成联动已生效。

透视表流程见 tables-and-pivot.md：先回读视图 ID，再提交透视配置，再读配置/数据。权限限制沿用平台，不能通过组件开关替代授权。

浏览器复核：9 类组件均可渲染；数值/日期边界符合需求；筛选后相关组件刷新；布局无拥挤或遮挡。没有实际浏览器/线上验证时在交付中明确范围。
