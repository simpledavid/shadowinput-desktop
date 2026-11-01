import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, rmSync, symlinkSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { installerInventory, windowsInventoryScript } from './shadowinput-installers.mjs';

test('uninstall inventory names only owned files and removes nested directories last', () => {
  const root = mkdtempSync(path.join(tmpdir(), 'shadowinput-inventory-'));
  try {
    mkdirSync(path.join(root, 'llm/licenses'), { recursive: true });
    for (const name of ['shadowinput.exe', 'llm/llama-server.exe', 'llm/licenses/LICENSE.txt']) writeFileSync(path.join(root, name), name);
    const inventory = installerInventory(root);
    assert.deepEqual(inventory.directories, ['llm/licenses', 'llm']);
    const script = windowsInventoryScript(inventory);
    assert.equal((script.match(/  Delete /g) ?? []).length, 3);
    assert.equal((script.match(/Call \$\{PREFIX\}CheckFileUnlocked/g) ?? []).length, 2);
    assert.match(script, /Delete "\$INSTDIR\\llm\\licenses\\LICENSE.txt"/);
    assert.doesNotMatch(script, /RMDir \/r|\$APPDATA|\*|\?/i);
    assert.ok(script.indexOf('RMDir "$INSTDIR\\llm\\licenses"') < script.indexOf('RMDir "$INSTDIR\\llm"'));
  } finally { rmSync(root, { recursive: true, force: true }); }
});

test('uninstall inventory rejects links and NSIS/wildcard filename injection', () => {
  const root = mkdtempSync(path.join(tmpdir(), 'shadowinput-inventory-'));
  try {
    writeFileSync(path.join(root, '$APPDATA'), 'unsafe');
    assert.throws(() => installerInventory(root), /Unsafe installer filename/);
    rmSync(path.join(root, '$APPDATA'));
    writeFileSync(path.join(root, 'all*'), 'unsafe');
    assert.throws(() => installerInventory(root), /Unsafe installer filename/);
    rmSync(path.join(root, 'all*'));
    symlinkSync(tmpdir(), path.join(root, 'outside'));
    assert.throws(() => installerInventory(root), /only regular files/);
  } finally { rmSync(root, { recursive: true, force: true }); }
});
