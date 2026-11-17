# 离线校验

需要 Node.js 18 或以上，无第三方依赖。命令中的路径相对于本 skill 目录；安装后先定位 SKILL.md 所在目录。

```bash
node scripts/validate-dashboard.mjs assets/examples/operating-overview.json \
  --schemas assets/examples/metadata.json \
  --pivot-configs assets/examples/pivot-configs.json \
  --require-bindings
```

输入是完整 config 或 `{appId,id?,name?,config}` 保存对象，不是 CLI 响应信封。回读结果先提取 result/config。输出 JSON：`ok`、`errors[{path,message}]`、`warnings`、`summary`；退出 0 表示静态校验通过，2 表示配置/文件/参数有误。未知配置扩展字段保留，不会自动改文件。

## 元数据与阶段

`--schemas` 是本地元数据文件：

```json
{
  "tables": [{
    "id":1001,"schemaId":2001,
    "columnsSchema":[{"id":2,"key":"data2001.2","title":"销售额","type":"number"}]
  }],
  "views":[{"id":8002,"tableId":1001,"type":"pivot"}]
}
```

上述为形状示意，实际从表格详情/Schema/视图查询汇总。公式字段附 functionFieldType。每个被引用表都要收录；views 可省略，提供时必须包含所有已绑定视图。视图 type 使用 basic/group/pivot。不要把取不到的视图冒充空列表。

- 默认：适合首次保存。表格组件缺 tableViewId 为 warning；校验结构、字段、布局与筛选依赖。
- `--require-bindings`：保存后使用；视图 ID 缺失变 error。配合 views 校验归属和类型。
- `--pivot-configs`：输入透视保存请求数组（每项含 tableId/viewId/businessId/businessType/config），检查引用、维度、度量、合计与汇总。
- 未给 schemas 时有 warning；此时无法检查字段是否存在、字段类型和视图归属，不能称为完成数据绑定验证。
- 单指标模式仅完整校验 sources[0]；其余保留来源检查 ID 唯一性并提示 warning，切回多指标时再校验其数据绑定和占比分母。

## 覆盖与边界

检查全部 9 类组件及其 type、唯一 ID、24 列布局、缺失/孤立/重叠布局、字段元数据、聚合与维度配置、指标源/分母/同环比、三类切片器、静态过滤、视图和透视第二阶段绑定。

按当前编辑器规范生成配置；历史读回可能缺默认字段或有旧布局，先理解 warnings/errors 再决定是否规范化。对任意服务端未知扩展不作拒绝。

离线校验不验证权限、接口兼容、真实数据数值、后端计算成功、视觉拥挤、过滤组合业务意义，也不实现服务端聚合。错误输出到具体 path 后修正并重新运行。最终还要保存回读、查询组件结果和浏览器检查。

仓库维护时运行 `make test-dashboard-skill`。测试覆盖全部示例、各类错误和命令行失败路径，不触发线上写入。
