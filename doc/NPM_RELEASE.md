# npm 自动发布

## 发布规则

- 仓库：`usiboy/lingtong-cli`，npm 包：`@lingtong-cli/cli`。
- PR 合并到 `main` 产生的 push 触发 `.github/workflows/release.yml`。
  工作流通过 GitHub API 校验该提交确实对应已合并到 `main` 的 PR。
- 普通 PR 更新、未合并关闭、`develop` 和直接推送到 `main` 不会自动发布。
- 保留手动推送稳定版标签 `vX.Y.Z` 的发布入口；预发布标签不在本流程支持范围内。
- 自动版本取 `package.json`、所有稳定版 Git 标签和 npm 已发布稳定版中的最大值，
  然后递增 patch。例如当前为 `0.1.0` 时，下次合并发布 `0.1.1`。
  已创建但尚未完成发布的标签也参与计算，避免重复分配版本。
- 重试同一合并提交时复用已有标签；该版本已经存在于 npm 时直接跳过。
- 若旧提交在创建标签前失败，而更新的提交已有稳定版标签，旧运行不能再分配更高版本。
  已有标签的历史补发会同时禁用 GitHub `latest` 和 npm `latest`，避免升级后代码回退。
- 发布过程中只修改 runner 中的 `package.json`，不向 `main` 回写版本提交，避免循环触发。
  Git 标签、GitHub Release 与 npm 包版本是正式发布记录。

## 执行顺序与安全边界

1. 将全部发布放入同一并发队列，串行分配版本并完成发布。
2. 检查精确的合并提交；执行 `go vet ./...`、`go test ./...`、npm 包测试和仪表盘测试。
3. 测试通过后创建并推送版本标签。
4. 使用 GoReleaser 构建 macOS、Linux、Windows 的 amd64/arm64 制品及 `checksums.txt`。
5. 发布同名 GitHub Release；已完整发布的制品直接复用，未完成的草稿允许替换附件后继续发布。
   若公开 Release 缺少制品则报错，不自动覆盖已公开文件。
6. 将 npm 包版本设为标签版本，并检查全部六个归档的 SHA-256 校验项。
7. 打包真实 npm tarball，在干净目录安装，验证安装器下载、校验、解压及 CLI 版本。
8. 将**同一个已验证的 tarball**发布到 npm，再独立等待注册表中该版本可见。
   npm 接受发布后仍可能处理数分钟，短暂 404 不表示发布失败。每隔 10 秒查询一次，最多 31 次，
   单次请求超时为 10 秒；网络错误、404、408、429 和 5xx 可重试，权限错误直接失败。
   等待只进行公开只读查询，不注入令牌、不重复执行 `npm publish`；超出次数仍会明确报错。

GitHub 的 `GITHUB_TOKEN` 创建标签不会再触发新的工作流，因此标签、GitHub Release 和
npm 发布在同一个工作流内完成。无需额外的 GitHub PAT。

发布队列使用 GitHub 官方支持的 `concurrency.queue: max`，最多等待 100 个运行；
超过平台队列上限的运行需手动重跑。`actionlint` v1.7.12 尚不认识该新增字段，
本地静态检查可仅忽略这条已知语法诊断，其余诊断仍须修复。

发布使用自己的测试门禁，不等同于整个 `CI/CD Pipeline` 通过。建议在仓库分支保护中要求
PR 的 CI 检查通过后才允许合并；本工作流不修改既有分支保护。

## GitHub Actions Secret

必须在仓库的 **Settings → Secrets and variables → Actions** 配置 `NPM_TOKEN`。

令牌要求：

- npm granular access token，具有 `@lingtong-cli/cli` 或其 scope 的写权限。
- 开启 **Bypass 2FA**，否则无人值守发布可能要求一次性验证码。
- 尚未到期；若配置了 IP 白名单，必须允许 GitHub runner 的出口地址。

令牌只在最后的 npm 发布步骤通过 `NODE_AUTH_TOKEN` 注入。
不要提交真实令牌、`.npm_key` 或含凭据的 `.npmrc`；相关文件已加入 `.gitignore`。

轮换时，可将仅含令牌的本地文件通过标准输入写入 Secret：

```bash
gh secret set NPM_TOKEN --repo usiboy/lingtong-cli < /path/to/npm-token-only.txt
```

不要把令牌直接放在命令参数、日志或 PR 内容中。`npm whoami` 成功只代表认证有效，
不等于发布权限、2FA 策略和有效期全部满足；令牌必须定期检查、轮换。
后续可迁移为 npm Trusted Publishing（OIDC），以消除令牌到期维护。

## 失败恢复

在 GitHub Actions 的 **Release** 运行页面选择 **Re-run all jobs**。
已有版本标签会被复用，已完成 npm 发布的版本不会重新发布。

也可从 `main` 手动补发已有标签，不会分配新版本：

```bash
gh workflow run release.yml --repo usiboy/lingtong-cli --ref main -f tag=v0.1.1
```

- npm 失败、GitHub Release 成功：复用原制品后重试 npm。
- 令牌失效或权限不足：更新 `NPM_TOKEN` 后重跑，不要删除或重建发布标签。
- 日志已有 `+ @lingtong-cli/cli@版本号`，但随后确认版本失败：先检查 npm 官方注册表。
  这可能只是注册表处理延迟，不要立刻重复发布；版本已经可见时，重跑会跳过重复发布。
- 新版本已经发布，再补发较旧版本：使用 npm `historical` dist-tag，避免回退 `latest`。
- 已有较新 Git 标签时，即使较新版本的 npm 发布尚未完成，历史补发也不会成为任一渠道的 `latest`。
- 旧提交尚未创建标签且已落后于已标记提交：工作流报错拒绝分配新版本；直接使用较新的发布运行，
  不要为了补发旧提交而删除较新标签。
- 注册表或 GitHub API 不可用：立即失败，不把网络错误当成“版本不存在”。

手动重试必须使用真实存在的稳定版标签。若需主动升级 major/minor，可在经过测试、已合并的
提交上手动创建对应标签；后续自动发布继续在该版本基础上递增 patch。

## 本地检查

```bash
npm run test:npm
make test-dashboard-skill
go vet ./...
go test ./...
node scripts/release.js plan
git diff --check
```

`plan` 仅读取 Git 和 npm 状态，不创建标签或发布包。完整的首次发布仍需合并配置 PR，
由 GitHub-hosted runner 执行验证；本地测试通过不能代替真实 npm 发布成功。
