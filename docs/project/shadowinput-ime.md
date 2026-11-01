# shadowinput 系统输入法预览

主入口为系统输入法设置，复用 Arion 的 Rime 拼音引擎、macOS InputMethodKit 宿主及 Windows TSF 服务。目标是从系统输入菜单切换后直接打字，普通词语上屏后出现独立 AI 短语联想。官网的 v0.1.0-preview.6 已包含该实现，提供 DMG 和 EXE 下载；完整系统输入验收仍待完成，不能视为已经可替代微信输入法的成品。

## 使用流程

1. 打开 shadowinput，在“输入法设置”安装 / 启用组件。macOS 首次需要系统允许；Windows 注册 TSF 服务需要管理员授权。
2. 在 macOS 菜单栏输入菜单或 Windows 的 Win + 空格菜单中选择 shadowinput。只有系统实际列出并选中输入法后，才可进行下面的输入验收。
3. 在编辑器里输入拼音。空格或数字选择普通候选，Esc 取消。普通拼音使用本地 Rime，不依赖 ChatGPT。
4. 在主程序下载并启动本地模型，明确开启“AI 候选”，并保持主程序运行。正常按空格分词上屏；至少 2 个字后暂停约 180 毫秒，请求一次短语续写。模型推理耗时另计。已下载且 AI 开关开启时，下次启动会预热模型；不会自动下载。
5. 只有请求完整成功且通过检查，才显示独立的 `AI · Tab` 候选；Tab 或点击只追加新短语，保留已上屏文字。继续输入、移动光标、删除、换焦点或关闭 AI 会撤销旧候选。请求超时（2 秒）、重复原文、回复式结果或不可用时不显示 AI 行。

拼音容错默认开启，复用 Rime 已有的纠错能力；它不调用大模型。短语联想只追加文字，不自动发送聊天、不读取整篇文档。ChatGPT 登录、润色、翻译、改写、续写和完整差异审阅保留在“长文写作工具”中，模型选择互不影响。AI 行不占用普通候选的数字槽位；AI 文本不写入用户词库。

## 错误恢复和更新

- 个人风格读取失败时，输入法设置、写作页和个人风格页显示同一错误和重试入口；不覆盖异常文件。
- 自动候选固定使用本地模型，不依赖 ChatGPT 登录和额度，不会退回云端。主程序中启动模型、暂停或失败时显示本地状态。主动写作请求的账户错误仍由写作页处理。
- Windows 安装前读取原生注册状态：首次安装和同版本修复使用 `install`，版本变化或迁移使用已有的 `update` 事务，传入之前的构建指纹并保留底层回滚机制；无法确认状态时不修改注册。
- Windows 每个候选请求独立持有取消状态。继续输入或结束组合会撤销旧请求；WinHTTP 使用异步请求，由工作线程取消连接，旧完成事件不能清除新请求。回调完全结束前保留其缓冲区和 DLL 引用。
- macOS 菜单使用透明镂空 SI 模板，由系统决定黑白颜色；应用、安装包和官网保留黄色品牌图标。菜单资源名包含内容指纹，父输入源和子模式都声明模板属性。

## 隐私和隔离

AI 默认关闭。自动联想使用已下载并启动的 Qwen 模型，只接收当前输入位置最近连续上屏的文字（最多 128 字）、风格及相关词表。上下文仅在内存中保留，30 秒超时或光标/焦点变化后失效。服务只监听随机的 `127.0.0.1` 端口，使用每次启动随机生成的令牌；拒绝浏览器来源请求，限制输入及结果长度，不持久化候选或输入日志。关闭 AI 会移除发现文件并取消当前请求。补全过滤是启发式检查，不保证建议符合语义；用户按 Tab 前应确认结果。

macOS 注册标识为 `org.shadowinput.inputmethod.shadowinput`，模式为 `org.shadowinput.inputmethod.shadowinput.pinyin`。Windows 使用独立的服务和语言配置 GUID。数据与上游分开存放；与 Arion 同时安装的完整兼容性尚未验收。macOS 开发预览的学习加密密钥保存在独立的登录钥匙串条目中，不声称具备上游正式签名版的 Data Protection 设备绑定保证。

## 构建

先准备 Rime SDK、固定版本依赖和词库，再构建主程序；模型权重不随包分发。macOS 需要 Xcode，原生学习层还使用 Swift/C++ 工具链。

```sh
node scripts/stage-keyboard-runtime.mjs
node scripts/shadowinput.mjs build macos
node scripts/test-shadowinput-ime.mjs
node scripts/test-shadowinput-candidates.mjs
node scripts/shadowinput.mjs check
```

Windows 的输入法构建需要 x64 MinGW-w64 GCC、CMake、Ninja、windres；不是 MSVC。Windows 本机运行 `node scripts/stage-keyboard-runtime-windows.mjs` 后再执行 `node scripts/shadowinput.mjs build windows`。

macOS 交叉构建 Windows 时，先完成 macOS Rime staging，以获得主机版本 `rime_deployer`。把 `SHADOWINPUT_DATA_DEPLOYER` 导出为 `.workspace/keyboard-runtime/source-build/<构建指纹>/build/bin/rime_deployer` 的实际路径，然后运行：

```sh
SHADOWINPUT_CROSS_WINDOWS=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc node scripts/stage-keyboard-runtime-windows.mjs
node scripts/shadowinput.mjs build windows
```

`node scripts/test-shadowinput-candidates.mjs` 在主机运行候选状态测试，在 Windows 上还运行真实回环 HTTP 取消/成功响应测试。macOS 可加 `--cross-windows` 仅交叉编译这个 Windows 测试，不能将其记为运行通过。`node scripts/shadowinput.mjs check` 已包含主机候选状态测试、Go 回归测试和静态检查。

脚本只用主机 deployer 编译词库，不运行 Windows 二进制。构建脚本在隔离的暂存源码中替换产品注册标识，保留上游原生源文件及版权。包内包含 Rime/词库、Go 模块、Source Han Sans 和平台运行库的许可证。暂存资源清单和本机二进制不提交到源码仓库。

## 2026-11-01 验收记录

- macOS 原生引擎和 NSTextView 测试通过：未上屏拼音不请求模型、连续空格分词累积上下文、短语只追加一次、旧修订和光标移动后拒绝结果、失焦取消、拼音纠错和正常拼音候选。
- 本地候选服务通过鉴权、浏览器来源拒绝、取消、关闭后拒绝旧令牌请求以及不合法候选过滤测试。
- 一个真实 ChatGPT 账号在主程序中恢复了会话；合成短句请求返回过 HTTP 200 的完整候选，约 2.39 秒。这是服务请求证据，不是系统候选窗验收。
- macOS 系统列表现已显示 shadowinput；用户截图及独立 TIS 查询确认组件已启用且可选中（`ownCount: 2, available: true, selected: true`）。本轮本机主程序和输入组件均已更新，旧包保留用于回退，原有账户恢复成功。用户已在文本编辑中确认上一版出现 `AI · Tab` 行。本轮改为本地上屏后联想，新的系统端到端体验仍需另行验收。
- Windows x64 TSF、主程序已交叉编译，PE 依赖闭包经过检查。没有 Windows 真机的安装、切换、输入、候选窗口、升级和卸载验收结果。

- 本轮先复现再修复：配置错误在输入法页被隐藏、401/403 后账户仍可用、迟到错误覆盖新来源、Windows 跨版本安装被拒绝，以及编辑后旧候选请求没有取消。对应回归测试现已通过。
- 项目检查、macOS 原生输入夹具、候选状态测试及双端构建通过；Windows WinHTTP 夹具仅交叉编译，没有真机执行结果。Crapper、Dryer、Mutator 未接入或运行，不声称达到相关指标。

不自动注销、不修改 HIToolbox 偏好或替换其他输入法。图标更新后若新进程的 TIS 图标 URL 已指向新文件，而菜单仍显示旧图案，属于菜单界面缓存；不要据此反复改变图案或重新安装。官网以开发预览版提供下载，系统端到端输入验收完成前不标为稳定版。

Windows 仍需检查不同缩放、浏览器/Office/记事本、安全输入框、快速换焦点和停用/卸载。AI 行服从原生候选 UI 的显示限制；屏幕阅读器及编辑器自绘 TSF UI 的 AI 候选能力尚未验收。自动候选不消耗 ChatGPT 套餐；长期输入的延迟、内存和耗电仍需实机观察。

## 本地短语联想验证

- 原生测试先复现“未上屏的拼音仍会启动请求”，修改后验证连续分词上屏、只追加一次、光标及焦点撤销；共享状态测试覆盖 UTF-8 长度、30 秒失效和 180 毫秒等待。
- Go 测试先复现自动输入会访问 ChatGPT、两字上下文被拒绝、整句回显被接受，以及输入设置缺少独立本地入口，修改后通过。合成样例暴露出回复式结果和边界重复，增加保守过滤；这不构成语义质量保证。
- `node scripts/shadowinput.mjs check`（含 race、vet）、原生 AppKit 测试、候选状态/Windows HTTP 交叉编译、固定 pnpm 的仓库检查通过。中英文、明暗主题的本地就绪/未安装/加载界面已渲染检查。
- macOS 构建来自 `a8d20603`，Windows 来自 `19a7c24c`；两者均为干净源码构建，后者不包含仅修正 macOS block 返回类型的提交。macOS 主程序和输入组件已安装，签名校验通过，二进制与构建产物哈希一致，旧包保留。已安装模型自动预热成功。
- 已安装主程序的真实本地候选接口：三个合成短语返回完整结果，耗时分别为 0.158、0.109、0.109 秒；一个边界重复样例被过滤为 204。上述耗时不含输入法停顿等待和窗口渲染，不是统计性能基准。
- 用户选择稍后验证实际键盘输入与 Tab 追加；端到端打字仍记为待测。Windows 原生行为也未验证。随后按用户要求发布了 v0.1.0-preview.6 并更新官网，安装包与源码、校验值一起提供；发布不代表这些验收已通过。

对应验收约定见[本地输入短语规范](../../specs/012-native-phrases.md)。
