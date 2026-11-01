# shadowinput Chinese website

The buildless website for [shadowinput.com](https://shadowinput.com/). Serve `public/` directly. This directory is separate from the upstream Arion website in `apps/site/`.

The interactive examples use preset text. They do not call a model. Downloads
link directly to the input-method preview DMG/EXE and standalone browser ZIP.
The HTML download rows own versions and URLs; each release includes corresponding
source and checksums. Input-method runtime limits and the extension's unverified
real authorization/browser flow are visible before download. See the
[extension verification ledger](../../docs/project/shadowinput-extension.md).

## Cloudflare deployment

Use the existing Pages project **shadowinput**, which already owns `shadowinput.com` and `shadowinput.pages.dev`. Do not create a replacement project or delete its existing resources.

From this directory, after `wrangler login`:

```sh
npx wrangler@4 pages deploy public --project-name shadowinput --branch desktop-site-preview
```

Verify the preview, then publish to the existing production branch:

```sh
npx wrangler@4 pages deploy public --project-name shadowinput --branch claude/initial-setup-Pt7FD
```

The unusual production branch name belongs to the existing Pages project. Its Git source remains the older `simpledavid/ShadowInput` repository. Automatic production and preview builds are disabled to prevent that repository from overwriting this website; updates use Wrangler. This folder is the maintained website source in `simpledavid/shadowinput-desktop`.

This site has no Pages Functions or database access. Existing project bindings and old deployment history are retained. For rollback, use the Pages deployment history; the production deployment immediately before the desktop-site replacement was `fc092367-198a-4c32-aac3-fbff2f21f1f8`.

## Preview and attribution

```sh
python3 -m http.server 4173 --directory public
```

Content and layout were imported from website source revision `9e9887e83d35888ad4d5cc77c37d5e393e16e357`, with canonical and Open Graph URLs set to `https://shadowinput.com/`. The warm ivory/yellow theme is retained; all six upstream flower image files were replaced with independently generated ginkgo, water, and paper backgrounds. See `public/credits.html` and the license files in `public/assets/` for asset credits. No hosting credentials are included in this directory.
