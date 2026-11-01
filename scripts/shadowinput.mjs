import { execFileSync, spawnSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { mkdirSync, cpSync, copyFileSync, readFileSync, readdirSync, rmSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { verifyWindowsPEDependencies } from './lib/windows-pe.mjs';
import { stageGoModuleNotices } from './lib/go-module-notices.mjs';
import { desktopFonts } from './lib/desktop-fonts.mjs';
import { createInstaller } from './lib/shadowinput-installers.mjs';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const mode = process.argv[2] ?? 'check';
const nativeTarget = process.platform === 'darwin' ? 'macos' : process.platform === 'win32' ? 'windows' : '';
const target = process.argv[3] ?? nativeTarget;
const source = './apps/desktop/cmd/shadowinput';
const repository = 'https://github.com/simpledavid/shadowinput-desktop';
function output(command, args, env = process.env) {
  return execFileSync(command, args, { cwd: root, env, encoding: 'utf8' }).trim();
}
function run(command, args, env = process.env, cwd = root) {
  const result = spawnSync(command, args, { cwd, env, stdio: 'inherit' });
  if (result.error) throw result.error;
  if (result.status !== 0) process.exit(result.status ?? 1);
}
if (mode === 'check') {
  run('go', ['test', '-race', './packages/core/...', './apps/desktop/infrastructure/chatgptplan', './apps/desktop/infrastructure/localwriting', './apps/desktop/infrastructure/llm/...', './apps/desktop/infrastructure/rewriting', './apps/desktop/infrastructure/provision', './apps/desktop/cmd/arion-llm', './apps/desktop/infrastructure/writingprefs', './apps/desktop/infrastructure/inputcandidate', './apps/desktop/infrastructure/shadowime', './apps/desktop/ui/shared/writing', './apps/desktop/internal/oauthloopback', './apps/desktop/internal/shadowinput', './apps/desktop/ui/features/writing', './apps/desktop/ui/shared/copy']);
  run('go', ['vet', source, './apps/desktop/infrastructure/chatgptplan', './apps/desktop/internal/shadowinput']);
  run(process.execPath, ['scripts/check-repository.mjs']);
} else if (mode === 'dev') {
  const platform = nativeTarget === 'macos' ? 'darwin-arm64' : 'windows-amd64';
  run(process.execPath, ['scripts/stage-llm-runtime.mjs', platform]);
  run('go', ['run', './cmd/shadowinput'], process.env, path.join(root, 'apps/desktop'));
} else if (mode === 'build' || mode === 'package') {
  if (!['macos', 'windows'].includes(target)) throw new Error('Choose macos or windows');
  if (target === 'macos' && process.platform !== 'darwin') throw new Error('macOS builds require a Mac');
  const version = mode === 'package' ? process.argv[4] : 'v0.1.0-dev';
  if (!version || !/^v\d+\.\d+\.\d+(?:-[a-zA-Z0-9.]+)?$/.test(version)) throw new Error('Provide a version such as v0.1.0-preview.1');
  const revision = output('git', ['rev-parse', 'HEAD']);
  const dirty = output('git', ['status', '--porcelain', '--untracked-files=normal']) !== '';
  if (mode === 'package') {
    if (dirty) throw new Error('Release packages require a clean source checkout');
    if (output('git', ['rev-parse', `${version}^{commit}`]) !== revision) throw new Error('Release tag must identify HEAD');
  }
  // Versioned package payloads do not replace the app a developer is using.
  const out = path.join(root, 'apps/desktop/build/shadowinput', ...(mode === 'package' ? [version] : []), target);
  const app = path.join(out, 'shadowinput.app');
  const executable = target === 'macos' ? path.join(app, 'Contents/MacOS/shadowinput') : path.join(out, 'shadowinput.exe');
  const resources = target === 'macos' ? path.join(app, 'Contents/Resources') : out;
  if (target === 'macos' && output('/bin/ps', ['-axo', 'comm=']).split('\n').some(line => line.trim() === executable)) {
    throw new Error('Quit this shadowinput build before rebuilding it. The running app has not been modified.');
  }
  rmSync(out, { recursive: true, force: true });
  mkdirSync(path.dirname(executable), { recursive: true });
  mkdirSync(resources, { recursive: true });
  const env = { ...process.env, GOOS: target === 'macos' ? 'darwin' : 'windows', GOARCH: target === 'macos' ? 'arm64' : 'amd64', CGO_ENABLED: target === 'macos' ? '1' : '0' };
  if (target === 'macos') env.MACOSX_DEPLOYMENT_TARGET = '14.0';
  const platform = `${env.GOOS}-${env.GOARCH}`;
  run(process.execPath, ['scripts/stage-llm-runtime.mjs', platform]);
  cpSync(path.join(root, 'apps/desktop/resources', platform, 'llm'), path.join(resources, 'llm'), {recursive: true});
  run(process.execPath, ['scripts/build-shadowinput-ime.mjs', target]);
  if (target === 'macos') cpSync(path.join(root,'apps/desktop/resources/shadowinput/ShadowInput.app'), path.join(resources,'ShadowInput.app'), {recursive:true});
  else {
    cpSync(path.join(root,'apps/desktop/resources/shadowinput/windows/ime'),path.join(resources,'ime'),{recursive:true});
    cpSync(path.join(root,'apps/desktop/resources/windows-amd64/keyboard'),path.join(resources,'keyboard'),{recursive:true});
  }
  run('go', ['build', '-trimpath', ...(target === 'windows' ? ['-ldflags=-H=windowsgui'] : []), '-o', executable, source], env);
  copyFileSync(path.join(root, 'LICENSE'), path.join(resources, 'LICENSE.txt'));
  copyFileSync(path.join(root, 'docs/project/shadowinput.md'), path.join(resources, 'SOURCE-AND-NOTICES.md'));
  const licenses = path.join(resources, 'licenses');
  const modules = output('go', ['list', '-deps', '-f', '{{with .Module}}{{.Path}}\t{{.Version}}\t{{.Dir}}{{end}}', source, './apps/desktop/cmd/arion-llm'], env);
  stageGoModuleNotices(modules, licenses);
  const goroot = output('go', ['env', 'GOROOT'], env);
  copyFileSync(path.join(goroot, 'LICENSE'), path.join(licenses, 'Go-LICENSE.txt'));
  copyFileSync(path.join(goroot, 'PATENTS'), path.join(licenses, 'Go-PATENTS.txt'));
  const icons = path.join(root, 'apps/desktop/ui/shared/icons/svg');
  for (const file of readdirSync(icons).filter(file => file.startsWith('LICENSE'))) {
    copyFileSync(path.join(icons, file), path.join(licenses, `icons-${file}`));
  }
  if (target === 'windows') copyFileSync(path.join(root, 'scripts/installers/NSIS-COPYING.txt'), path.join(licenses, 'NSIS-COPYING.txt'));
  const fonts=desktopFonts(root);
  copyFileSync(path.join(fonts.directory,fonts.manifest.license),path.join(licenses,'SourceHanSans-OFL.txt'));
  copyFileSync(path.join(fonts.directory,'manifest.json'),path.join(licenses,'source-han-sans.json'));
  const buildInfo = { product: 'shadowinput', version, revision, dirty, target, arch: env.GOARCH, go: output('go', ['env', 'GOVERSION'], env), source: `${repository}/tree/${revision}`, license: 'AGPL-3.0', localRuntime: JSON.parse(readFileSync(path.join(root, 'scripts/llm-runtime.json'), 'utf8')).platforms[platform], localModel: 'qwen3.5-2b-q4-k-m-metal', weightsBundled: false };
  writeFileSync(path.join(resources, 'BUILD-INFO.json'), `${JSON.stringify(buildInfo, null, 2)}\n`);
  writeFileSync(path.join(out, 'README.txt'), `shadowinput ${version} — development preview / 开发预览版

${target === 'macos' ? 'macOS 14+ / Apple silicon (arm64). Move shadowinput.app to Applications.\n仅支持 Apple 芯片。打开 DMG（或解压 ZIP），将 shadowinput.app 移到 Applications。\nAd hoc signed; no Developer ID signing or Apple notarization.\n本包尚无 Developer ID 签名或公证，系统可能阻止打开。\nSee https://support.apple.com/102445 for Apple\'s app-opening guidance.' : 'Windows x64. Run the setup EXE; open shadowinput from the Start menu. For the portable ZIP, extract the entire folder and launch shadowinput.exe.\n运行 setup.exe 安装后，从开始菜单打开 shadowinput。便携 ZIP 请完整解压后运行 shadowinput.exe。\nNo publisher signature. Native Windows behavior is not yet verified.\n本包尚无发布者签名，Windows 原生运行仍待验收。'}

Install the system input method from Input method settings, then switch to shadowinput in the OS input menu.
Experimental IME: installed macOS selection/typing and Windows native behavior have not passed acceptance. API registration success is not proof of enablement.
实验性输入法：macOS 的系统切换、实际打字以及 Windows 原生使用尚未通过验收；注册接口成功不等于系统已启用。
Type Pinyin directly in your editor. Space/numbers select ordinary candidates; Tab accepts the separate AI candidate.
Automatic AI candidates require explicit enablement and the desktop app running with a ready model.
在“输入法设置”安装 / 启用后，从系统输入菜单切换到 shadowinput。
直接输入拼音：空格 / 数字选普通候选；Tab 采纳 AI 候选；Esc 取消。
AI 候选默认关闭；开启后保持主程序和模型可用。ChatGPT 会接收当前组合短句及风格、词表，并消耗套餐用量。
长文工具仍支持：选中文字 → Cmd/Ctrl+Shift+J → 生成 → 审阅 → 应用修改。
If selection is unavailable, use Paste original or type a draft; copy the result manually.
无法读取选区时，请用“粘贴原文”或手动输入，完成后复制结果。
Flagged changes require review. Styles, terms and task defaults stay on this device.
数字、日期、词表及承诺的可疑变化需要确认；风格、词表和操作默认值保存在本机。
macOS selection and Apply changes require Accessibility permission.
macOS 的选区读取和“应用修改”需要辅助功能权限。
Local model: choose Local model, then Download & start (1.28 GB). Subsequent Start model works offline.
本地模型：选择“本地模型”后“下载并启动”（1.28 GB）；此后可离线启动。
Qwen3.5-2B Q4_K_M runs on this computer; no ChatGPT account or API key is needed.
本机运行 Qwen3.5-2B Q4_K_M，无需 ChatGPT 账号或 API Key。
One real ChatGPT account and candidate request were verified on macOS. Other accounts and installed Windows TSF behavior remain unverified.
macOS 已验证一个 ChatGPT 账号和候选请求；其他账号及 Windows 已安装输入法行为仍待验收。

Exact source / 对应源码: ${repository}/tree/${revision}
Release / 发布说明: ${repository}/releases/tag/${version}
Build metadata and licenses are ${target === 'macos' ? 'inside shadowinput.app/Contents/Resources' : 'alongside this file'}.
Based on AstraFoundry/Arion. Distributed under AGPL-3.0; upstream and dependency notices retained.
`);
  if (target === 'macos') {
    writeFileSync(path.join(app, 'Contents/Info.plist'), `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleName</key><string>shadowinput</string>
<key>CFBundleDisplayName</key><string>shadowinput</string>
<key>CFBundleExecutable</key><string>shadowinput</string>
<key>CFBundleIdentifier</key><string>org.shadowinput.desktop.dev</string>
<key>CFBundlePackageType</key><string>APPL</string>
<key>CFBundleShortVersionString</key><string>${version.slice(1).split('-')[0]}</string>
<key>CFBundleVersion</key><string>${version.match(/preview\.(\d+)$/)?.[1] ?? "1"}</string>
<key>LSMinimumSystemVersion</key><string>14.0</string>
<key>NSHighResolutionCapable</key><true/>
</dict></plist>\n`);
    const llm = path.join(resources, 'llm');
    for (const file of readdirSync(llm).filter(name => name.endsWith('.dylib') || ['llama-server', 'arion-llm'].includes(name))) {
      run('/usr/bin/codesign', ['--force', '--sign', '-', path.join(llm,file)]);
    }
    run('/usr/bin/codesign', ['--force', '--sign', '-', app]);
    run('/usr/bin/codesign', ['--verify', '--deep', '--strict', app]);
  }
  if (target === 'windows') verifyWindowsPEDependencies([executable, ...['arion-llm.exe','llama-server.exe','llama-server-avx2.exe'].map(file=>path.join(resources,'llm',file))]);
  console.log(`Built ${executable}`);
  if (mode === 'package') {
    const releaseDirectory = path.join(root, 'apps/desktop/build/shadowinput/releases');
    mkdirSync(releaseDirectory, { recursive: true });
    const archive = path.join(releaseDirectory, `shadowinput-${version}-${target}-${target === 'macos' ? 'arm64' : 'x64'}.zip`);
    rmSync(archive, { force: true });
    if (process.platform === 'darwin') run('/usr/bin/ditto', ['-c', '-k', '--keepParent', out, archive]);
    else if (process.platform === 'win32') run('tar.exe', ['-a', '-c', '-f', archive, '-C', path.dirname(out), path.basename(out)]);
    else throw new Error('Package archives on macOS or Windows');
    const installer = createInstaller({ root, target, version, payload: out, releaseDirectory });
    for (const file of [archive, installer]) {
      const digest = createHash('sha256').update(readFileSync(file)).digest('hex');
      writeFileSync(`${file}.sha256`, `${digest}  ${path.basename(file)}\n`);
      console.log(`Packaged ${file}\nSHA256 ${digest}`);
    }
  }
} else {
  throw new Error('Use check, dev, build [macos|windows], or package [macos|windows] <tag>');
}
