import { execFileSync } from 'node:child_process';
import { cpSync, mkdtempSync, readdirSync, readFileSync, rmSync, statSync, symlinkSync, writeFileSync } from 'node:fs';
import path from 'node:path';

// Enumerate owned files so uninstall never recursively deletes a user directory.
export function installerInventory(directory) {
  const files = [], directories = [];
  function visit(relative) {
    for (const entry of readdirSync(path.join(directory, relative), { withFileTypes: true })) {
      if (/[\x00-\x1f$"*?:\\]/.test(entry.name)) throw new Error(`Unsafe installer filename: ${entry.name}`);
      const name = relative ? `${relative}/${entry.name}` : entry.name;
      if (entry.isDirectory()) { directories.push(name); visit(name); }
      else if (entry.isFile()) files.push(name);
      else throw new Error(`Installer payload must contain only regular files: ${name}`);
    }
  }
  visit('');
  return { files: files.sort(), directories: directories.sort((a, b) => b.split('/').length - a.split('/').length || b.localeCompare(a)) };
}

export function windowsInventoryScript({ files, directories }) {
  const win = value => value.replaceAll('/', '\\');
  return [
    '!macro CheckPayload PREFIX',
    ...files.filter(file => /\.(exe|dll)$/i.test(file)).flatMap(file => [`  Push "$INSTDIR\\${win(file)}"`, '  Call ${PREFIX}CheckFileUnlocked']),
    '!macroend',
    '!macro RemovePayload',
    ...files.map(file => `  Delete "$INSTDIR\\${win(file)}"`),
    ...directories.map(directory => `  RMDir "$INSTDIR\\${win(directory)}"`),
    '!macroend',
    '',
  ].join('\n');
}

export function createInstaller({ root, target, version, payload, releaseDirectory }) {
  const destination = path.join(releaseDirectory, `shadowinput-${version}-${target}-${target === 'macos' ? 'arm64.dmg' : 'x64-setup.exe'}`);
  const scratch = mkdtempSync(path.join(releaseDirectory, '.installer-'));
  const run = (command, args) => execFileSync(command, args, { stdio: 'inherit' });
  try {
    if (target === 'macos') {
      const stage = path.join(scratch, 'volume');
      cpSync(payload, stage, { recursive: true });
      symlinkSync('/Applications', path.join(stage, 'Applications'));
      run('/usr/bin/hdiutil', ['create', '-ov', '-volname', 'shadowinput', '-fs', 'HFS+', '-srcfolder', stage, '-format', 'UDZO', destination]);
      run('/usr/bin/hdiutil', ['verify', destination]);
    } else {
      // NSIS is required only when packaging Windows, not for ordinary builds.
      const inventory = installerInventory(payload);
      const manifest = path.join(scratch, 'payload.nsh');
      writeFileSync(manifest, windowsInventoryScript(inventory));
      const numericVersion = version.slice(1).split('-')[0] + '.' + (version.match(/preview\.(\d+)$/)?.[1] ?? '0');
      const defines = { PAYLOAD: payload, OUTPUT: destination, VERSION: version, NUMERIC_VERSION: numericVersion, INVENTORY: manifest,
        INSTALLED_KIB: Math.ceil(inventory.files.reduce((size, file) => size + statSync(path.join(payload, file)).size, 0) / 1024) };
      for (const value of Object.values(defines)) if (/[\x00-\x1f$"]/.test(String(value))) throw new Error('Unsupported character in NSIS build path');
      const prefix = process.platform === 'win32' ? '/' : '-';
      run(process.env.MAKENSIS || 'makensis', [`${prefix}V3`, ...Object.entries(defines).map(([key, value]) => `${prefix}D${key}=${value}`), path.join(root, 'scripts/installers/shadowinput.nsi')]);
      const bytes = readFileSync(destination);
      if (bytes.toString('ascii', 0, 2) !== 'MZ') throw new Error('Installer is not a Windows executable');
    }
    return destination;
  } finally {
    rmSync(scratch, { recursive: true, force: true });
  }
}
