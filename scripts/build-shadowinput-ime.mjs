// Reuses Arion's native IME and engine while isolating product registration/data.
import {execFileSync} from 'node:child_process';
import {createHash} from 'node:crypto';
import {cpSync,copyFileSync,existsSync,mkdirSync,readFileSync,writeFileSync,readdirSync,rmSync,renameSync} from 'node:fs';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import {buildShadowWindowsIME} from './lib/build-shadowinput-windows-ime.mjs';
import {buildIMELearning} from './lib/build-ime-learning.mjs';
const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),'..');
const brandAssets=path.join(root,'apps/desktop/ui/shared/shadowbrand');
const menuIcon=`shadowinput-menu-${createHash('sha256').update(readFileSync(path.join(brandAssets,'mark.svg'))).digest('hex').slice(0,12)}.tiff`;
const target=process.argv[2]??'macos';
if(!['macos','windows'].includes(target))throw new Error('Choose macos or windows');
const runtime=path.join(root,'apps/desktop/resources',target==='macos'?'darwin':'windows-amd64','keyboard','runtime.json');
if(!existsSync(runtime))throw new Error(`Stage the keyboard runtime first: node scripts/${target==='macos'?'stage-keyboard-runtime.mjs':'stage-keyboard-runtime-windows.mjs'}. See docs/project/shadowinput-ime.md for cross-compilation prerequisites.`);
const work=path.join(root,'.workspace/shadowinput-ime',target);
const sourceRoot=path.join(work,'source');
const source=path.join(sourceRoot,'apps/desktop/ime');
mkdirSync(source,{recursive:true});
cpSync(path.join(root,'apps/desktop/ime'),source,{recursive:true});
mkdirSync(path.join(sourceRoot,'apps/desktop/infrastructure/imebridge'),{recursive:true});
copyFileSync(path.join(root,'apps/desktop/infrastructure/imebridge/peer_darwin.h'),path.join(sourceRoot,'apps/desktop/infrastructure/imebridge/peer_darwin.h'));
function brand(dir) {
 for(const ent of readdirSync(dir,{withFileTypes:true})) {
  const file=path.join(dir,ent.name);
  if(ent.isDirectory()) {brand(file);continue;}
  if(!/\.(m|mm|h|cpp|plist)$/.test(file))continue;
  let text=readFileSync(file,'utf8');
  text=text.replaceAll('si.astrafoundry.inputmethod.arion','org.shadowinput.inputmethod.shadowinput')
   .replaceAll('si.astrafoundry.arion','org.shadowinput')
   .replaceAll('ArionInput_Connection','ShadowInput_Connection')
   .replaceAll('ARION_DATA_DIR','SHADOWINPUT_DATA_DIR')
   .replaceAll('Arion.Voice.TSF','ShadowInput.TSF').replaceAll('Arion.Native','ShadowInput.Native').replaceAll('arion://','shadowinput://')
   .replaceAll('"Arion"','"shadowinput"').replaceAll('\\\\Arion\\\\','\\\\shadowinput\\\\');
  text=text.replace('0x0cf9199c,0xbc90,0x44f0,{0x93,0x31,0x47,0x4f,0xa6,0x74,0x22,0x71}', '0xe150a8e4,0x9e39,0x4d80,{0xbf,0xa6,0x40,0xe4,0x7e,0xc9,0x75,0x39}')
    .replace('0x78fbf1a9,0x3b48,0x4b37,{0xa5,0x2c,0xb7,0x3d,0x9d,0x31,0x22,0x63}', '0x09def652,0x8bdd,0x4ea7,{0xae,0x89,0xe4,0x4a,0x63,0x83,0x47,0x25}');
  writeFileSync(file,text);
 }
}
brand(source);
const run=(cmd,args,cwd=root)=>execFileSync(cmd,args,{cwd,stdio:'inherit'});
if(target==='windows') {buildShadowWindowsIME({root,source,work,run});process.exit(0);}
if(target!=='macos')throw new Error('Choose macos or windows');
if(process.platform!=='darwin')throw new Error('macOS IME requires macOS');
const keyboard=path.join(root,'apps/desktop/resources/darwin/keyboard');
// Finish the hidden bundle before exposing a .app that LaunchServices can index.
const outputBundle=path.join(root,'apps/desktop/resources/shadowinput/ShadowInput.app');
const bundle=path.join(work,'ShadowInput.app');
rmSync(bundle,{recursive:true,force:true});
for(const dir of ['MacOS','Resources','Frameworks'])mkdirSync(path.join(bundle,'Contents',dir),{recursive:true});
let plist=readFileSync(path.join(source,'macos/Info.plist'),'utf8').replaceAll('<string>Arion</string>','<string>shadowinput</string>');
// macOS can default an input mode to a template even when the parent is not.
// Preserve the yellow SI artwork at both levels and invalidate old icon URLs.
plist=plist.replaceAll('menu_icon.tiff',menuIcon)
 .replaceAll('<key>TISInputSourceID</key>','<key>TISIconIsTemplate</key><false/>\n        <key>TISInputSourceID</key>');
plist=plist.replace('</dict>\n</plist>','<key>TISIconIsTemplate</key><false/>\n</dict>\n</plist>');
plist=plist.replace('</dict>\n</plist>','<key>ArionAppIdentifier</key><string>org.shadowinput.desktop.dev</string>\n</dict>\n</plist>');
plist=plist.replace('</dict>\n</plist>','<key>CFBundleIconFile</key><string>shadowinput.icns</string>\n</dict>\n</plist>');
writeFileSync(path.join(bundle,'Contents/Info.plist'),plist);
const objects=path.join(work,'objects');mkdirSync(objects,{recursive:true});
const native=[];
for(const name of ['main','candidate_panel','composition','keyboard','ranking','typing']) {
 const obj=path.join(objects,name+'.o');
 run('xcrun',['clang','-fobjc-arc','-O2','-DSHADOWINPUT_IME','-target','arm64-apple-macos14.0','-I',path.join(keyboard,'include'),'-c',path.join(source,'macos',name+'.m'),'-o',obj]);native.push(obj);
}
const learning=buildIMELearning({root:sourceRoot,outputDir:objects,arch:'arm64',minimumSystemVersion:'14.0',sdkInclude:path.join(root,'.workspace/keyboard-runtime/sdk/include'),defines:['-DSHADOWINPUT_IME']});
run('xcrun',[learning.linker,'-target','arm64-apple-macos14.0',...native,...learning.objects,path.join(keyboard,'lib/librime.1.dylib'),...learning.linkArgs,'-Xlinker','-lbsm','-framework','Cocoa','-framework','QuartzCore','-framework','InputMethodKit','-framework','Carbon','-Xlinker','-rpath','-Xlinker','@executable_path/../Frameworks','-o',path.join(bundle,'Contents/MacOS/ArionInput')]);
copyFileSync(path.join(keyboard,'lib/librime.1.dylib'),path.join(bundle,'Contents/Frameworks/librime.1.dylib'));
for(const dir of ['data','licenses'])cpSync(path.join(keyboard,dir),path.join(bundle,'Contents/Resources/keyboard',dir),{recursive:true});
copyFileSync(path.join(keyboard,'runtime.json'),path.join(bundle,'Contents/Resources/keyboard/runtime.json'));
for(const [locale,language] of [['zh-CN','zh-Hans'],['en','en']]) {
 const catalog=JSON.parse(readFileSync(path.join(root,'apps/desktop/ui/shared/copy/locales',locale+'.json'),'utf8'));
 const dir=path.join(bundle,'Contents/Resources',language+'.lproj');mkdirSync(dir,{recursive:true});
 const lines=Object.entries(catalog).filter(([k])=>k.startsWith('Keyboard')||k.startsWith('HoldKey')).map(([k,v])=>`${JSON.stringify(k)} = ${JSON.stringify(v.replaceAll('Arion','shadowinput'))};`);
 writeFileSync(path.join(dir,'Localizable.strings'),lines.join('\n')+'\n');
 writeFileSync(path.join(dir,'InfoPlist.strings'),'"CFBundleName"="shadowinput";\n"CFBundleDisplayName"="shadowinput";\n"org.shadowinput.inputmethod.shadowinput.pinyin"="shadowinput";\n');
}
copyFileSync(path.join(brandAssets,'shadowinput.icns'),path.join(bundle,'Contents/Resources/shadowinput.icns'));
const menuSlices=[16,32].map(size=>{
 const tiff=path.join(objects,`menu-${size}.tiff`);
 run('sips',['-s','format','tiff',path.join(brandAssets,`icon-${size}.png`),'--out',tiff]);
 return tiff;
});
run('tiffutil',['-cathidpicheck',...menuSlices,'-out',path.join(bundle,'Contents/Resources',menuIcon)]);
for(const file of [path.join(bundle,'Contents/Frameworks/librime.1.dylib'),bundle])run('/usr/bin/codesign',['--force','--sign','-',file]);
run('/usr/bin/codesign',['--verify','--deep','--strict',bundle]);
mkdirSync(path.dirname(outputBundle),{recursive:true});
rmSync(outputBundle,{recursive:true,force:true});
renameSync(bundle,outputBundle);
console.log(`Built ${outputBundle}`);
