import {execFileSync} from 'node:child_process';
import {mkdirSync} from 'node:fs';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),'..');
const output=path.join(root,'.workspace/tests/shadowinput-candidates');
mkdirSync(output,{recursive:true});
const run=(cmd,args)=>execFileSync(cmd,args,{cwd:root,stdio:'inherit',timeout:60000});
const windows=process.platform==='win32',cross=process.argv.includes('--cross-windows');
const cxx=process.env.CXX??(windows?'g++':'c++');
const source=path.join(root,'apps/desktop/ime/shadowinput');
const state=path.join(output,windows?'candidate-state.exe':'candidate-state');
run(cxx,['-std=c++17','-Wall','-Wextra','-Werror',path.join(source,'candidate_state_test.cpp'),'-o',state]);
run(state,[]);
if(windows || cross) {
 const transport=path.join(output,'candidate-http.exe');
 run(cross?'x86_64-w64-mingw32-g++':cxx,['-std=c++17','-Wall','-Wextra','-Werror','-static',path.join(source,'candidate_http_windows_test.cpp'),'-o',transport,'-lwinhttp','-lws2_32']);
 if(windows)run(transport,[]);
 else console.log('Windows HTTP fixture cross-built; native execution NOT RUN.');
}
