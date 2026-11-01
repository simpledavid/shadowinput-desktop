import { copyFileSync, mkdirSync, readdirSync, writeFileSync } from 'node:fs';
import path from 'node:path';

const licenseName = /^(LICENSE|LICENCE|COPYING|NOTICE)([._-].*)?$/i;

function licenseFiles(directory, prefix = '') {
  return readdirSync(path.join(directory, prefix), { withFileTypes: true }).flatMap(entry => {
    const relative = path.join(prefix, entry.name);
    if (entry.isDirectory() && !entry.name.startsWith('.')) return licenseFiles(directory, relative);
    return entry.isFile() && licenseName.test(entry.name) ? [relative] : [];
  }).sort();
}

// Accept tab-separated output from go list -m or the linked package closure.
export function stageGoModuleNotices(modules, destination) {
  const records = [];
  const seen = new Set();
  for (const row of modules.trim().split('\n')) {
    const [name, version, directory] = row.split('\t');
    if (!name || name.startsWith('arion.local/') || seen.has(name)) continue;
    seen.add(name);
    const licenses = licenseFiles(directory);
    if (!licenses.some(file => !file.includes(path.sep))) throw new Error(`No root license found for ${name}@${version}`);
    const folder = name.replaceAll('/', '_');
    for (const file of licenses) {
      const target = path.join(destination, folder, file);
      mkdirSync(path.dirname(target), { recursive: true });
      copyFileSync(path.join(directory, file), target);
    }
    records.push({ module: name, version, ...(name === 'github.com/egoist/mygo' ? { source: 'third_party/mygo', patch: 'prerelease-channel-v1, url-channel-v1, staged-update-v1, native-notifications-v1, codesign-metadata-v1, windows-native-transparency-v1, native-clipboard-writer-v1, screen-anchored-flyout-v1' } : {}), licenses: licenses.map(file => `${folder}/${file.split(path.sep).join('/')}`) });
  }
  records.sort((a, b) => a.module.localeCompare(b.module));
  mkdirSync(destination, { recursive: true });
  writeFileSync(path.join(destination, 'go-modules.json'), `${JSON.stringify(records, null, 2)}\n`);
  return records;
}
