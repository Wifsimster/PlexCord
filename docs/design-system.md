# PlexCord Design System — "ON AIR"

**Status: FINAL / canonical.** This document is the single source of truth for the interface. It supersedes "SIGNAL" (v1.14 – v4.3), whose behavioral rules — honest state, one state in one place, the motion catalogue, accessibility — it keeps; what changes is the identity layer: the metaphor, the color contract, the type, the mark and the Dashboard composition. Implementation agents must not invent any color, size, duration, or copy string — if a value is needed and not here, it is a spec bug; flag it, don't improvise. The brand layer (mark, wordmark, assets) lives in [`brand.md`](brand.md).

Stack constraints honored throughout: PrimeVue 4 `definePreset(Aura, …)`, Tailwind CSS 4 (`@theme` + utilities via `tailwindcss-primeui`), CSS custom properties, scoped/global keyframes, Vue 3 `<Transition>`. **One bundled display face** (Bricolage Grotesque, OFL, `@fontsource-variable/bricolage-grotesque`, loaded offline from the app bundle — never from a CDN); everything read in volume uses the platform UI stack. Audit findings referenced as **F1–F36** (see `UX_AUDIT.md` history).

---

## 1. Direction & principles

**Direction: ON AIR** (a broadcast desk). PlexCord does exactly one thing — it takes what you are playing on Plex and puts it *on air* on your Discord profile. The interface is built around the one signal a broadcast studio shows without words: the **ON AIR lamp**. A warm-graphite canvas, interaction in plain **ink** (the text color, filled), one lamp color — the **tally** — spent on nothing but "Discord is showing your media right now", and the media itself as the hero: the Dashboard is a **stage** washed in the colors of the artwork that is on air.

Why it replaced SIGNAL: SIGNAL was disciplined but anonymous — cool grey panels, a generic blue accent sitting beside Discord blurple (three adjacent blues), a mark too small to register, and a Dashboard whose actual output (the Discord card) was a small box three containers deep. ON AIR keeps the discipline and gives the product a voice and a face.

### 1.1 Principles

1. **On air, or not.** Every surface answers one question first: *is my media on my profile right now?* The tally lamp (§5.0.3) answers it in the topbar on every page and on the stage — lit, unlit, held, fault or standby.
2. **The media is the hero.** What is playing gets the biggest type and the only atmospheric color in the app (the stage wash, §4.1). The chrome around it stays graphite and ink.
3. **One state, one place.** A failure appears inside the owning connection tile and in the tally — nowhere else. Errors resolve or collapse; they are never "dismissed" into lying.
4. **Honest state, always.** Off air looks off air everywhere at once (tally, stage, specimen, Discord tile). The specimen is the single source of truth for *what Discord shows*; the stage is *what Plex plays* — input and output, side by side.
5. **Saturation = meaning.** The only saturated things on screen are: the tally (on air), status colors (connection health), the two endpoint pins, and the media's own artwork. Nothing is saturated for decoration.
6. **Motion narrates state change, never entertains.** 120–320ms, compositor-safe properties, ambient loops only on things that are literally live (the lit lamp, the wash, the equalizer), full reduced-motion substitutions.
7. **Precision layer everywhere.** Mono type is load-bearing: URLs, PINs, timestamps, format tokens, countdowns, versions. Tabular numerals so ticking numbers don't jitter.

### 1.2 Color contract (the review rule)

> **Ink is what you click. The tally is what's on air. Gold and blurple are who you're connected to. Green, amber and red are how those connections are doing. Everything else is graphite.**

Enforcement details:
- **Ink (`--pc-accent`)**: every interactive affordance — primary buttons (filled ink, pill), toggles-on, focus rings (in `--pc-text-secondary`), selected states (1px ink border + `--pc-accent-dim`), the wizard rail progress, the settings rail active item. Because ink is the text color, **links are told apart by an underline** (`.pc-link`), never by color alone. Never used for status.
- **Tally (`--pc-tally`)**: `#FF4D8D` dark / `#D0195E` light — the lit tally pill, the progress fill of media that is on air, the mark's badge, and the faint radial glow behind the relay on the Welcome step. Nothing else. Not a button, not a link, not a heading, not a selection, not a chart color.
- **Plex gold `#E5A00D` / Discord blurple `#5865F2`**: identify *endpoints only* — the node dots' neighbours in the topbar route, the two logos, wizard rail brand ticks, the 2px identity keyline on the Plex/Discord connection tiles, the Discord glyph on the "On your Discord profile" panel. Max area: a 24–28px icon or a 2px line. Never buttons, links, focus rings, or body text.
- **Semantic colors**: status only (dots, labels, tinted rows). Danger is **vermilion** (`#FF7155` / `#C7381C`), pushed toward orange so it can never be mistaken for the tally rose. A component never mixes brand pigment and semantic color in the same element.
- **There is no separate "info" hue.** Informational notices use neutral (muted text on `--pc-raised`).
- **The Discord specimen is theme-exempt.** It uses the dedicated `--pc-discord-*` palette exclusively (a faithful reproduction, framed like a specimen).
- **The stage wash is the only place artwork color reaches the chrome** (§4.1), always under its scrim.

### 1.3 Type

| Role | Face | Size / weight / tracking | Used for |
| --- | --- | --- | --- |
| Hero | Bricolage Grotesque | 40px / 700 / −0.035em, lh 1.08 | the stage title (what is on air) |
| Display | Bricolage Grotesque | 32px / 700 / −0.03em, lh 1.1 | wizard step headings |
| Title | Bricolage Grotesque | 24px / 700 / −0.025em | page title (one per page) |
| Tally | Bricolage Grotesque | 11.5px (12.5 lg) / 750 / +0.12em, `font-stretch: 85%`, uppercase | the tally label only |
| Wordmark | Bricolage Grotesque | 16px / 700 / −0.025em | `<BrandMark>` |
| Heading → Micro | platform UI stack | 15 / 14 / 12.5 / 11px (see tokens) | everything read in volume |
| Mono | platform mono stack | 13px | technical values |

Bricolage is display-only on purpose: it carries the voice at sizes where its character shows, and never has to render a settings caption at 12px on a Windows machine with ClearType.

---

## 2. Design tokens

File: `frontend/src/assets/tokens.css`, imported in `main.js` after the font and **before** `styles.scss`. This block is canonical.

Rules of use: **components reference role aliases (`--pc-bg`, `--pc-panel`, `--pc-text-muted`, `--pc-tally`, …), never raw ramp steps.** Raw steps exist so the preset (§3), Tailwind, and bespoke CSS resolve to identical hexes.

```css
/* ============================================================
   PlexCord "ON AIR" design tokens — canonical. Do not fork values.
   Contract: ink is what you click; tally is what's on air;
   gold & blurple are who you're connected to; green/amber/red
   are how the connections are doing. See docs/design-system.md.
   ============================================================ */

:root {
    /* ---------- Surface ramp — "studio graphite" (warm neutral, hue ≈ 30°, sat 3–8%) ---------- */
    --pc-surface-0: #ffffff;
    --pc-surface-50: #f5f3ef; /* light canvas ("paper") */
    --pc-surface-100: #eeebe6;
    --pc-surface-200: #e2ded8;
    --pc-surface-300: #c6c1b9;
    --pc-surface-400: #948e86; /* dark-mode muted text floor (AA on panel) */
    --pc-surface-500: #716b64;
    --pc-surface-600: #524d47; /* strong border / disabled text */
    --pc-surface-700: #34302d; /* hover border, active-row bg */
    --pc-surface-800: #211f1d; /* raised surface: inputs, chips, tiles */
    --pc-surface-850: #171514; /* panel/card background */
    --pc-surface-900: #121110; /* topbar / rails / overlays */
    --pc-surface-950: #0e0d0c; /* app canvas (window background) */

    /* ---------- Ink — interaction (what you click). Same ramp as the surfaces, read backwards. ---------- */
    --pc-accent-50: #fbfaf8;
    --pc-accent-100: #f2eee8; /* dark-mode primary fill */
    --pc-accent-200: #e2ded8;
    --pc-accent-300: #c6c1b9;
    --pc-accent-400: #948e86;
    --pc-accent-500: #716b64;
    --pc-accent-600: #524d47;
    --pc-accent-700: #34302d;
    --pc-accent-800: #211f1d;
    --pc-accent-900: #1a1817; /* light-mode primary fill */
    --pc-accent-950: #0e0d0c;

    /* ---------- Tally — the on-air lamp. PlexCord's own color, and only this job. ---------- */
    --pc-tally-400: #ff4d8d; /* dark mode: 5.8:1 on panel */
    --pc-tally-500: #ec2f74;
    --pc-tally-600: #d0195e; /* light mode: 5.3:1 on white */

    /* ---------- Brand pigments (endpoint pins only — see §1.2) ---------- */
    --pc-plex: #e5a00d;
    --pc-plex-dim: rgba(229, 160, 13, 0.12);
    --pc-blurple: #5865f2;
    --pc-blurple-dim: rgba(88, 101, 242, 0.12);

    /* ---------- Discord specimen palette (theme-exempt, pixel-faithful) ---------- */
    --pc-discord-bg: #111214;
    --pc-discord-raised: #1e1f22;
    --pc-discord-text: #dbdee1;
    --pc-discord-muted: #b5bac1;
    --pc-discord-green: #23a55a;
    --pc-discord-font: 'gg sans', 'Noto Sans', 'Helvetica Neue', Helvetica, Arial, sans-serif;

    /* ---------- Typography ----------
       Display: Bricolage Grotesque (bundled, OFL) — the wordmark, titles,
       the stage headline, and the tally label. Everything a user reads in
       volume stays in the platform UI face. */
    --pc-font-display: 'Bricolage Grotesque Variable', var(--pc-font-ui);
    --pc-font-ui:
        -apple-system, BlinkMacSystemFont, 'Segoe UI Variable', 'Segoe UI', system-ui, Roboto,
        'Helvetica Neue', Arial, sans-serif;
    --pc-font-mono:
        ui-monospace, 'SF Mono', 'Cascadia Code', 'Segoe UI Mono', 'Roboto Mono', Menlo, Consolas,
        monospace;

    --pc-text-hero: 2.5rem; /* 40px display w700 lh1.05 ls-0.035em — the stage title */
    --pc-text-display: 2rem; /* 32px display w700 lh1.1  ls-0.03em  — wizard heroes */
    --pc-text-title: 1.5rem; /* 24px display w700 lh1.2  ls-0.025em — page title (one/page) */
    --pc-text-heading: 0.9375rem; /* 15px ui w600 lh1.4  ls-0.01em — panel headers */
    --pc-text-body: 0.875rem; /* 14px ui w400 lh1.5 — default */
    --pc-text-caption: 0.78125rem; /* 12.5px ui w400 lh1.45 — helper, timestamps */
    --pc-text-micro: 0.6875rem; /* 11px ui w600 lh1.3 +0.08em UPPERCASE — eyebrows, badges */
    --pc-text-mono: 0.8125rem; /* 13px mono w400–500 — technical values */

    /* ---------- Spacing (strict 4px grid: 4 8 12 16 20 24 32 40 48 64) ---------- */
    --pc-space-control-x: 14px; /* inside buttons/inputs */
    --pc-space-panel: 20px; /* panel padding */
    --pc-space-panel-gap: 16px; /* gap between panels */
    --pc-space-section: 32px; /* between page sections */
    --pc-page-gutter: 24px; /* window edge → content */

    /* ---------- Radius (soft, but one family: 14 → 10 → 8 → 6, pill for tally & buttons) ---------- */
    --pc-radius-xs: 6px; /* chips, badges, mono tokens */
    --pc-radius-sm: 8px; /* inputs, toggles, icon buttons */
    --pc-radius-md: 10px; /* tiles, nested surfaces, toasts */
    --pc-radius-lg: 14px; /* panels/cards, dialogs, the stage */
    --pc-radius-full: 999px; /* buttons, tally, status dots, avatars */

    /* ---------- Elevation (depth = background steps + borders; shadows float only) ---------- */
    --pc-shadow-panel: inset 0 1px 0 rgba(255, 255, 255, 0.035); /* machined top-light */
    --pc-shadow-overlay: 0 6px 20px rgba(0, 0, 0, 0.38), 0 1px 3px rgba(0, 0, 0, 0.4);
    --pc-shadow-modal: 0 16px 48px rgba(0, 0, 0, 0.55), 0 2px 8px rgba(0, 0, 0, 0.4);

    /* ---------- Motion ---------- */
    --pc-dur-1: 120ms; /* micro: hover, press, focus */
    --pc-dur-2: 180ms; /* state color/border changes, toggles */
    --pc-dur-3: 240ms; /* element enter/exit, collapse/expand */
    --pc-dur-4: 320ms; /* route/step transitions, hero entrances */
    --pc-ease-out: cubic-bezier(0.22, 1, 0.36, 1); /* default: decisive arrival */
    --pc-ease-inout: cubic-bezier(0.65, 0, 0.35, 1); /* between two visible states */
    --pc-ease-in: cubic-bezier(0.4, 0, 1, 1); /* exits only */
    --pc-ease-snap: cubic-bezier(0.2, 0.9, 0.25, 1.1); /* tiny overshoot: dots, checks, toggles */
    /* Ambient loop periods (exempt from the 320ms cap — they ARE state) */
    --pc-loop-pulse: 2400ms; /* tally lamp glow / live status-dot ring */
    --pc-loop-blink: 1200ms; /* retrying dot */
    --pc-loop-eq: 1100ms; /* equalizer base period */
    --pc-loop-breathe: 6000ms; /* stage artwork wash */
    --pc-loop-shimmer: 1600ms; /* skeletons */
}

/* ================= DARK (default, design target) ================= */
.dark {
    --pc-bg: var(--pc-surface-950);
    --pc-panel: var(--pc-surface-850);
    --pc-raised: var(--pc-surface-800);
    --pc-overlay: var(--pc-surface-900);
    --pc-border: #292624; /* hairline */
    --pc-border-subtle: #1f1d1b; /* dividers inside panels */
    --pc-border-strong: var(--pc-surface-600);
    --pc-text: #f2eee8; /* 15.8:1 on panel */
    --pc-text-secondary: #b3ada5; /* 8.2:1 on panel */
    --pc-text-muted: #948e86; /* 5.6:1 on panel — AA floor */
    --pc-text-faint: #716b64; /* 3.5:1 — large/disabled only */

    /* Ink: the primary is the text color itself, filled. */
    --pc-accent: var(--pc-accent-100);
    --pc-accent-hover: #ffffff;
    --pc-accent-active: var(--pc-accent-200);
    --pc-accent-contrast: #141210; /* text on ink fills: 16:1 */
    --pc-accent-dim: rgba(242, 238, 232, 0.08);

    --pc-tally: var(--pc-tally-400);
    --pc-tally-contrast: #1a0710;
    --pc-tally-dim: rgba(255, 77, 141, 0.14);
    --pc-tally-glow: rgba(255, 77, 141, 0.45);

    --pc-success: #3dd68c; /* 9.7:1 on panel */
    --pc-success-dim: rgba(61, 214, 140, 0.1);
    --pc-warn: #f5b944; /* 10.3:1 on panel */
    --pc-warn-dim: rgba(245, 185, 68, 0.1);
    --pc-danger: #ff7155; /* 6.7:1 on panel — vermilion, kept well away from the tally's rose */
    --pc-danger-dim: rgba(255, 113, 85, 0.1);

    --pc-ring-focus: 0 0 0 2px var(--pc-bg), 0 0 0 4px var(--pc-text-secondary);
    --pc-backdrop-opacity: 0.55; /* stage artwork wash, under its scrim */
    --pc-stage-scrim: linear-gradient(90deg, rgba(23, 21, 20, 0.35) 0%, rgba(23, 21, 20, 0.72) 55%, rgba(23, 21, 20, 0.9) 100%);
    --pc-dialog-mask: rgba(6, 5, 4, 0.62);
}

/* ================= LIGHT (secondary, fully specified) ================= */
.light,
:root:not(.dark) {
    --pc-bg: var(--pc-surface-50);
    --pc-panel: #ffffff;
    --pc-raised: #f3f0eb;
    --pc-overlay: #ffffff;
    --pc-border: #e6e2dc;
    --pc-border-subtle: #eeebe6;
    --pc-border-strong: #cbc6be;
    --pc-text: #1a1817; /* 17.7:1 on white */
    --pc-text-secondary: #4f4a45; /* 8.8:1 */
    --pc-text-muted: #645e58; /* 6.4:1 */
    --pc-text-faint: #958f88; /* large/disabled only */

    --pc-accent: var(--pc-accent-900);
    --pc-accent-hover: var(--pc-accent-700);
    --pc-accent-active: #000000;
    --pc-accent-contrast: #ffffff;
    --pc-accent-dim: rgba(26, 24, 23, 0.06);

    --pc-tally: var(--pc-tally-600);
    --pc-tally-contrast: #ffffff;
    --pc-tally-dim: rgba(208, 25, 94, 0.09);
    --pc-tally-glow: rgba(208, 25, 94, 0.3);

    --pc-success: #0b7a4c; /* 5.5:1 on white */
    --pc-success-dim: rgba(11, 122, 76, 0.09);
    --pc-warn: #a15705; /* 5.3:1 on white */
    --pc-warn-dim: rgba(161, 87, 5, 0.09);
    --pc-danger: #c7381c; /* 5.2:1 on white */
    --pc-danger-dim: rgba(199, 56, 28, 0.09);

    --pc-ring-focus: 0 0 0 2px var(--pc-bg), 0 0 0 4px var(--pc-text-secondary);
    --pc-backdrop-opacity: 0.4;
    --pc-stage-scrim: linear-gradient(90deg, rgba(255, 255, 255, 0.45) 0%, rgba(255, 255, 255, 0.8) 55%, rgba(255, 255, 255, 0.92) 100%);
    --pc-dialog-mask: rgba(26, 24, 23, 0.42);
    --pc-shadow-panel: 0 1px 2px rgba(26, 24, 23, 0.04);
}

/* ================= Global base ================= */
html {
    font-size: 16px; /* removes the 14px Sakai override */
}
body {
    font-family: var(--pc-font-ui);
    font-size: var(--pc-text-body);
    color: var(--pc-text);
    background: var(--pc-bg);
    -webkit-font-smoothing: antialiased;
}
time,
.pc-num {
    font-variant-numeric: tabular-nums;
}
code,
.pc-mono {
    font-family: var(--pc-font-mono);
}
```

Theme plumbing (unchanged from SIGNAL): theme persisted in `localStorage['plexcord-theme']` (`'dark'` | `'light'`, **default `'dark'`**), applied to `document.documentElement` in `main.js` before mount; `Alt+D` and Settings → App → Theme both go through `toggleDarkMode`. One shared `transitions.css` holds all `<Transition>` classes and global keyframes.

---

## 3. PrimeVue preset mapping

`frontend/src/main.js` defines `PlexCordPreset = definePreset(Aura, …)`:

- **`primitive.ink`** — the graphite ramp (`#FBFAF8` … `#0E0D0C`). `semantic.primary.*` maps 1:1 onto it, so PrimeVue's "primary" is ink.
- **Dark:** `primary.color = ink.100 (#F2EEE8)`, `contrastColor #141210`, hover `#FFFFFF`, active `ink.200`. **Light:** `primary.color = ink.900 (#1A1817)`, `contrastColor #FFFFFF`, hover `ink.700`, active `#000`.
- **Surfaces** — the warm graphite ramp in both schemes; as in SIGNAL, dark `--p-surface-900` is deliberately the panel step (`#171514`), and the near-black rail value `#121110` is reachable only via `--pc-overlay`.
- `highlight` is ink at 6–16% alpha; `formField` borders are the hairline tokens, focus border is ink (`ink.300` dark / `ink.900` light); `focusRing.color = {text.muted.color}`.
- Component radii: button **999px (pill)**, input / inputnumber / select 8px, card / dialog 14px, toast 10px; toggle switch 36×20 with a 14px handle.

**Tailwind utility rule (unchanged):** `tailwindcss-primeui` bridges `--p-surface-*` → `bg-surface-*`; when a template needs canvas or rail values it uses role aliases (`bg-[var(--pc-bg)]`, `bg-[var(--pc-overlay)]`) or `.pc-panel`, never ad-hoc `bg-surface-950`.

---

## 4. Motion catalogue

Global rules: nothing transitions longer than 320ms except the ambient loops listed in §2; only `transform` and `opacity` animate (plus `background-color`/`border-color`/`color` on ≤180ms state changes); no property animates by default — every transition is declared per component. Precision tools don't bounce: no hover lifts, no springy translates (the only overshoot is `--pc-ease-snap` on ≤16px elements).

| # | What | Trigger | Property | Duration / easing | Reduced-motion fallback |
|---|---|---|---|---|---|
| M1 | Button hover | `pointerenter` | background/border-color | `--pc-dur-1` `--pc-ease-out` | instant color swap |
| M2 | Button press | `:active` | `transform: scale(0.985)` | 60ms in, `--pc-dur-1` release | none (skip) |
| M3 | Focus ring | `:focus-visible` | ring opacity 0.6→1 | `--pc-dur-1` | ring appears instantly |
| M4 | Status dot — live pulse | state = connected & playing | `::after` ring `scale 1→2.2`, opacity .45→0 | loop `--pc-loop-pulse` `--pc-ease-out`; dot itself static | static dot + its text label (label always present) |
| M5 | Status dot — retrying | state = retrying | dot opacity .4↔1 | loop `--pc-loop-blink` ease-in-out alternate | static warn dot + `RETRYING` label |
| M6 | Status label/dot change | store state flips | old: fade + translateY(−4px) out; new: fade + translateY(4px→0) in (`<Transition mode="out-in">`); dot color crossfade | `--pc-dur-2`; out `--pc-ease-in`, in `--pc-ease-out` | opacity-only 80ms crossfade (`.pc-fade-ok`) |
| M7 | Route transition (Dashboard ⇄ Settings) | router | out: opacity→0 + translateY(−4px) 120ms `--pc-ease-in`; in: opacity 0→1 + translateY(6px→0) 240ms `--pc-ease-out` | opacity-only 80ms |
| M8 | Wizard step slide | next/back | direction-aware: exit translateX(∓24px)+fade 160ms; enter translateX(±24px→0)+fade 240ms `--pc-ease-out`; rail marker slides `--pc-dur-3` `--pc-ease-inout` | opacity-only 80ms swap; rail marker jumps |
| M9 | Panel entrance (first mount) | page load | opacity 0→1 + translateY(8px→0), 280ms `--pc-ease-out`, 40ms stagger, max 4 panels | panels appear instantly |
| M10 | Specimen track change | new `sessionKey` | old artwork (absolute) fades 160ms; new fades in 240ms; text lines slide up 6px + fade — title +0ms, artist +30ms, album +60ms; ambient backdrop crossfades 320ms | instant swap |
| M11 | Playing equalizer | `isPlaying` | 3 bars (2×10px max) in `--pc-success`; heights loop 0.9s/1.1s/1.3s ease-in-out infinite, phase-shifted; paused → freeze at 30% height, tint `--pc-warn` over 180ms | static `pi pi-volume-up` glyph in state color |
| M12 | Progress bar | poll tick | width, 300ms linear (matches poll cadence → reads continuous) | width updates discretely (no transition) |
| M13 | Pause overlay | presence paused | specimen `filter: grayscale(.9) brightness(.8)` + `PAUSED` micro-badge fade-in | `--pc-dur-3` `--pc-ease-out` | filter + badge apply instantly |
| M14 | Toggle switch | change | handle translateX `--pc-dur-2` `--pc-ease-snap`; track color `--pc-dur-2` | instant |
| M15 | Inline "✓ Saved" | autosave success | check scales 0→1 `--pc-ease-snap` 180ms, holds 1600ms, fades 240ms | appears/disappears instantly, same 1600ms hold |
| M16 | Toast enter/exit | toast | in: translateX(16px→0)+fade 240ms `--pc-ease-out`; out: fade+translateX(8px) 180ms `--pc-ease-in` | opacity-only 80ms |
| M17 | Collapse/expand | Advanced sections, error details | `grid-template-rows 0fr→1fr`, `--pc-dur-3` `--pc-ease-inout`; content fade 120ms delayed 80ms | instant expand/collapse |
| M18 | Retry countdown | error with auto-retry | mono seconds tick (`tabular-nums`); per tick digit slides up 4px + fade 120ms | digits swap without motion |
| M19 | Success check draw | validation/connect success | SVG check `stroke-dashoffset` draw 320ms `--pc-ease-out` + circle fill fade 180ms delayed 120ms | check appears instantly |
| M20 | Skeleton shimmer | loading | `background-position` sweep, loop `--pc-loop-shimmer` linear, gradient `--pc-raised`→`--pc-surface-700` | static flat `--pc-raised` block |
| M21 | Dialog | open/close | open: scale .97→1 + fade 240ms `--pc-ease-out`, backdrop fade 180ms; close: 160ms `--pc-ease-in` | opacity-only 80ms |
| M22 | Stage artwork wash | Dashboard stage, artwork present | blurred art breathes between 80% and 100% of `--pc-backdrop-opacity` + `scale(1→1.04)` drift, loop `--pc-loop-breathe` (6s) ease-in-out while on air; `animation-play-state: paused` when held/off air (off air also desaturates it); unmount fade 320ms when idle | static at 80% of `--pc-backdrop-opacity`, no drift |

Implementation homes: M1–M3, M14 in preset/global CSS; M4, M5, M11, M20, M22 as keyframes in `transitions.css` (`pc-pulse`, `pc-blink`, `pc-eq`, `pc-shimmer`, `pc-breathe`); M6–M10, M13, M16–M19, M21 via Vue `<Transition>` classes + scoped styles.

**Reduced-motion policy** (in `transitions.css`):

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
  /* Whitelist: pure-opacity communication stays perceivable */
  .pc-fade-ok { transition: opacity 80ms linear !important; }
}
```

Semantic substitutions per the table above — never just frozen frames. State is never communicated by animation alone: every animated state has a color + text counterpart (WCAG 2.2.2 / 2.3.3). Wizard auto-advance (§5.4 step 3) is disabled for reduced-motion and screen-reader users.

### 4.1 The stage wash — exact spec

One absolutely-positioned `<img>` inside the Dashboard stage (§5.2), zero dependencies:

```css
.stage { position: relative; overflow: hidden; isolation: isolate; border-radius: var(--pc-radius-lg); }
.stage-wash {
  position: absolute; inset: -30%; width: 160%; height: 160%; object-fit: cover;
  filter: blur(64px) saturate(1.5);
  opacity: var(--pc-backdrop-opacity);           /* 0.55 dark · 0.40 light */
  z-index: -2; pointer-events: none;
  animation: pc-breathe var(--pc-loop-breathe) ease-in-out infinite;
}
.stage-scrim { position: absolute; inset: 0; z-index: -1; background: var(--pc-stage-scrim); }
```

The scrim is a left-to-right gradient of the panel color (35% → 72% → 90% dark; 45% → 80% → 92% light) — lightest behind the artwork, heaviest behind the text — so the title keeps AA contrast whatever the artwork is. Rules: renders only when `track.thumbUrl` exists; held or off air → `animation-play-state: paused`, off air also `grayscale(1)`; idle → unmounted via a 320ms opacity `<Transition>`; track change → crossfade (keyed img). **This is the only place artwork color reaches the chrome — it is never reused on other panels.**

---

## 5. Page-by-page implementation blueprint

### 5.0 Shared components (build first)

**5.0.1 `.pc-panel` — the one card system** (kills all three competing card systems):

```css
.pc-panel { background: var(--pc-panel); border: 1px solid var(--pc-border);
  border-radius: var(--pc-radius-lg); padding: var(--pc-space-panel);
  box-shadow: var(--pc-shadow-panel); }
.pc-panel--raised { background: var(--pc-raised); border-radius: var(--pc-radius-md); }
```

Header pattern: micro eyebrow (`--pc-text-micro`, `--pc-text-muted`, uppercase) + optional right-aligned meta/action; 12px below; hairline `--pc-border-subtle` divider only when the body is a list. Never nest more than one raised level. Rows: settings/list rows 44px min-height (12px vertical padding), hairline-divided; dense stat rows 32px; key-value rows in tiles 28px.

**5.0.2 Buttons** (PrimeVue `Button` via preset + severity mapping). Height 32px (sm 28, lg 38 — wizard primaries, 600 label, 20px x-padding), **pill** (`--pc-radius-full`), 14px/500 label, 14px x-padding, icon-only 32×32 round. One primary per view region max.

| Variant | Fill | Border | Text | Hover |
|---|---|---|---|---|
| Primary (ink) | `--pc-accent` | none | `--pc-accent-contrast` | fill `--pc-accent-hover` |
| Secondary | `--pc-raised` | 1px `--pc-border` | `--pc-text` | border `--pc-border-strong`, bg `--pc-surface-700` (light: `--pc-surface-200`) |
| Ghost | transparent | none | `--pc-text-secondary` | bg `--pc-raised`, text `--pc-text` |
| Ghost-danger | transparent | 1px `--pc-border` | `--pc-danger` | bg `--pc-danger-dim`, border danger @40% |

Loading: label persists at 40% opacity, 12px spinner absolutely positioned (no width jump). Disabled: 45% opacity, `cursor: not-allowed`; a gated disabled button is always paired with an inline caption stating the reason — never a bare dead button.

**5.0.3 The tally, and status indicators.**
- **Tally** (`.pc-tally`, the signature component): a 26px pill (30px `--lg` on the stage) holding a 7px lamp and an uppercase Bricolage label. It projects `usePresenceStatus().status` through `tally` — `{ kind, label }`:

  | status | kind | label (en) | look |
  |---|---|---|---|
  | `live` | `on` | ON AIR | tally fill, dark text, lamp with a 2.4s pulse ring (M4) |
  | `paused` (by you) | `off` | OFF AIR | hairline outline, hollow lamp |
  | `track-paused` | `hold` | PAUSED | hairline outline, amber lamp + label |
  | `plex-error` / `discord-error` | `fault` | PLEX DOWN / DISCORD DOWN | vermilion 40% outline, vermilion lamp + label |
  | `idle` | `idle` | STANDBY | hairline outline, grey lamp |

  Only the `on` state spends the tally pigment. The label is always present — state is never carried by the lamp's color alone.
- Dot: 8px circle; success / warn / danger / `--pc-surface-500` (idle). M4 live pulse, M5 retrying blink, M6 change. **Always paired** with an 11px micro label (`CONNECTED`, `RETRYING`, `PAUSED`, `IDLE`, `ERROR`) in the matching color.
- Badge/chip: micro type, radius-xs, 2px/6px padding, `*-dim` bg + full-strength text, optional 1px border at color@30%. Variants: `LOCAL` / `REMOTE` (neutral), `SAMPLE`, `PAUSED`, `CUSTOM ID`.
- Mono chip: `--pc-raised` bg, radius-xs, 1px/6px padding, 13px mono `--pc-text-secondary` — URLs, codes, versions, timestamps.

**5.0.4 `<DiscordSpecimen>`** — one shared component (Dashboard, Settings format editor, wizard Complete) replacing `setup/DiscordPreview.vue`. Pixel-faithful Discord dark activity card, theme-exempt (§1.2): bg `--pc-discord-bg`, radius 8px, 12px padding, width 340px max 100%, font `--pc-discord-font`; header `LISTENING TO PLEX` 11px/700 +0.02em `--pc-discord-muted` with `--pc-discord-green` headphone glyph; 64px artwork radius 6px (ghost placeholder: `--pc-discord-raised` + ♪ glyph in `--pc-blurple`); line 1 = rendered `details` format (14px/600 `--pc-discord-text`), lines 2–3 = rendered `state` / album (13px `--pc-discord-muted`), single-line ellipsis; progress: 4px track `--pc-discord-raised`, fill `--pc-discord-text`, mono tabular timestamps.

It renders the **user's real format strings** via a new shared util `frontend/src/utils/presenceFormat.js` → `renderPresenceFormat(format, track)` supporting `{track} {artist} {album} {year} {player}` (fixes F2/F11 at the root). Props: `track` (null → ghost idle), `formats {details, state}`, `paused`, `sample`. States: idle ghost (dashed 1px `--pc-border` frame, centered `–` glyph, caption per page spec); paused (M13 + `PAUSED` badge); track change (M10); skeleton (M20). Framed in the `.pc-specimen-well` inset (bg `--pc-bg`, 1px `--pc-border-subtle`, radius-md, 16px padding) with caption beneath: `Exactly what your Discord profile shows.` Pure renderer: no store init/cleanup of its own (F35).

**5.0.5 `<BrandMark>` / `<BrandSymbol>`** — the lockup (topbar, wizard rail): `<BrandSymbol>` at 20px (small optical geometry: ink ring + tally badge) + "PlexCord" in Bricolage 16px/700/−0.025em. `<BrandSymbol>` renders either optical size (picked by `size`, < 32 → small) and an `unlit` form (ring only) used as the Dashboard's idle artwork; see [`brand.md`](brand.md).

**5.0.6 Composables:** `useVersion()` (module-level cached `GetVersion()` — deletes the 3 duplicate calls), `usePresenceStatus()` (computed over playback + both connection stores + `IsPresencePaused` → the headline state machine in §5.1), `usePlayback()` (Dashboard-level playback event init/cleanup — F35).

**5.0.7 Toasts & dialogs.** Toast: `--pc-overlay` bg, `--pc-shadow-overlay`, radius-md, 1px `--pc-border`, 2px left rail in severity color, 14px title / caption body, top-right offset 56px, M16, auto-dismiss 4s (errors 8s + explicit close) with a 2px bottom progress hairline. **Usage contract: failures, update-available, reset only — never routine saves** (M15 handles those, F36). The update toast is a sticky `group="updates"` toast (no `life`) rendered in `AppLayout` and, because it needs an open window to be seen at all, it is **mirrored into the system tray** — `platform.UpdateNotice`, fed from `updater.OnStatusChange` (see docs/getting-started.md → Updates). Both surfaces say the same thing; the tray is the one that reaches a background-only session. Dialog: `--pc-overlay`, `--pc-shadow-modal`, radius-lg, 1px border, backdrop `--pc-dialog-mask`, M21; footer right-aligned ghost `Cancel` + primary/ghost-danger confirm; destructive confirms name the object and consequence ("Remove plex.local? PlexCord will stop publishing presence."). Add a `<ConfirmDialog>` host to `App.vue` — the registered ConfirmationService finally earns its keep.

### 5.1 Shell: AppLayout + AppTopbar + AppFooter

Topbar = the **route strip**: fixed 48px, bg `--pc-overlay`, bottom 1px `--pc-border`. Hidden on `/setup/*`.

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ ◐• PlexCord     ● Plex ── (● ON AIR) Bohemian Rhapsody ── ● Discord   ⏸ ⚙ │
└──────────────────────────────────────────────────────────────────────────────┘
```

- Left: `<BrandMark>`.
- Center: **route** — three pill-shaped button nodes joined by 20px hairline connectors (connectors tint `--pc-success` when both adjacent ends healthy, else `--pc-border`).
  - Plex node: 8px dot (state from `plexConnection.statusLabel`) + `Plex` 12.5px. Click → popover (overlay recipe): server host (mono), account, last sync, `Reconnect` ghost button.
  - Center node — **the tally + what is on air**: the tally pill (§5.0.3) followed, while `on` or `hold`, by the track title in Bricolage 14.5px/650 `--pc-text` (ellipsis max 280px). Click toggles pause (when relay is healthy) or routes to Dashboard (when errored). Transitions via M6.
  - Discord node: mirror of Plex node; popover shows Rich Presence state + last sync + `Reconnect`.
- Right: `Pause/Resume presence` icon button (global — F4), Settings cog (on `/settings` it swaps to a back-to-dashboard arrow). All 32px ghost icon buttons with tooltips (tooltip shows the shortcut in a mono chip) + focus rings. The theme lives in Settings → App (`Theme` select), not in the bar.
- **Keyboard shortcuts** (registered in AppLayout, guarded against input targets): `Ctrl/⌘+P` toggle pause, `Ctrl/⌘+,` settings, `Ctrl/⌘+R` retry the failed connection (no-op + toast "Nothing to retry" when healthy), `Alt/⌥+D` toggle dark/light theme (same M-transition path as the Settings → App theme select).
- Footer: one quiet caption line — `PlexCord v1.4.2 · a1b2c3d` (11.5px mono, `--pc-text-faint`, no chip, from `useVersion()`). **The only version display** (removed from Dashboard header; Settings→About shows it too but from the same composable).
- Layout container: `.layout-main-container { padding: 64px var(--pc-page-gutter) 24px; }` (48px bar + 16px). Pages **stop** declaring their own `min-h-screen`/background/padding (fixes the double-padding). Canvas `--pc-bg` comes from `body`. Route transitions: M7 on the main `<router-view>`. Remove `.layout-mask` and the dead `containerClass` from `AppLayout.vue`.
- The centered signal path is `position: absolute`, so the flanks are not part of its layout. It is capped at `min(60vw, 640px, calc(100% - 2 * (var(--topbar-flank) + var(--topbar-flank-gap))))` — `--topbar-flank` (239px) is the right cluster, the wider flank — so a long headline title ellipsizes instead of running into the action buttons at narrow window widths. The `640px` spec cap governs above ~1200px.

**Window geometry** (`window_state.go`, applied in `main.go`): the window is sized from the content above, not from round numbers.

| | Value | Derivation |
|---|---|---|
| Preferred width | 1248 | Dashboard content cap 1200 (§5.2) + 2×24px page gutter — the widest surface. Settings (1088) and the wizard (848) center inside it. |
| Preferred height | 700 | Dashboard in its tallest (idle) state: 64px top gutter + 525px panels + 46px footer + 24px ≈ 663, plus air. The view that stays open sets the height; Settings (~2170px) and the wizard's Complete step (796px) scroll instead. |
| Minimum width | 880 | Where the wizard's step column reaches its full 560px (240px rail + 2×24px gutter + 560). Below it the Dashboard (<992px) and Settings (<860px) responsive fallbacks are still correct — the floor just stops anything being *forced* into them. |
| Minimum height | 600 | Deliberately below every view's natural height: pages scroll and the wizard has its own scroller with a pinned footer. A taller floor is what stops the window fitting a scaled or small display. |

Wails takes its window size before any screen is known, so the preferred size is the static option and `App.adaptWindowToScreen` corrects it on startup: it clamps to 95% × 90% of the screen the window opened on (Wails v2 reports screen size, not work area, so the remainder absorbs the taskbar/dock), never below the minimum and never larger than the screen itself, then re-centers.

### 5.2 Dashboard (`/`) — the stage

Target: one glance answers *on air / what / broken*; off air is honest everywhere; one failure appears in exactly one place; the media is the hero.

```
┌──────────────────────────── STAGE (artwork wash) ────────────────────────────┐
│ ┌────────┐  (● ON AIR)  Now playing on Plexamp                               │
│ │  ART   │  Midnight City                              ← hero, Bricolage 40  │
│ │ 184px  │  M83 — Hurry Up, We're Dreaming                                   │
│ └────────┘  1:39 ━━━━━━━━━━━━━───────────── 4:03       ← tally fill          │
└──────────────────────────────────────────────────────────────────────────────┘
┌── ON YOUR DISCORD PROFILE ─────────┐ ┌── CONNECTIONS ── Polling every 5s · … ─┐
│   <DiscordSpecimen>                │ │ [Plex tile]           [Discord tile]   │
└────────────────────────────────────┘ └────────────────────────────────────────┘
```

Layout ≥ `lg` (992px): content `max-width: 1200px`, centered, grid `minmax(0,5fr) minmax(0,7fr)`, gap `--pc-space-panel-gap`; the stage spans both columns. Below `lg`: single column (stage, specimen, connections); connection tiles side-by-side ≥ `md`. Panels enter with M9 stagger. No manual Refresh button (F7). Playback event lifecycle initialized once here via `usePlayback()` (F35).

**Stage** (full width, radius-lg, hairline border, min-height 248px, 32px padding):
- Background: the **stage wash** (§4.1) under `--pc-stage-scrim`.
- Artwork: 184×184 (music) or 132×198 poster (film/episode), radius-md, a soft drop shadow + 1px inner hairline — the only shadow on a non-floating element, because the artwork is the one physical object in the app.
- Meta column: tally `--lg` + `Now playing on {player}` caption (on/hold only) → **title** (hero type, single-line ellipsis, `<TruncatedText>` tooltip) → subtitle 16px `--pc-text-secondary` shaped by media type (music `artist — album`; film `year`; episode `show · S1 · E1`) → progress: 4px track `--pc-text`@14%, fill `--pc-tally` while on air (`--pc-text-muted` while held), mono tabular times, M12.
- **Idle (F1/F2):** artwork slot becomes a dashed 1.5px `--pc-border-strong` square holding the **unlit mark** (`<BrandSymbol unlit>`, ring only — the logo off air); tally `STANDBY`; title **"Nothing on air."**; caption (`{n}` = live polling interval): **"Play something in Plexamp or any Plex player and it appears here — and on Discord — within ~{n}s."** Wash unmounted.
- **Off air (paused by you — F4):** artwork `grayscale(.9) brightness(.8)`, wash desaturated and frozen; tally `OFF AIR`; progress replaced by caption **"Hidden from Discord while presence is paused."** + primary ink button **"Go back on air"** (▶). Specimen M13 grayscale + `PAUSED`; Discord tile row reads `presence  Hidden (paused)`. One click (this button, the topbar tally, the topbar ⏸, or `Ctrl+P`) resumes.
- **Held (player paused):** tally `PAUSED` (amber lamp), wash frozen, progress fill muted.
- **Loading:** skeleton art + tally pill + two bars (M20), minimum 400ms.

**On your Discord profile** (`.pc-panel`, left): `ON YOUR DISCORD PROFILE` eyebrow + blurple Discord glyph (endpoint pin); the `<DiscordSpecimen>` centered, max-width 460px, caption hidden (the eyebrow says it). Idle: specimen ghost "Nothing playing on Plex".

**Connections panel** (`.pc-panel`, right; two side-by-side `--pc-panel--raised` tiles, header `CONNECTIONS` with the caption `Polling every 5s · Ctrl+P to pause` right-aligned):
- Each tile: `border-top: 2px solid var(--pc-plex | --pc-blurple)` identity keyline (the only non-neutral borders in the app); header row: 16px brand glyph in brand pigment, name, dot + micro label right (M4/M5/M6); then 28px key-value caption rows, values in mono chips where technical: Plex → `acct battistella`, `srv plex.local:32400`, `sync 12s ago`; Discord → `presence Active | Hidden (paused) | Inactive`, `sync 3s ago` (real store data — F5).
- **Error/retry state (replaces `ErrorBanner.vue` — F5/F9):** tile gains `border-left: 2px solid var(--pc-danger)`; a collapsible detail region (M17) opens inside: error title (body-strong), suggestion (caption muted), error code (mono chip), and when auto-retrying `Retry #3 in 12s` (M18; note: store `retryState.nextRetryIn` is **nanoseconds** — convert) beside an always-available `Retry now` ghost-danger button. Nothing to dismiss; on recovery it collapses (M17) and the dot flips green (M6). The expanded/collapsed choice is store-tracked **per error code** (Crossfade graft) so a user-collapsed detail doesn't zombie open on the next status event of the same failure.
- `Reconnect` is available in the tile popover/expanded state even when healthy (F5).
- **Setup-skipped (F22):** when `CheckSetupComplete()` is false, a full-width slim tile: `Setup incomplete — Resume setup →` (underlined ink link) routing to the first unfinished wizard step.

### 5.3 Settings (`/settings`)

Two-pane, content max-width 1040px: sticky left rail (200px) + anchored sections. Header: back arrow handled by topbar swap (§5.1); page title `Settings` in `--pc-text-title` (the one per-page title).

- **Rail:** `SETTINGS` eyebrow, then 5 anchor links: Connection / Presence / App / Advanced / About. Scroll-spy via IntersectionObserver; active link = accent text + 2px left accent bar sliding between items (`--pc-dur-3` `--pc-ease-inout`). Rail is a listbox: ↑/↓ + Enter.
- **Sections (goal-based regrouping — F15):**
  - **Connection:** Plex servers list + polling interval (InputNumber 1–60s).
  - **Presence:** details/state format inputs with **token chips** beneath (`{track} {artist} {album} {year} {player}` as clickable mono chips inserting at caret); live `<DiscordSpecimen>` below rendering live playback when present, else Queen sample data with `SAMPLE` badge; edits update the specimen per keystroke (local), save on blur. Hide-when-paused toggle + delay InputNumber (0–300s).
  - **App:** Start on login, Minimize to tray (ToggleSwitch rows).
  - **Advanced:** Discord Client ID — validated input with the one deliberate explicit `Apply` button in Settings + caption `Applying reconnects Discord` (F16); `Send test presence` button (F13) with transient inline result (M15).
  - **About:** version/commit mono chips (`useVersion()`), `Check for updates`, update-available row (neutral raised recipe + primary `Download`), changelog link, then a hairline-separated **Danger zone**: heading in `--pc-danger`, ghost-danger `Reset application…` → ConfirmDialog per §5.0.7 (keeps existing consequence list).
- **Save model (F10/F36):** toggles apply instantly (optimistic + revert kept); text/number fields save on blur or 600ms debounce; feedback is the inline `✓ Saved` micro-indicator (M15) at the field's right edge. **No toasts for routine saves.** Fields requiring a poll/reconnect cycle (polling interval, client ID) carry a one-line caption: `Applies on next poll cycle`.
- **Servers (F12/F14):** each row (44px): health dot (lightweight validate ping), name, mono URL chip, monitored user, per-row `Test` ghost button, active ToggleSwitch, delete trash. Delete → `useConfirm` ConfirmDialog: "Remove plex.local? PlexCord will stop publishing presence." (stronger copy when it's the only active server). **Add Server** dialog uses the URL validator extracted from SetupPlex into `frontend/src/utils/plexUrl.js`; servers saved without auth render with warn dot + `Needs sign-in → Authenticate` link deep-linking into the wizard Plex step (PIN auth reuse) — no silent dead servers.
- States: initial load = skeleton rows (M20); per-control loading = 12px inline spinner, never button-wide.

### 5.4 Setup wizard (`/setup/*`)

Full-window (no topbar), canvas `--pc-bg`. `SetupWizard.vue` replaces the horizontal StepList with a **left-rail wizard** — the rail is the signal path being assembled:

```
┌────────────────────┬─────────────────────────────────────────────┐
│ ⟅⟆ PlexCord SETUP  │  (step content, max-width 560px, left-set)  │
│ ● Welcome          │                                             │
│ ◉ Plex Server      │                                             │
│   plex.local       │                                             │
│ ○ Select User      │                                             │
│ ○ Discord          │                                             │
│ ○ Done             │─────────────────────────────────────────────│
│ Skip setup →       │            [← Back]          [Continue →]   │
└────────────────────┴─────────────────────────────────────────────┘
```

- **Rail** (240px, bg `--pc-overlay`, right 1px `--pc-border`): steps as 40px rows — glyph (done: 14px success check, M19 draw-on; current: accent ring dot; locked: hollow muted dot), label, and **for done steps a one-line mono summary of the accumulated result** (`plex.local`, `battistella`, `Connected`) — the setup visibly assembles the working configuration. A 2px accent progress line runs down the left edge to the current step (M8). Plex/Discord steps carry a 6px brand-pigment tick beside their glyph (only brand color in wizard chrome). Back-navigation clickable; forward steps `aria-disabled` at 40% opacity with tooltip `Complete the current step first` (F19).
- **Footer bar** (right pane, top hairline): Back (ghost) — **including on Complete** (F20); primary **Continue** — **always rendered, disabled when gated**, with inline caption stating the gate: `Validate your server to continue` / `Select a user to continue` / `Connect Discord to continue` (F18). Last step: `Finish setup`. `Skip setup` bottom-left in rail, caption size, steps 2–4 only; skipping sets the flag behind the Dashboard resume tile (F22).
- **Keyboard (F17):** `handleKeydown` early-returns when `event.target.closest('input, textarea, [contenteditable], .p-inputtext')`. ArrowRight only advances when Continue is enabled; Enter submits the current step's primary action.
- Step content transitions: M8. Each step: `--pc-text-display` heading + caption lede, then panels.

**Step 1 — Welcome.** `display` heading **"Put your Plex on air."**, 15px lede naming all three media kinds, static route diagram (Plex glyph ┄ dashed hairline ┄ the regular-geometry mark at 56px on an 88px `--pc-overlay` tile labelled "PlexCord" in Bricolage ┄ Discord glyph; brand pigments at icon size; a faint radial `--pc-tally-dim` glow behind the relay tile — the wizard's only glow), two-row checklist ("A Plex account", "Discord running on this computer") as 44px rows with muted check glyphs. **Single CTA:** footer Continue labeled `Get started` — hero button removed (F21). M9 stagger.

**Step 2 — Plex Server (auto-chained — F23).** Happy path: *Sign in → click your server → Continue.*
1. Sign-in panel, state machine kept: `initial` = primary `Sign in with Plex` + caption "Opens plex.tv in your browser"; `waiting` = PIN in 28px mono +0.12em `--pc-text` on `--pc-raised` chip, caption "Enter this code at plex.tv/link", 12px spinner, ghost `Open browser again` / `Cancel`; `success` = row collapses (M17) to `✓ Signed in as battistella` (M19). Stale "enter your token" copy deleted (F24).
2. On auth success **discovery auto-runs**: skeleton rows → selectable server rows (status dot, name, mono URL, `LOCAL`/`REMOTE` badge). Persistent caption-link below the list with visible inline hint (F25): "Docker or remote server? `Enter its address manually`." Manual mode: labeled input, live format validation (border + caption in success/danger, via `utils/plexUrl.js`), placeholder `http://192.168.1.10:32400`.
3. Selecting a server (or valid manual URL + Enter) **auto-validates**: inline 12px spinner in the row → success: row check + `✓ Reachable — 2 music libraries` (the datum that matters; version/library metric cards and Re-validate removed — F26), Continue enables. Failure: row bordered danger, inline error + suggestion + ghost `Retry` (M17).

**Step 3 — Select User (F27/F28).** Exactly one user: auto-select, done-panel `✓ Monitoring battistella` + caption "Only one user on this server", **auto-advance after 800ms** with a `Stay` link during the beat; reduced-motion/screen-reader users get no auto-advance — Continue is enabled and focused instead. Multiple users: select-card grid (raised tiles; selected = 1px `--pc-accent` border + `--pc-accent-dim` bg + check); the redundant confirmation Message is removed — the selected card is the confirmation. Loading: 4 skeleton cards. Error: inline danger panel + `Retry` / `Back`.

**Step 4 — Discord (F29/F30/F31).**
- `initial`: caption notice "Discord must be running on this computer", primary `Connect to Discord`.
- `connected`: done-panel `✓ Connected to Discord` (M19) + secondary `Send test presence` (transient inline `✓ Sent — check your Discord profile`, M15). `Disconnect` removed; `Test Again` renamed `Reconnect`, demoted to caption-link (F30).
- `error`: danger panel, friendly-mapped message, `Retry`.
- Gating (F29): Continue disabled + caption `Connect Discord to continue`, with escape hatch caption-link `Continue without Discord →` (sets a flag surfaced on Complete).
- Advanced (collapsible M17): custom Client ID with live validation; **collapsing no longer clears the input** (F31) — value persists, `CUSTOM ID` badge shows when set; portal instructions in a quiet raised sub-panel.

**Step 5 — Complete (F32).** Side effects move from `onMounted` to the **Finish action** (`finishSetup`: connect Discord if needed → `StartSessionPolling` → `CompleteSetup`), idempotent behind a store flag, failures surfaced as a danger panel on this step — never swallowed. Renders: 40px M19 drawn success check, `display` **"Setup complete."**, live `<DiscordSpecimen>` (real playback, or ghost idle labeled "Play something on Plex to see it live"), and a 3-row "What happens next" checklist. If Discord was skipped: warn panel `Discord not connected — presence won't publish until you connect it in Settings.` Back remains available (F20).

### 5.5 NotFound (`/:pathMatch(.*)*`)

Rebuild `views/pages/NotFound.vue` in place (route stays): one `.pc-panel`, mono `404` in `--pc-text-faint` at 26px, caption "This route doesn't exist.", primary `Back to dashboard`. No FloatingConfigurator.

---

## 6. Accessibility rules

1. **Contrast:** all token pairings in §2 meet WCAG AA at their specified sizes (ratios annotated inline). Standing rules: `--pc-text-muted` is the floor for body-size text; `--pc-text-faint` only ≥18.66px/bold or disabled states; gold/blurple never carry text; warn-filled elements would need dark text — but warn is never a fill in this system (dim tints + colored text only).
2. **Focus:** `:focus-visible` ring on **every** interactive element — `box-shadow: var(--pc-ring-focus)` (2px offset ring, M3). Never `outline: none` without the ring replacement. Wizard rail and settings rail are keyboard-navigable listboxes.
3. **Reduced motion:** policy + per-animation substitutions in §4. No information is motion-only; every animated state has a color + text twin. Wizard auto-advance disabled under reduced-motion/AT.
4. **Hit targets:** minimum 32×32px for icon buttons, 44px row height for list/settings rows; the 8px status dots are non-interactive (their parent node/tile is the target).
5. **State semantics:** status dots always paired with visible text labels; form errors rendered as text below the field with `aria-describedby` and `role="alert"`/live-region announcement, icon + color (never color alone); disabled-with-reason pattern (inline caption) instead of hidden controls; toasts use `aria-live="polite"` (errors `assertive`).
6. **Dialogs:** focus trapped, `Esc` closes, focus returns to invoker (PrimeVue defaults — do not disable).
7. **Keyboard shortcuts** never fire when an input/textarea/contenteditable has focus.
8. **Language of state:** the topbar headline, specimen, and tiles must never disagree; paused/error copy names the consequence ("Discord is not showing your music"), not just the state.

---

## 7. What to delete

Execute in the same PR as the token/preset landing (step 1 of sequencing). All paths relative to `frontend/`.

### 7.1 Delete outright (no importers — verified against inventory)

| Path | Why |
|---|---|
| `src/components/StatusCard.vue` | superseded generic status card |
| `src/components/MetricDisplay.vue` | no importers |
| `src/components/Greet.vue` | Wails demo |
| `src/components/URL.vue` | no importers |
| `src/components/FloatingConfigurator.vue` | only used by dead auth pages + old NotFound |
| `src/layout/AppMenu.vue`, `src/layout/AppMenuItem.vue` | no sidebar exists |
| `src/layout/AppConfigurator.vue` | only via FloatingConfigurator |
| `src/views/pages/auth/Login.vue`, `Access.vue`, `Error.vue` (whole `auth/` dir) | not routed |
| `src/stores/connection.js` | replaced by split stores |
| `src/stores/history.js` + `src/stores/__tests__/history.test.js` | calls nonexistent Go bindings |
| `src/service/CountryService.js`, `CustomerService.js`, `NodeService.js`, `PhotoService.js`, `ProductService.js` | Sakai demo data (~11k lines) |
| `public/demo/` (whole dir) | demo assets shipped to dist |
| `src/assets/layout/_menu.scss`, `_responsive.scss`, `_preloading.scss` | style nonexistent markup |
| `src/assets/layout/variables/_light.scss`, `_dark.scss` | replaced by `tokens.css` role aliases (after migrating `--surface-ground`/`--surface-card` reads) |
| `src/assets/images/wails-logo-universal.png` | unused |
| `src/assets/fonts/lato.css` + `src/assets/fonts/lato/` (all woffs) | Lato dropped for system stack — also remove the `<link>` in `index.html` |
| `chart.js` entry in `package.json` | no usage in src |

### 7.2 Retired by the redesign (deleted once their replacement lands)

| Path | Replaced by |
|---|---|
| `src/components/ErrorBanner.vue` | error detail inside Connections tiles (§5.2) |
| `src/components/NowPlaying.vue` | Presence panel + `<DiscordSpecimen>` |
| `src/components/setup/DiscordPreview.vue` | `<DiscordSpecimen>` |
| `src/components/ConnectionStatus.vue`, `PlexStatusCard.vue`, `DiscordStatusCard.vue` | Connections panel tiles (one component, `source` prop) |

### 7.3 Edit in place (dependencies to update when deleting)

- **`index.html`:** remove `class="app-dark"` from `<html>`; remove the Lato `<link>`; title stays `plexcord`.
- **`src/App.vue`:** remove the unconditional `.dark` add in `onMounted` (theme now applied pre-mount in `main.js`); add the `<ConfirmDialog>` host next to `<Toast>`.
- **`src/layout/AppLayout.vue`:** remove `.layout-mask` div and the dead `containerClass` computed. It keeps wrapping Topbar/router-view/Footer.
- **`src/layout/composables/layout.js`:** delete `toggleMenu`, `setActiveMenuItem`, `layoutState.activeMenuItem`, `getPrimary`, `getSurface`, and stale `layoutConfig.primary/preset/menuMode`; keep only dark-mode state + persisted `toggleDarkMode`.
- **`src/router/index.js`:** no route deletions — `NotFound.vue` is rebuilt at the same path (§5.5); the legacy `pages/dashboard` redirect stays. The `beforeEach` setup guard stays.
- **`src/assets/layout/_core.scss`:** delete the `html { font-size: 14px }` override and body font/background rules (moved to `tokens.css`).
- **`src/assets/layout/_topbar.scss`:** delete the `.config-panel` block; rework remaining rules for the 48px signal strip.
- **`src/assets/layout/_utils.scss`:** delete `.clearfix`; delete `.card` **after** Settings and NowPlaying no longer use it (they're rebuilt on `.pc-panel`); the `.p-toast` 100px offset rule is replaced by the 56px offset (§5.0.7).
- **`src/assets/layout/_typography.scss`:** delete the global h1–h6/p/blockquote sizing — the type scale (§2) replaces it.
- **`src/assets/layout/layout.scss`:** drop the deleted `@import`s.
- **`src/types/events.js`:** delete the unused `SessionEvents` enum; keep the typedefs.
- **`src/main.js`:** preset swap (§3); import order `tokens.css` → `tailwind.css` → `styles.scss` → `transitions.css`; ConfirmationService registration **stays** (now used).
- Per-view scoped keyframes (`fadein` ×3, `slideDown`, `scaleIn`) are deleted as each view is rebuilt on `transitions.css`.

### 7.4 Implementation sequencing

1. Tokens + preset + theme plumbing + §7.1 purge (one PR).
2. Shell: signal strip, BrandMark, footer, layout padding, route transitions, shortcuts, `transitions.css`.
3. Shared components: `.pc-panel`, buttons, status indicators, `<DiscordSpecimen>` + `presenceFormat.js`, M15 indicator.
4. Dashboard merge (+ retire §7.2 components).
5. Settings regroup + rail + autosave + confirms.
6. Wizard rail shell + step fixes (keyboard guard first — it's a bug), auto-chain, Complete side-effect move.
7. Reduced-motion + focus-visible audit; light-theme sanity pass; Tailwind utility re-audit (§3 rule).
