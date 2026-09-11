# PlexCord Brand

The interface has its own canonical spec — [`design-system.md`](design-system.md),
"SIGNAL". This document covers the layer above it: the product's own mark, and
the rules that keep it from colliding with the two brands it sits between.

## Why PlexCord needed a mark of its own

Until v1.14 the mark was Plexamp's polyline glyph recolored to Plex gold. That
was a problem on three counts, and all three are the reason this exists:

1. **It was not ours.** Plex's marks belong to Plex, and PlexCord now ships them
   inside a Windows installer, an *Apps & features* entry and release binaries.
2. **It broke the app's own color contract.** In SIGNAL §1.2, gold identifies
   *the Plex endpoint of the signal path* — a 24–28px chip or a 2px keyline.
   Making gold the entire product mark spent an endpoint pigment on the product.
3. **It made PlexCord invisible.** In a taskbar, an amber chevron is Plexamp.

## The mark

Three masses on one axis: a dash, a dot, a dash.

It is the signal path the app already draws across its own topbar —
`● Plex — ▮ Live — ● Discord` — compressed until only the structure is left:
two ends, and the relay between them. PlexCord plays nothing and stores nothing;
it takes a signal from one side and puts it on the other, and that is the whole
drawing.

The masses are separated, never joined by a connecting line. The alignment does
the work, and the gaps are what survive at 16px.

### Two optical sizes

The mark is drawn twice. This is not a variant to pick by taste — it is picked
by size:

| Geometry | Use at | Dash | Dot | Glyph box |
| --- | --- | --- | --- | --- |
| **Regular** | 48px and above | 15 × 9, r4.5 | r8.5 | 74% of the tile |
| **Small** | 32px and below | 13 × 11, r5.5 | r10.5 | 86% of the tile |

Below ~32px the regular geometry's masses fall under a pixel and the motif
closes up into a smudge. The small geometry shortens and thickens them and lets
the glyph breathe wider in its tile. The in-app `<BrandMark>` (18px) and the
tray icon both use it.

### Color: neutral, and only neutral

The mark is `--pc-text` (`#EDEDF0` on dark, the near-black ink on light). It
never takes a pigment, and there is nothing arbitrary about that — the app's
color contract has already spent the spectrum:

- **Signal Blue** is what you click.
- **Plex gold** and **Discord blurple** are what you are connected to.
- **Green / amber / red** are what is happening.

A product mark in any of those either lies about its role or competes with the
endpoint it sits next to. Neutral is the one register left, and it is the right
one: the relay is not a participant in the conversation it carries.

Gold and blurple may still appear *beside* the mark as endpoint pins — the
README banner does exactly that, one dot per end of the line — but never
*inside* it.

### The update badge

While an update is waiting to be applied, the tray icon carries an amber dot in
its bottom-right corner (`--pc-warn`, the app's "something wants your
attention" register). It exists because the rest of the update notice is
passive: the tray menu item has to be opened to be read and the tooltip has to
be hovered, while a badged icon is visible at rest — which is the whole point
for an app built to run minimized.

Two details are load-bearing:

- **The gap around the dot is erased, not filled with the tile color**, so the
  badge reads the same whatever the icon sits on.
- **The badge is sized optically too** — `r` 0.105 of the tile above 32px,
  0.150 at and below it. One that still registers at 16px would dominate the
  mark at 128px. It is placed at 0.78 of the tile on both axes: far enough into
  the corner to clear the symbol's right dash, close enough that its gap stays
  inside the tile's rounded silhouette rather than hanging a crescent off it.

### Clear space and minimum size

- **Clear space**: the height of the dot, on all four sides.
- **Minimum size**: 16px for the symbol. The lockup's floor is a 16px symbol;
  below that, drop the wordmark and keep the symbol alone.

## The wordmark

"PlexCord", set in the design system's UI stack at weight 600 and `-0.02em`
tracking. No webfont — SIGNAL forbids them, and the wordmark honors the same
rule, so the lockup in the topbar and the lockup in the README are the same
object composed with whatever the host has.

The lockup is horizontal: symbol, then a gap of 42% of the symbol's height,
then the word. The optional uppercase suffix (`SETUP`) is a third element in
`--pc-text-micro`, muted — it is a mode label, not part of the mark.

## Files

| File | What it is |
| --- | --- |
| `build/brand/plexcord-mark.svg` | Symbol, regular geometry, `currentColor` |
| `build/brand/plexcord-mark-small.svg` | Symbol, small optical geometry |
| `build/brand/plexcord-icon.svg` | App icon master: tile + symbol |
| `build/brand/plexcord-icon-small.svg` | App icon, small optical geometry |
| `build/brand/plexcord-lockup.svg` | Symbol + wordmark |
| `build/brand/plexcord-icon-update.svg` | App icon with the update badge |
| `build/appicon.png` | 1024px app icon (macOS, Linux, tray) |
| `build/appicon-update.png` | Badged tray icon (macOS, Linux) |
| `build/windows/icon.ico` | 16/24/32/48/64/128/256, optical variant below 48 |
| `build/windows/icon-update.ico` | Badged tray icon (Windows), same sizes |
| `docs/images/banner.svg` | README banner |
| `frontend/src/components/BrandMark.vue` | The in-app lockup |

The icon tile is `--pc-surface-850` (`#17181D`) with a corner radius of 22.5% of
its side. On a dark taskbar the tile nearly disappears and the ink carries the
icon; on a light one the tile reads as a dark squircle.

## Don't

- Don't recolor the symbol — not to gold, not to blurple, not to Signal Blue.
- Don't use Plex's or Plexamp's marks as PlexCord's.
- Don't add a gradient. SIGNAL bans orange→indigo gradients specifically; the
  mark bans all of them.
- Don't join the three masses with a connecting line. The gaps are the drawing.
- Don't rebuild the mark by hand at a new size — use the SVG, or the small
  geometry under 32px.
- Don't put the badge on anything but the tray icon. It is a transient state,
  not part of the mark.
