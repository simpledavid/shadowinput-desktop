# shadowinput 系统输入法预览

本分支把主入口改为系统输入法设置，复用 Arion 的 Rime 拼音引擎、macOS InputMethodKit 宿主及 Windows TSF 服务。目标是从系统输入菜单切换后直接打字，普通候选旁出现独立 AI 候选。当前是待系统验收的实现，不能视为已经可替代微信输入法的成品；官网的 v0.1.0-preview.5 仍是先前的写作工具版本，不含本分支的输入法。

## 使用流程

1. 打开 shadowinput，在“输入法设置”安装 / 启用组件。macOS 首次需要系统允许；Windows 注册 TSF 服务需要管理员授权。
2. 在 macOS 菜单栏输入菜单或 Windows 的 Win + 空格菜单中选择 shadowinput。只有系统实际列出并选中输入法后，才可进行下面的输入验收。
3. 在编辑器里输入拼音。空格或数字选择普通候选，Esc 取消。普通拼音使用本地 Rime，不依赖 ChatGPT。
4. 在主程序选择 ChatGPT 或已启动的本地模型，明确开启“AI 候选”，并保持主程序运行。当前首个普通候选必须覆盖完整拼音组合、至少 4 个字符，停顿约 700 毫秒后才请求一次自然表达建议。
5. 只有请求成功且通过检查，才显示独立的 `AI · Tab` 候选；Tab 或点击采纳。继续输入、移动组合位置、换焦点或关闭 AI 会撤销旧候选。超时、无有效改写、不可用或触发风险检查时不显示 AI 行。

这一版的输入候选只做同义润色，不自动发送聊天、不读取整篇文档。翻译、改写、续写和完整差异审阅仍在“长文写作工具”中。AI 行不占用普通候选的数字槽位；未采纳的 AI 文本不写入用户词库。

## 隐私和隔离

AI 默认关闭。ChatGPT 模式开启后，当前尚未上屏的组合短句、所选风格和相关词表会发送给 OpenAI，并消耗授权套餐用量。本地模式使用已经下载并启动的 Qwen 模型。服务只监听随机的 `127.0.0.1` 端口，使用每次启动随机生成的令牌；拒绝浏览器来源请求，限制输入及结果长度，不持久化候选或输入日志。关闭 AI 会移除发现文件并取消当前请求。风险检查是启发式过滤，不保证识别所有语义变化。

macOS 注册标识为 `org.shadowinput.inputmethod.shadowinput`，模式为 `org.shadowinput.inputmethod.shadowinput.pinyin`。Windows 使用独立的服务和语言配置 GUID。数据与上游分开存放；与 Arion 同时安装的完整兼容性尚未验收。macOS 开发预览的学习加密密钥保存在独立的登录钥匙串条目中，不声称具备上游正式签名版的 Data Protection 设备绑定保证。

## 构建

先准备 Rime SDK、固定版本依赖和词库，再构建主程序；模型权重不随包分发。macOS 需要 Xcode，原生学习层还使用 Swift/C++ 工具链。

```sh
node scripts/stage-keyboard-runtime.mjs
node scripts/shadowinput.mjs build macos
node scripts/test-shadowinput-ime.mjs
node scripts/shadowinput.mjs check
```

Windows 的输入法构建需要 x64 MinGW-w64 GCC、CMake、Ninja、windres；不是 MSVC。Windows 本机运行 `node scripts/stage-keyboard-runtime-windows.mjs` 后再执行 `node scripts/shadowinput.mjs build windows`。

macOS 交叉构建 Windows 时，先完成 macOS Rime staging，以获得主机版本 `rime_deployer`。把 `SHADOWINPUT_DATA_DEPLOYER` 导出为 `.workspace/keyboard-runtime/source-build/<构建指纹>/build/bin/rime_deployer` 的实际路径，然后运行：

```sh
SHADOWINPUT_CROSS_WINDOWS=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc node scripts/stage-keyboard-runtime-windows.mjs
node scripts/shadowinput.mjs build windows
```

脚本只用主机 deployer 编译词库，不运行 Windows 二进制。构建脚本在隔离的暂存源码中替换产品注册标识，保留上游原生源文件及版权。包内包含 Rime/词库、Go 模块、Source Han Sans 和平台运行库的许可证。暂存资源清单和本机二进制不提交到源码仓库。

## 2026-11-01 验收记录

- macOS 原生引擎和 NSTextView 测试通过：完整组合提取、旧修订拒绝、AI 文本只提交一次、提交后普通拼音继续工作、失去焦点后拒绝迟到结果。
- 本地候选服务通过鉴权、浏览器来源拒绝、取消、关闭后拒绝旧令牌请求以及不合法候选过滤测试。
- 一个真实 ChatGPT 账号在主程序中恢复了会话；合成短句请求返回过 HTTP 200 的完整候选，约 2.39 秒。这是服务请求证据，不是系统候选窗验收。
- macOS 输入组件已复制并签名验证。用户报告点击系统“允许”后，新进程 TIS 查询仍返回 `ownCount: 0, available: false, selected: false`；系统设置中搜索 shadowinput 也没有结果。系统启用、选中、实际按键分发、可见 AI 行和 Tab 上屏尚未通过。主程序必须显示待启用，不能依据当前进程的注册返回值报告成功。
- Windows x64 TSF、主程序已交叉编译，PE 依赖闭包经过检查。没有 Windows 真机的安装、切换、输入、候选窗口、升级和卸载验收结果。

macOS 下一步是在用户保存工作、主动注销并重新登录后复查系统列表。注销不是已证实的修复，不自动执行、不修改 HIToolbox 偏好、不重启系统输入代理，也不替换其他输入法。若登录后仍不可见，需要继续定位系统注册问题。此门槛通过前不替换官网现有安装包。

Windows 仍需检查不同缩放、浏览器/Office/记事本、安全输入框、快速换焦点和停用/卸载。AI 行服从原生候选 UI 的显示限制；屏幕阅读器及编辑器自绘 TSF UI 的 AI 候选能力尚未验收。自动候选请求会消耗套餐用量，长期输入的性能与用量也需要实机观察。
