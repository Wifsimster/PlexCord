# PlexCord Brand

The interface has its own canonical spec — [`design-system.md`](design-system.md),
"ON AIR". This document covers the layer above it: what PlexCord is as a brand,
its mark, its one color, and the rules that keep it from colliding with the two
brands it sits between.

## The idea: PlexCord puts you on air

PlexCord plays nothing and stores nothing. It takes what you are playing on
Plex and **broadcasts** it to your Discord profile. The brand is built on that
one verb. A broadcast studio has exactly one light that everybody understands
without reading a word: the **ON AIR** lamp. PlexCord's whole identity is that
lamp, and everything else is kept quiet so the lamp means something.

It replaces "SIGNAL" (v1.14 – v4.3), whose neutral three-dot mark and
blue-for-interaction contract were correct but anonymous. SIGNAL looked like
every other dark developer tool, its blue accent sat right next to Discord
blurple, and the mark was too small to register as a symbol. ON AIR keeps
SIGNAL's discipline (saturation = meaning, honest state, one state in one
place). It changes the voice:

| | SIGNAL | ON AIR |
| --- | --- | --- |
| Metaphor | an instrument panel for a relay | a broadcast desk: you are on air, or you are not |
| Product color | none (neutral mark) | **Tally rose** `#FF4D8D`, the on-air lamp |
| Interaction | Signal Blue | **Ink**: the text color, filled |
| Neutrals | cool blue-grey | warm "studio graphite" |
| Type | system UI only | **Bricolage Grotesque** for display + system UI for reading |
| Mark | dash · dot · dash | **the live badge**: a ring with a presence badge |

## The mark: the live badge

A thick ring, with a disc in its bottom-right corner, separated from the ring by
an erased gap. The disc carries a play triangle.

Read one way, the ring is a **record** and the disc a play button. Read the
other way, the ring is a **profile picture** and the disc is a **presence
badge**, the little dot Discord puts on the corner of your avatar. That second
reading is exactly what PlexCord does: it gives your profile a "now playing"
badge. The symbol is the product's output, drawn.

The ring is **ink**. The badge is the **tally**. That is the only place the
tally appears in the mark, and it is lit because the product's job is to be on
air.

**Unlit** — the ring with no badge — is the mark "off air". The app uses it as
the dashboard's idle artwork: nothing is on air, so the badge is gone.

### Two optical sizes

The mark is drawn twice, picked by size, not by taste:

| Geometry | Use at | Ring (64-unit box) | Badge | Gap | Play glyph | Glyph box |
| --- | --- | --- | --- | --- | --- | --- |
| **Regular** | 32px and above | c(29,29) R25, hole r10 | c(47,47) r13 | r17.5 | yes | 80% of the tile |
| **Small** | below 32px | c(28,28) R26, hole r9 | c(46,46) r14 | r19 | no | 92% of the tile |

Below 32px, the regular drawing's play triangle falls under a pixel and its
thinner ring starts to fill in. The small geometry thickens the ring, grows the
badge and drops the triangle, so at 16px what survives is exactly the idea:
*a ring with a lit dot on its corner.* The in-app lockup (20px) uses it.

### The update badge

While an update waits to be applied, the **tray** icon carries an amber dot
(`--pc-warn`) in its **top-right** corner. The live badge owns the
bottom-right, so the update dot takes the free one. Everything else is
unchanged from SIGNAL:

- **The gap around the dot is erased, not filled with the tile color**, so it
  reads the same on any taskbar.
- **The dot is sized optically**: `r` 0.100 of the tile at 32px and above,
  0.150 below it, at (0.80, 0.20).
- It is a transient state and never appears anywhere but the tray.

### Clear space and minimum size

- **Clear space**: the badge's diameter, on all four sides.
- **Minimum size**: 16px for the symbol. Below 20px, drop the wordmark.

## The one color: tally

`--pc-tally` is `#FF4D8D` on dark and `#D0195E` on light (5.8:1 and 5.3:1 on
their panels). It is a rose picked for the one hue region nobody else on screen
uses:

- **Plex gold** (~40°) and **Discord blurple** (~235°) are the endpoints.
- **Green / amber / vermilion** (~150° / 40° / 10°) are connection health. The
  danger red moved to a vermilion `#FF7155` so it could never be mistaken for
  the tally.
- The tally sits at ~340°, clear of all of them.

It means one thing, **"Discord is showing your media right now"**, and it is
spent only on that: the lit tally pill, the progress fill of what's on air, the
mark's badge, and the faint glow behind the relay on the Welcome step. It is
never a button, a link, a focus ring or a selection. Those are **ink**.

## The wordmark

"PlexCord" in **Bricolage Grotesque** at weight 700 and `-0.025em` tracking.
Bricolage is bundled with the app (`@fontsource-variable/bricolage-grotesque`,
OFL-1.1, ~130 KB latin), so the lockup is the same object on every OS. Its ink
traps and slightly grotesque squareness give the name a voice the system UI
face never had. It is used for display only: the wordmark, page titles, the
stage headline, the tally label. Everything read in volume stays in the
platform's own UI face.

The lockup is horizontal: symbol, then a gap of 40–45% of the symbol's height,
then the word. The optional uppercase suffix (`SETUP`) is a third element in
`--pc-text-micro`, muted. It is a mode label, not part of the mark.

## Files

| File | What it is |
| --- | --- |
| `build/brand/generate.py` | Regenerates everything below from the canonical geometry |
| `build/brand/plexcord-mark.svg` | Symbol, regular geometry (ring `currentColor`, badge tally) |
| `build/brand/plexcord-mark-small.svg` | Symbol, small optical geometry |
| `build/brand/plexcord-icon.svg` | App icon master: tile + symbol |
| `build/brand/plexcord-icon-small.svg` | App icon, small optical geometry |
| `build/brand/plexcord-lockup.svg` | Symbol + wordmark |
| `build/brand/plexcord-icon-update.svg` | App icon with the update badge |
| `build/appicon.png` | 1024px app icon (macOS, Linux, tray) |
| `build/appicon-update.png` | Badged tray icon (macOS, Linux) |
| `build/windows/icon.ico` | 16/24/32/48/64/128/256, small geometry below 32 |
| `build/windows/icon-update.ico` | Badged tray icon (Windows), same sizes |
| `docs/images/banner.svg` | README banner |
| `frontend/src/components/BrandSymbol.vue` | The symbol in-app (both optical sizes, `unlit`) |
| `frontend/src/components/BrandMark.vue` | The in-app lockup |

The icon tile is `--pc-surface-850` (`#171514`) with a corner radius of 22.5%
of its side.

## Don't

- Don't put the tally anywhere it doesn't mean "on air": no tally buttons,
  links, selections, headings or decorative fills.
- Don't recolor the ring: it is ink (`--pc-text`), or `currentColor` in a
  single-color context. Never gold, blurple or tally.
- Don't recolor the badge: it is the tally, or absent (unlit). Never green, even
  though "online" dots are green elsewhere.
- Don't close the gap between ring and badge. The erased gap is what makes the
  badge read as a badge.
- Don't use Plex's or Discord's marks as PlexCord's.
- Don't add gradients to the mark. The one sanctioned glow is the tally's own
  radial halo behind the relay, on the Welcome step and the banner.
- Don't rebuild the mark by hand at a new size. Use the SVG or `BrandSymbol`,
  and the small geometry below 32px.
