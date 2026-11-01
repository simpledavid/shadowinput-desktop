// Preserve licenses of the pinned Go modules and the model provenance in packages.
import { execFileSync } from 'node:child_process';
import { copyFileSync, mkdirSync, readFileSync, rmSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { desktopFonts } from './lib/desktop-fonts.mjs';
import { pinnedFile } from './lib/pinned-file.mjs';
import { stageGoModuleNotices } from './lib/go-module-notices.mjs';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const desktop = path.join(root, 'apps', 'desktop');
const destination = path.join(desktop, 'resources', 'licenses');
rmSync(destination, { recursive: true, force: true });
mkdirSync(destination, { recursive: true });
copyFileSync(path.join(root, 'LICENSE'), path.join(destination, 'Arion-LICENSE'));
copyFileSync(path.join(root, 'models', 'catalog.json'), path.join(destination, 'model-catalog.json'));
const env = { ...process.env, GOWORK: 'off' };
execFileSync('go', ['mod', 'download', 'all'], { cwd: desktop, env, stdio: 'inherit' });
const modules = execFileSync('go', ['list', '-m', '-f', '{{.Path}}\t{{.Version}}\t{{.Dir}}', 'all'], { cwd: desktop, env, encoding: 'utf8' });
const records = stageGoModuleNotices(modules, destination);
const native = JSON.parse(readFileSync(path.join(root, 'scripts/native-notices.json'), 'utf8'));
for (const spec of native.files) copyFileSync(pinnedFile(spec), path.join(destination, spec.name));
copyFileSync(path.join(root, 'scripts/native-notices.json'), path.join(destination, 'native-notices.json'));
if (process.platform === 'win32') {
  const { directory, manifest } = desktopFonts(root);
  copyFileSync(path.join(directory, manifest.license), path.join(destination, 'SourceHanSans-OFL.txt'));
  copyFileSync(path.join(directory, 'manifest.json'), path.join(destination, 'source-han-sans.json'));
}
console.log(`Staged Arion, model provenance and ${records.length} module license records.`);
