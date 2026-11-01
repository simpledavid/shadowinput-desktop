import {test} from 'node:test';
import assert from 'node:assert/strict';
import {JSDOM} from 'jsdom';
import {paragraphs,originalText,annotate,restore,draft} from '../build/features/reading/index.js';

test('bilingual reading finds main content once, excluding menus, code, hidden text and editable fields',()=>{
  const {document}=new JSDOM('<nav><p>Navigation label</p></nav><main><div><p>Hello <strong>world</strong>.<span hidden>secret</span></p><p>Second paragraph.</p></div><pre><p>Sample code</p></pre><div contenteditable>Private draft</div><textarea>Private field</textarea><p style="display:none">Hidden paragraph</p></main>').window;
  assert.deepEqual(paragraphs(document).map(originalText),['Hello world.','Second paragraph.']);
});
test('translation is inert text and restoring it leaves original formatting and event handlers intact',()=>{
  const {document}=new JSDOM('<main><p>Hello <a href="/link">world</a>.</p></main>').window;
  const p=document.querySelector('p'),link=document.querySelector('a'),before=originalText(p);
  assert.equal(annotate(p,before,'<script>bad()</script>'),true);
  assert.equal(originalText(p),before);
  assert.equal(p.querySelector('script'),null);
  assert.equal(document.querySelector('a'),link);
  assert.equal(p.querySelector('[data-shadowinput]').shadowRoot.textContent.includes('<script>bad()</script>'),true);
  restore(document);
  assert.equal(p.querySelector('[data-shadowinput]'),null);
  assert.equal(document.querySelector('a'),link);
});
test('a changed or disconnected paragraph rejects its stale translation',()=>{
  const {document}=new JSDOM('<main><p>Hello</p></main>').window;
  const p=document.querySelector('p'),before=originalText(p);
  p.textContent='Changed';
  assert.equal(annotate(p,before,'translation'),false);
  const updated=originalText(p);p.remove();
  assert.equal(annotate(p,updated,'translation'),false);
});
test('input translation applies only after review to an unchanged supported field and never submits',()=>{
  const {document}=new JSDOM('<form><textarea>Hello</textarea><input type="password" value="secret"><div contenteditable="true"><b>Rich text</b></div></form>').window;
  const field=document.querySelector('textarea');
  let inputs=0,submits=0;
  field.addEventListener('input',()=>inputs++);
  document.querySelector('form').addEventListener('submit',()=>submits++);
  const pending=draft(field);
  assert.equal(pending.text,'Hello');
  assert.equal(field.value,'Hello');
  assert.equal(pending.apply('你好'),true);
  assert.equal(field.value,'你好');
  assert.equal(inputs,1);assert.equal(submits,0);
  const stale=draft(field);field.value='New typing';
  assert.equal(stale.apply('old result'),false);
  assert.equal(field.value,'New typing');
  assert.equal(draft(document.querySelector('input')),undefined);
  assert.equal(draft(document.querySelector('[contenteditable]')),undefined);
});
