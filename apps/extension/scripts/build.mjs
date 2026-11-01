import {cp, mkdir, readFile, readdir, writeFile, rm} from 'node:fs/promises';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import {createHash} from 'node:crypto';
import {zipSync} from 'fflate';
import {build} from 'esbuild';
import {execFileSync} from 'node:child_process';

const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),'..');
const destination=path.join(root,'dist','shadowinput');
await rm(path.join(root,'dist'),{recursive:true,force:true});
await mkdir(destination,{recursive:true});
await cp(path.join(root,'build'),destination,{recursive:true});
await cp(path.join(root,'public'),destination,{recursive:true});
await build({entryPoints:[path.join(root,'src/app/content.ts')],bundle:true,format:'iife',target:'chrome120',outfile:path.join(destination,'content.js'),legalComments:'none'});
await cp(path.join(root,'..','..','LICENSE'),path.join(destination,'LICENSE'));
await cp(path.join(root,'INSTALL.txt'),path.join(destination,'INSTALL.txt'));
const revision=execFileSync('git',['rev-parse','HEAD'],{cwd:root,encoding:'utf8'}).trim();
const dirty=execFileSync('git',['status','--porcelain'],{cwd:root,encoding:'utf8'}).trim()!=='';
const manifest=JSON.parse(await readFile(path.join(destination,'manifest.json'),'utf8'));
await writeFile(path.join(destination,'SOURCE.txt'),'AGPL-3.0-only\nSource: https://github.com/simpledavid/shadowinput-desktop/tree/'+revision+'/apps/extension\n');
await writeFile(path.join(destination,'BUILD-INFO.json'),JSON.stringify({product:'shadowinput',version:manifest.version_name,revision,dirty},null,2)+'\n');
const entries={};
async function collect(directory,relative='') {
  for(const item of (await readdir(directory,{withFileTypes:true})).sort((a,b)=>a.name.localeCompare(b.name))) {
    const name=path.posix.join(relative,item.name), absolute=path.join(directory,item.name);
    if(item.isDirectory()) await collect(absolute,name);
    else entries['shadowinput/'+name]=[new Uint8Array(await readFile(absolute)),{mtime:new Date('2026-01-01T00:00:00Z')}];
  }
}
await collect(destination);
const archive=zipSync(entries,{level:9});
const filename='shadowinput-chrome-'+manifest.version_name+'.zip';
await writeFile(path.join(root,'dist',filename),archive);
await writeFile(path.join(root,'dist',filename+'.sha256'),createHash('sha256').update(archive).digest('hex')+'  '+filename+'\n');
console.log('Built '+filename+' ('+archive.length+' bytes)');
