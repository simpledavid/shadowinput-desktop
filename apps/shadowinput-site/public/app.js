// Illustrative examples only. This page never submits text to an AI service.
const source = '这周的进度有点慢，有两个问题还没处理完，我想周五再给你一个完整的版本。';
const examples = {
  polish: { label: '润色后', source, result: '目前还有两个问题待处理，进度稍有延后。我计划在周五提供完整版本。' },
  translate: { label: '翻译后', source, result: 'Progress has been a little slow this week, with two issues still to resolve. I’m aiming to send you a complete version on Friday.' },
  rewrite: { label: '改写后', source, result: '还有两个问题待处理，我计划周五给你完整版本。' },
  continue: { label: '续写示例', source: '一个好的写作工具，应该让人更专注于表达。', result: '一个好的写作工具，应该让人更专注于表达。你负责把想法写下来，工具帮你理顺语言，最后的决定仍然留给你。' }
};

const tabs = [...document.querySelectorAll('[role="tab"]')];
const panel = document.querySelector('#example-panel');
const resultText = document.querySelector('#example-result');
const copyButton = document.querySelector('#copy-example');
const status = document.querySelector('#demo-status');
let mode = 'polish';
let copyRevision = 0;
let feedbackTimer;

function resetCopy() {
  clearTimeout(feedbackTimer);
  copyButton.querySelector('span').textContent = '复制示例';
  copyButton.querySelector('use').setAttribute('href', '#copy');
}

function selectMode(next, focus = false) {
  if (!Object.hasOwn(examples, next)) return;
  mode = next;
  copyRevision++;
  resetCopy();
  const example = examples[mode];
  for (const tab of tabs) {
    const selected = tab.dataset.mode === mode;
    tab.setAttribute('aria-selected', String(selected));
    tab.tabIndex = selected ? 0 : -1;
    if (selected && focus) tab.focus();
  }
  panel.setAttribute('aria-labelledby', `tab-${mode}`);
  document.querySelector('#example-source').textContent = example.source;
  document.querySelector('#result-label').textContent = example.label;
  resultText.textContent = example.result;
  status.textContent = `已切换到${example.label}。此内容为预置示例。`;
}

tabs.forEach((tab, index) => {
  tab.addEventListener('click', () => selectMode(tab.dataset.mode));
  tab.addEventListener('keydown', event => {
    const positions = { ArrowRight: (index + 1) % tabs.length, ArrowLeft: (index + tabs.length - 1) % tabs.length, Home: 0, End: tabs.length - 1 };
    if (!Object.hasOwn(positions, event.key)) return;
    event.preventDefault();
    selectMode(tabs[positions[event.key]].dataset.mode, true);
  });
});

copyButton.addEventListener('click', async () => {
  const revision = ++copyRevision;
  const text = examples[mode].result;
  resetCopy();
  try {
    if (!navigator.clipboard?.writeText) throw new Error('Clipboard unavailable');
    await navigator.clipboard.writeText(text);
    if (revision !== copyRevision) return;
    copyButton.querySelector('span').textContent = '已复制';
    copyButton.querySelector('use').setAttribute('href', '#check');
    status.textContent = '示例已复制到剪贴板。';
    feedbackTimer = setTimeout(resetCopy, 2400);
  } catch {
    if (revision !== copyRevision) return;
    copyButton.querySelector('span').textContent = '请手动复制';
    status.textContent = '浏览器未允许复制，请选中结果后手动复制。';
    const range = document.createRange();
    range.selectNodeContents(resultText);
    const selection = window.getSelection();
    selection.removeAllRanges();
    selection.addRange(range);
  }
});

if (/Windows/i.test(navigator.userAgent)) {
  document.querySelectorAll('[data-modifier]').forEach(key => { key.textContent = 'Ctrl'; key.classList.add('windows'); });
}

// The choice is stored on this device only; no preference is sent to a server.
const themeButton = document.querySelector('#theme-toggle');
function updateThemeButton() {
  const isDark = document.documentElement.dataset.theme === 'dark';
  const label = isDark ? '切换到浅色模式' : '切换到深色模式';
  themeButton.setAttribute('aria-label', label);
  themeButton.title = label;
  themeButton.querySelector('use').setAttribute('href', isDark ? '#sun' : '#moon');
  document.querySelector('meta[name="theme-color"]').content = isDark ? '#1d1d1b' : '#faf8f5';
}
themeButton.hidden = false;
updateThemeButton();
themeButton.addEventListener('click', () => {
  const next = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark';
  document.documentElement.dataset.theme = next;
  try { localStorage.setItem('shadowinput-theme', next); } catch { /* Private browsing can block storage. */ }
  updateThemeButton();
});

// These decorative mockups illustrate the desktop app, rather than running it.
const chapters = [...document.querySelectorAll('[data-chapter]')];
const tourArt = document.querySelector('#tour-art');
const tourStage = tourArt.parentElement;
function mountScene(target, scene) {
  const template = document.querySelector(`#scene-${scene}`);
  if (template) target.replaceChildren(template.content.cloneNode(true));
}
for (const stage of document.querySelectorAll('.mobile-stage')) mountScene(stage, stage.dataset.scene);
mountScene(tourArt, 'selection');
const desktopTour = matchMedia('(min-width: 651px)');
let scrollScheduled = false;
function updateTour() {
  scrollScheduled = false;
  if (!desktopTour.matches) return;
  const middle = window.innerHeight * .5;
  const active = chapters.reduce((nearest, chapter) => {
    const rect = chapter.getBoundingClientRect();
    const distance = Math.abs(rect.top + rect.height / 2 - middle);
    return distance < nearest.distance ? { chapter, distance } : nearest;
  }, { chapter: chapters[0], distance: Infinity }).chapter;
  const scene = active.dataset.chapter;
  if (tourStage.dataset.scene === scene) return;
  tourStage.dataset.scene = scene;
  mountScene(tourArt, scene);
  for (const chapter of chapters) chapter.classList.toggle('active', chapter === active);
}
function scheduleTour() {
  if (!scrollScheduled) { scrollScheduled = true; requestAnimationFrame(updateTour); }
}
window.addEventListener('scroll', scheduleTour, { passive: true });
window.addEventListener('resize', scheduleTour);
updateTour();
