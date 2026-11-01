import { execFileSync, spawnSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { mkdirSync, copyFileSync, readFileSync, readdirSync, rmSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { stageGoModuleNotices } from './lib/go-module-notices.mjs';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const mode = process.argv[2] ?? 'check';
const nativeTarget = process.platform === 'darwin' ? 'macos' : process.platform === 'win32' ? 'windows' : '';
const target = process.argv[3] ?? nativeTarget;
const source = './apps/desktop/cmd/shadowinput';
const repository = 'https://github.com/simpledavid/shadowinput-desktop';
function output(command, args, env = process.env) {
  return execFileSync(command, args, { cwd: root, env, encoding: 'utf8' }).trim();
}
function run(command, args, env = process.env) {
  const result = spawnSync(command, args, { cwd: root, env, stdio: 'inherit' });
  if (result.error) throw result.error;
  if (result.status !== 0) process.exit(result.status ?? 1);
}
if (mode === 'check') {
  run('go', ['test', '-race', './packages/core/...', './apps/desktop/infrastructure/chatgptplan', './apps/desktop/infrastructure/writingprefs', './apps/desktop/ui/shared/writing', './apps/desktop/internal/oauthloopback', './apps/desktop/internal/shadowinput', './apps/desktop/ui/features/writing', './apps/desktop/ui/shared/copy']);
  run('go', ['vet', source, './apps/desktop/infrastructure/chatgptplan', './apps/desktop/internal/shadowinput']);
  run(process.execPath, ['scripts/check-repository.mjs']);
} else if (mode === 'dev') {
  run('go', ['run', source]);
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
  const out = path.join(root, 'apps/desktop/build/shadowinput', target);
  const app = path.join(out, 'shadowinput.app');
  const executable = target === 'macos' ? path.join(app, 'Contents/MacOS/shadowinput') : path.join(out, 'shadowinput.exe');
  const resources = target === 'macos' ? path.join(app, 'Contents/Resources') : out;
  rmSync(out, { recursive: true, force: true });
  mkdirSync(path.dirname(executable), { recursive: true });
  mkdirSync(resources, { recursive: true });
  const env = { ...process.env, GOOS: target === 'macos' ? 'darwin' : 'windows', GOARCH: target === 'macos' ? 'arm64' : 'amd64', CGO_ENABLED: target === 'macos' ? '1' : '0' };
  if (target === 'macos') env.MACOSX_DEPLOYMENT_TARGET = '14.0';
  run('go', ['build', '-trimpath', ...(target === 'windows' ? ['-ldflags=-H=windowsgui'] : []), '-o', executable, source], env);
  copyFileSync(path.join(root, 'LICENSE'), path.join(resources, 'LICENSE.txt'));
  copyFileSync(path.join(root, 'docs/project/shadowinput.md'), path.join(resources, 'SOURCE-AND-NOTICES.md'));
  const licenses = path.join(resources, 'licenses');
  const modules = output('go', ['list', '-deps', '-f', '{{with .Module}}{{.Path}}\t{{.Version}}\t{{.Dir}}{{end}}', source], env);
  stageGoModuleNotices(modules, licenses);
  const goroot = output('go', ['env', 'GOROOT'], env);
  copyFileSync(path.join(goroot, 'LICENSE'), path.join(licenses, 'Go-LICENSE.txt'));
  copyFileSync(path.join(goroot, 'PATENTS'), path.join(licenses, 'Go-PATENTS.txt'));
  const icons = path.join(root, 'apps/desktop/ui/shared/icons/svg');
  for (const file of readdirSync(icons).filter(file => file.startsWith('LICENSE'))) {
    copyFileSync(path.join(icons, file), path.join(licenses, `icons-${file}`));
  }
  const buildInfo = { product: 'shadowinput', version, revision, dirty, target, arch: env.GOARCH, go: output('go', ['env', 'GOVERSION'], env), source: `${repository}/tree/${revision}`, license: 'AGPL-3.0' };
  writeFileSync(path.join(resources, 'BUILD-INFO.json'), `${JSON.stringify(buildInfo, null, 2)}\n`);
  writeFileSync(path.join(out, 'README.txt'), `shadowinput ${version} — development preview / 开发预览版

${target === 'macos' ? 'macOS 14+ / Apple silicon (arm64). Move shadowinput.app to Applications.\n仅支持 Apple 芯片。解压后将 shadowinput.app 移到 Applications。\nAd hoc signed; no Developer ID signing or Apple notarization.\n本包尚无 Developer ID 签名或公证，系统可能阻止打开。\nSee https://support.apple.com/102445 for Apple\'s app-opening guidance.' : 'Windows x64. Extract the entire folder and launch shadowinput.exe.\n解压整个文件夹后运行 shadowinput.exe。\nNo publisher signature. Native Windows behavior is not yet verified.\n本包尚无发布者签名，Windows 原生运行仍待验收。'}

Select text in a supported editor, press Cmd/Ctrl+Shift+J, generate, review, then apply.
在支持的编辑器中选中文字 → Cmd/Ctrl+Shift+J → 生成 → 审阅 → 应用修改。
If selection is unavailable, use Paste original or type a draft; copy the result manually.
无法读取选区时，请用“粘贴原文”或手动输入，完成后复制结果。
Flagged changes require review. Styles, terms and task defaults stay on this device.
数字、日期、词表及承诺的可疑变化需要确认；风格、词表和操作默认值保存在本机。
macOS selection and Apply changes require Accessibility permission.
macOS 的选区读取和“应用修改”需要辅助功能权限。
Real ChatGPT login, account eligibility and inference still need acceptance testing.
真实 ChatGPT 登录、账号资格与推理仍待验收。

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
    run('/usr/bin/codesign', ['--force', '--sign', '-', app]);
    run('/usr/bin/codesign', ['--verify', '--strict', app]);
  }
  console.log(`Built ${executable}`);
  if (mode === 'package') {
    const releaseDirectory = path.join(root, 'apps/desktop/build/shadowinput/releases');
    mkdirSync(releaseDirectory, { recursive: true });
    const archive = path.join(releaseDirectory, `shadowinput-${version}-${target}-${target === 'macos' ? 'arm64' : 'x64'}.zip`);
    rmSync(archive, { force: true });
    if (process.platform === 'darwin') run('/usr/bin/ditto', ['-c', '-k', '--keepParent', out, archive]);
    else if (process.platform === 'win32') run('tar.exe', ['-a', '-c', '-f', archive, '-C', path.dirname(out), path.basename(out)]);
    else throw new Error('Package archives on macOS or Windows');
    const digest = createHash('sha256').update(readFileSync(archive)).digest('hex');
    writeFileSync(`${archive}.sha256`, `${digest}  ${path.basename(archive)}\n`);
    console.log(`Packaged ${archive}\nSHA256 ${digest}`);
  }
} else {
  throw new Error('Use check, dev, build [macos|windows], or package [macos|windows] <tag>');
}
