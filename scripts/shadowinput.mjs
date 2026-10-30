import { spawnSync } from 'node:child_process';
import { mkdirSync, copyFileSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const mode = process.argv[2] ?? 'check';
const nativeTarget = process.platform === 'darwin' ? 'macos' : process.platform === 'win32' ? 'windows' : '';
const target = process.argv[3] ?? nativeTarget;
const source = './apps/desktop/cmd/shadowinput';
function run(command, args, env = process.env) {
  const result = spawnSync(command, args, { cwd: root, env, stdio: 'inherit' });
  if (result.error) throw result.error;
  if (result.status !== 0) process.exit(result.status ?? 1);
}
if (mode === 'check') {
  run('go', ['test', '-race', './packages/core/...', './apps/desktop/infrastructure/chatgptplan', './apps/desktop/internal/oauthloopback', './apps/desktop/internal/shadowinput', './apps/desktop/ui/features/writing', './apps/desktop/ui/shared/copy']);
  run('go', ['vet', source, './apps/desktop/infrastructure/chatgptplan', './apps/desktop/internal/shadowinput']);
  run(process.execPath, ['scripts/check-repository.mjs']);
} else if (mode === 'dev') {
  run('go', ['run', source]);
} else if (mode === 'build') {
  if (!['macos', 'windows'].includes(target)) throw new Error('Choose macos or windows');
  if (target === 'macos' && process.platform !== 'darwin') throw new Error('macOS builds require a Mac');
  const out = path.join(root, 'apps/desktop/build/shadowinput', target);
  const app = path.join(out, 'shadowinput.app');
  const executable = target === 'macos' ? path.join(app, 'Contents/MacOS/shadowinput') : path.join(out, 'shadowinput.exe');
  const resources = target === 'macos' ? path.join(app, 'Contents/Resources') : out;
  mkdirSync(path.dirname(executable), { recursive: true });
  mkdirSync(resources, { recursive: true });
  const env = { ...process.env, GOOS: target === 'macos' ? 'darwin' : 'windows', GOARCH: target === 'macos' ? 'arm64' : 'amd64', CGO_ENABLED: target === 'macos' ? '1' : '0' };
  if (target === 'macos') env.MACOSX_DEPLOYMENT_TARGET = '14.0';
  run('go', ['build', '-trimpath', ...(target === 'windows' ? ['-ldflags=-H=windowsgui'] : []), '-o', executable, source], env);
  copyFileSync(path.join(root, 'LICENSE'), path.join(resources, 'LICENSE.txt'));
  copyFileSync(path.join(root, 'docs/project/shadowinput.md'), path.join(resources, 'SOURCE-AND-NOTICES.md'));
  copyFileSync(path.join(root, 'third_party/mygo/LICENSE'), path.join(resources, 'MyGo-LICENSE.txt'));
  if (target === 'macos') {
    writeFileSync(path.join(app, 'Contents/Info.plist'), `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleName</key><string>shadowinput</string>
<key>CFBundleDisplayName</key><string>shadowinput</string>
<key>CFBundleExecutable</key><string>shadowinput</string>
<key>CFBundleIdentifier</key><string>org.shadowinput.desktop.dev</string>
<key>CFBundlePackageType</key><string>APPL</string>
<key>CFBundleShortVersionString</key><string>0.1.0</string>
<key>CFBundleVersion</key><string>1</string>
<key>LSMinimumSystemVersion</key><string>14.0</string>
<key>NSHighResolutionCapable</key><true/>
</dict></plist>\n`);
    run('/usr/bin/codesign', ['--force', '--sign', '-', app]);
    run('/usr/bin/codesign', ['--verify', '--strict', app]);
  }
  console.log(`Built ${executable}`);
} else {
  throw new Error('Use check, dev, or build [macos|windows]');
}
