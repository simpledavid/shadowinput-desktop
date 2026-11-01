import {execFileSync} from 'node:child_process';
import {mkdtempSync,mkdirSync,rmSync} from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import {buildIMETypingFixture} from './lib/build-ime-typing-fixture.mjs';
const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),'..');
const temporary=mkdtempSync(path.join(os.tmpdir(),'shadowinput-ime-test-'));
try {
 const binary=buildIMETypingFixture({root,outputDir:temporary,arch:'arm64',fixture:'tests/shadow_candidate_test.mm',defines:['SHADOWINPUT_IME']});
 const user=path.join(temporary,'user');mkdirSync(user);
 execFileSync(binary,[path.join(root,'apps/desktop/resources/darwin/keyboard/data'),user],{stdio:'inherit'});
} finally {rmSync(temporary,{recursive:true,force:true});}
