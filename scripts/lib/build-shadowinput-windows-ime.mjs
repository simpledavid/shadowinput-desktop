import {createHash} from 'node:crypto';
import {mkdirSync,readFileSync,writeFileSync} from 'node:fs';
import path from 'node:path';
import {recordWindowsKeyboardModule,verifyWindowsKeyboardRuntime} from './windows-keyboard.mjs';
import {verifyWindowsPEDependencies} from './windows-pe.mjs';
import {desktopFonts} from './desktop-fonts.mjs';
export function buildShadowWindowsIME({root,source,work,run}) {
 const native=path.join(source,'windows'),shared=path.join(source,'shared');
 const keyboard=path.join(root,'apps/desktop/resources/windows-amd64/keyboard');
 const sdk=path.join(root,'.workspace/keyboard-runtime/windows/sdk');
 const out=path.join(root,'apps/desktop/resources/shadowinput/windows/ime');mkdirSync(out,{recursive:true});
 const objects=path.join(work,'objects');mkdirSync(objects,{recursive:true});
 const cxx=process.env.CXX??(process.platform==='win32'?'g++':'x86_64-w64-mingw32-g++');
 const cc=process.env.CC??(process.platform==='win32'?'gcc':'x86_64-w64-mingw32-gcc');
 const windres=process.env.WINDRES??(process.platform==='win32'?'windres':'x86_64-w64-mingw32-windres');
 const cObjects=[];
 for(const name of ['pinyin','input','menu']){const obj=path.join(objects,name+'.o');run(cc,['-std=c11','-O2','-Wall','-Wextra','-Werror','-I',path.join(sdk,'include'),'-c',path.join(shared,name+'.c'),'-o',obj]);cObjects.push(obj);}
 const base=['-std=c++17','-DINITGUID','-DSHADOWINPUT_IME','-O2','-Wall','-Wextra','-Werror','-static','-isystem',path.join(sdk,'include')];
 const module=path.join(keyboard,'lib/ArionKeyboard.dll');
 run(cxx,[...base,'-DRIME_IMPORTS','-shared',path.join(native,'keyboard_runtime.cpp'),...['learning','learning_rime','learning_windows','ranking','ranking_unicode','emoji','spelling','personal_vocabulary','personal_vocabulary_store','english','english_lexicon'].map(n=>path.join(shared,n+'.cpp')),cObjects[0],path.join(sdk,'lib/libArionRime.dll.a'),'-o',module,'-lcrypt32','-lole32','-luuid','-ladvapi32','-ldwrite','-ld2d1']);
 recordWindowsKeyboardModule(keyboard);verifyWindowsKeyboardRuntime(keyboard);
 const hash=file=>createHash('sha256').update(readFileSync(file)).digest('hex');
 const keyboardManifestSHA256=hash(path.join(keyboard,'runtime.json'));
 const loader=path.join(out,'ArionTSFLoader.dll');
 run(cxx,[...base,'-Wl,--no-insert-timestamp','-Wl,--image-base,0x180000000','-shared',path.join(native,'loader.cpp'),'-o',loader,'-lole32','-luuid','-ladvapi32','-lshell32']);
 const loaderHash=hash(loader),font=path.join(objects,'candidate-fonts.o');
 run(windres,['-I',desktopFonts(root).directory,'-i',path.join(native,'candidate_fonts.rc'),'-O','coff','-o',font]);
 for(const [input,filename,isDLL] of [['service.cpp','ArionTSF.dll',true],['control.cpp','arion-ime.exe',false]]){
  run(cxx,[...base,`-DARION_KEYBOARD_MANIFEST_SHA256="${keyboardManifestSHA256}"`,`-DARION_TSF_LOADER_SHA256="${loaderHash}"`,...(isDLL?['-shared']:['-municode']),path.join(native,input),...(isDLL?[...cObjects,font]:[]),'-o',path.join(out,filename),'-lole32','-loleaut32','-luuid','-ladvapi32','-lshell32','-lgdi32',...(isDLL?['-ldwrite','-ld2d1','-lwinhttp']:[])]);
 }
 const files=Object.fromEntries(['ArionTSFLoader.dll','ArionTSF.dll','arion-ime.exe'].map(name=>[name,hash(path.join(out,name))]));
 verifyWindowsPEDependencies(Object.keys(files).map(n=>path.join(out,n)));
 writeFileSync(path.join(out,'runtime.json'),JSON.stringify({schema:2,product:'shadowinput',architecture:'amd64',keyboardManifestSHA256,files},null,2)+'\n');
 console.log(`Built shadowinput TSF at ${out}`);
}
