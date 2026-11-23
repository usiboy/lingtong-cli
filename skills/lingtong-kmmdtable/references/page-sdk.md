# 快麦页面组件 SDK

自定义页面需要选择快麦 ERP 商品时，可以接入 SDK。通过 `<script src>` 加载 [kuaimai-page-sdk.js](https://app1.ltpass.com/mdtable/kuaimai-page-sdk.js)，再调用浏览器全局 `KuaimaiPageSDK.selectProduct`。SDK 包含样式和运行依赖，每页引用一次，放在调用它的页面脚本之前。其余页面代码仍内联。没有商品选择需求的页面无需引入 SDK。

## 接入与编辑回显

下面是单选示例。将入口放在需要商品的业务表单里，例如规则类型为“商品订单量监控”时显示“选择商品”。需要多选时传 `multiple: true`，用完整商品行的数组保存和回显。

```html
<button id="pick-product" type="button">选择商品</button>
<p id="product-status" role="status"></p>
<script src="https://app1.ltpass.com/mdtable/kuaimai-page-sdk.js"></script>
<script>
  const pickButton = document.getElementById('pick-product');
  const statusNode = document.getElementById('product-status');
  let selectedProduct = null; // 编辑时从现有业务记录恢复完整商品行

  pickButton.addEventListener('click', async () => {
    pickButton.disabled = true;
    try {
      if (!window.KuaimaiPageSDK?.selectProduct) {
        throw new Error('商品选择组件加载失败，请刷新后重试');
      }
      const result = await window.KuaimaiPageSDK.selectProduct({
        multiple: false,
        selected: selectedProduct ? [selectedProduct] : [],
        pageSize: 50,
        autoLoad: false,
        validate(rows) {
          if (rows.length !== 1) throw new Error('请选择一个商品');
        }
      });
      if (result === null) return;
      selectedProduct = result.rows[0];
      statusNode.textContent = selectedProduct.title || '已选择商品';
      // 点业务表单的保存按钮时，再提交商品字段和完整商品行。
    } catch (error) {
      statusNode.textContent = error.message || '选择商品失败';
    } finally {
      pickButton.disabled = false;
    }
  });
</script>
```

- 确认返回 `{ rows }`，取消返回 `null` 并保留表单原选择。打开弹窗时保持 `autoLoad: false`，商品列表等用户点击“查询”后加载。每页支持 50 或 100 条。
- SPU 业务用 `sysItemId` 作为商品主档 ID，`outerId` 是主商家编码。规格识别使用 `sysItemId + '_' + (sysSkuId || 0)`，`skuOuterId` 是规格商家编码。保留长 ID 的字符串精度，勿转成 `Number`。
- 保存规则时沿用现有表单与数据写入流程。除业务字段外，若需刷新后仍可在 SDK 回显，另存完整商品行（例如在规则表的长文本字段存 JSON）。编辑时恢复成对象并传给 `selected`。历史记录只有名称和编码时，表单仍可展示这些字段，完整商品行在重新选择后补齐。“展示已选商品”直接使用传入的商品行。
- 若表单使用原生 `dialog.showModal()`，SDK 挂到 `body` 的弹层可能被浏览器顶层弹窗遮挡，单纯增大 `z-index` 无法解决。此时把选择器放在可交互的顶层弹窗内，或使用非原生模态层承载表单。验证选择器关闭后表单内容和焦点都能恢复。

## 浏览器运行时

在草稿和发布页面中，SDK 通过当前站点的业务接口发请求，由浏览器附带绫通 Cookie。业务地址与 SDK 的 OSS 地址分别处理。部署环境默认使用同源接口，只有确认环境存在同源代理前缀时才设置 `KuaimaiPageSDK.configure({ apiBase: '/dev-api' })`。

SDK 打开时按以下顺序解析连接器：

1. GET `/admin/upms/ltapp/refreshUserInfo`，取 `result.user.tenantId`。
2. GET `/admin/ltapp/account/list?pageNum=1&pageSize=1000`，在 `result.data` 中取第一条 `connector === 'kuaimai-cli'` 的账号。这一步按当前约定只查第一页。
3. 用账号的 `name` 作为 `authName`，用 `envs.split(',')[0].trim()` 作为 `env`。查不到账号时，核对草稿实际所在站点和浏览器登录的租户，请用户在该环境、该租户完成现有授权流程。
4. 用户查询商品时，POST `/admin/ltapp/connector/invoke`，请求包含 `connector: 'kuaimai-cli'`、`authName`、`env`、`tenantId`、`method: '/stock/queryExtList'` 和 ERP 参数对象 `requestBody`。其他筛选数据也通过连接器按需查询。

SDK 已封装这些步骤，页面无需重复实现。CLI 的 Token 不写入 HTML，页面也不手工读取或拼装 Cookie。

## 开发取数与联调

- 管理绫通应用、表格和页面，以及读写草稿，继续用当前 `lingtong-cli` 的 Profile 和授权。按 [页面接口与回传](page-api.md#保存草稿) 保存原 `pageId` 的新草稿并回读。
- Agent 开发阶段有可用的快麦系统授权时，可以用 `kuaimai-cli` 直接读取 ERP 真数据。若需用 `lingtong-cli` 联调指定快麦连接器，可用下方命令查询，环境和账号取自已核对的目标租户。先看 `connector invoke --help`：支持直调代理的版本可直接查询，旧版本可能在无缓存时创建通用工作流，只有在该操作已获授权时才使用旧版本。CLI 取数成功后，还要在目标草稿里验证浏览器会话的实际访问。

```bash
lingtong-cli --profile '<目标Profile>' connector invoke \
  --connector kuaimai-cli \
  --method /stock/queryExtList \
  --auth-account '<已确认的连接器账号名>' \
  --env '<该账号首个环境>' \
  --body '{"pageNo":1,"pageSize":20}' \
  --format json
```

真实联调在已登录的目标草稿中完成。先点击查询，确认网页业务响应成功、`receiptStatus === 'SUCCESS'` 且 ERP `responseBody` 的业务状态成功。选中商品后核对回填字段。页面有持久化需求时，在授权范围内保存测试记录，回读核对主档 ID 和商品行，再刷新并进入编辑，检查表单和 SDK 的“展示已选商品”。若缺少当前租户的连接器账号，记录商品查询受阻的原因，分别报告已经完成的保存、回显验证。
