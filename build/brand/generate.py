# -*- coding: utf-8 -*-
"""Regenerates PlexCord's brand assets from the canonical geometry.

Run from the repository root: python3 build/brand/generate.py
Documented in docs/brand.md. Requires Pillow.

The symbol is "the live badge": a ring (a record, or a profile picture) in
ink, with a presence badge on its corner in the tally. Canonical geometry,
64x64 box (mirrored by frontend/src/components/BrandSymbol.vue):

  regular (>= 32 px): ring c(29,29) R25 r10, badge c(47,47) r13, erased gap
                      r17.5, play triangle knocked out of the badge
  small   (<  32 px): ring c(28,28) R26 r9,  badge c(46,46) r14, erased gap
                      r19, solid badge (no triangle)

Small is an optical size, not a style: at 16 px the regular drawing's play
triangle falls below a pixel and its thinner ring starts to fill in.
"""
import io, struct, pathlib
from PIL import Image, ImageDraw

REPO = pathlib.Path(__file__).resolve().parents[2]
TILE_BG   = (0x17, 0x15, 0x14, 255)  # --pc-surface-850
INK       = (0xF2, 0xEE, 0xE8, 255)  # --pc-text (dark)
TALLY     = (0xFF, 0x4D, 0x8D, 255)  # --pc-tally-400
BADGE_INK = (0xF5, 0xB9, 0x44, 255)  # --pc-warn: "something wants your attention"

RADIUS_R = 0.225                      # tile corner radius, as a fraction of its side
SS       = 8                          # supersampling factor

GEOM = {
    'regular': dict(ring=(29.0, 29.0, 25.0, 10.0), badge=(47.0, 47.0, 13.0), gap=17.5,
                    play=((43.2, 40.6), (53.4, 47.0), (43.2, 53.4)), box=0.80),
    'small':   dict(ring=(28.0, 28.0, 26.0, 9.0), badge=(46.0, 46.0, 14.0), gap=19.0,
                    play=None, box=0.92),
}

# Update badge (tray only), in fractions of the tile side. The live badge owns
# the bottom-right corner, so the update dot takes the top-right one. Sized
# optically, like the symbol: one that registers at 16px would dominate 128px.
UPDATE_C = (0.80, 0.20)
UPDATE = {
    'regular': dict(gap=0.140, r=0.100),
    'small':   dict(gap=0.190, r=0.150),
}


def _circle(d, cx, cy, r, fill):
    d.ellipse([cx - r, cy - r, cx + r, cy + r], fill=fill)


def symbol_layer(side, variant, ink=INK, tally=TALLY, box=None):
    """The symbol alone on a transparent square of `side` px (already supersampled)."""
    g = GEOM[variant]
    box = side * (g['box'] if box is None else box)
    s = box / 64.0
    o = (side - box) / 2.0
    P = lambda v: o + v * s

    # Ring: outer disc minus hole minus the badge's erased gap, as an alpha mask.
    ring = Image.new('L', (side, side), 0)
    rd = ImageDraw.Draw(ring)
    cx, cy, R, r = g['ring']
    _circle(rd, P(cx), P(cy), R * s, 255)
    _circle(rd, P(cx), P(cy), r * s, 0)
    bx, by, br = g['badge']
    _circle(rd, P(bx), P(by), g['gap'] * s, 0)

    # Badge: tally disc, with the play triangle knocked out (regular only).
    badge = Image.new('L', (side, side), 0)
    bd = ImageDraw.Draw(badge)
    _circle(bd, P(bx), P(by), br * s, 255)
    if g['play']:
        pts = [(P(x), P(y)) for x, y in g['play']]
        bd.polygon(pts, fill=0)
        bd.line(pts + [pts[0]], fill=0, width=max(1, int(2 * s)), joint='curve')

    out = Image.new('RGBA', (side, side), (0, 0, 0, 0))
    out.paste(Image.new('RGBA', (side, side), ink), (0, 0), ring)
    out.paste(Image.new('RGBA', (side, side), tally), (0, 0), badge)
    return out


def icon(size, variant=None, update=False):
    variant = variant or ('small' if size < 32 else 'regular')
    big = size * SS
    tile = Image.new('L', (big, big), 0)
    ImageDraw.Draw(tile).rounded_rectangle([0, 0, big - 1, big - 1], radius=big * RADIUS_R, fill=255)
    img = Image.new('RGBA', (big, big), (0, 0, 0, 0))
    img.paste(Image.new('RGBA', (big, big), TILE_BG), (0, 0), tile)
    img.alpha_composite(symbol_layer(big, variant))
    if update:
        # The gap is erased rather than filled with the tile color, so the badge
        # reads the same whatever the icon is sitting on.
        u = UPDATE[variant]
        cx, cy = big * UPDATE_C[0], big * UPDATE_C[1]
        hole = Image.new('L', (big, big), 255)
        _circle(ImageDraw.Draw(hole), cx, cy, big * u['gap'], 0)
        alpha = Image.composite(img.getchannel('A'), Image.new('L', (big, big), 0), hole)
        img.putalpha(alpha)
        _circle(ImageDraw.Draw(img), cx, cy, big * u['r'], BADGE_INK)
    return img.resize((size, size), Image.LANCZOS)


# ---------------------------------------------------------------- appicon.png
icon(1024).save(REPO / 'build/appicon.png', 'PNG', optimize=True)
# Tray variant shown while an update is waiting to be applied.
icon(1024, update=True).save(REPO / 'build/appicon-update.png', 'PNG', optimize=True)


# ---------------------------------------------------------------- icon.ico
def ico(path, sizes, update=False):
    """Write an .ico with PNG payloads — one image per size, optical variant included."""
    payloads = []
    for s in sizes:
        buf = io.BytesIO()
        icon(s, update=update).save(buf, 'PNG', optimize=True)
        payloads.append((s, buf.getvalue()))
    header = struct.pack('<HHH', 0, 1, len(payloads))
    offset = 6 + 16 * len(payloads)
    entries, blobs = b'', b''
    for s, data in payloads:
        entries += struct.pack('<BBBBHHII', s if s < 256 else 0, s if s < 256 else 0,
                               0, 0, 1, 32, len(data), offset)
        offset += len(data)
        blobs += data
    path.write_bytes(header + entries + blobs)


ico(REPO / 'build/windows/icon.ico', [16, 24, 32, 48, 64, 128, 256])
ico(REPO / 'build/windows/icon-update.ico', [16, 24, 32, 48, 64, 128, 256], update=True)


# ---------------------------------------------------------------- SVG
def ring_d(variant):
    cx, cy, R, r = GEOM[variant]['ring']
    return (f'M{cx} {cy - R}a{R} {R} 0 1 1 0 {2 * R}a{R} {R} 0 1 1 0 {-2 * R}z'
            f'M{cx} {cy - r}a{r} {r} 0 1 0 0 {2 * r}a{r} {r} 0 1 0 0 {-2 * r}z')


def symbol_svg_body(variant, ink, tally, prefix):
    """Masks + shapes of the symbol in the 64-unit box. `prefix` keeps mask ids unique."""
    g = GEOM[variant]
    bx, by, br = g['badge']
    out = (f'<mask id="{prefix}gap"><rect width="64" height="64" fill="#fff"/>'
           f'<circle cx="{bx}" cy="{by}" r="{g["gap"]}" fill="#000"/></mask>\n')
    badge_mask = ''
    if g['play']:
        pts = ' '.join(f'{x} {y}' for x, y in g['play'])
        out += (f'<mask id="{prefix}play"><rect width="64" height="64" fill="#fff"/>'
                f'<path d="M{pts} Z" fill="#000" stroke="#000" stroke-width="2" '
                f'stroke-linejoin="round"/></mask>\n')
        badge_mask = f' mask="url(#{prefix}play)"'
    out += (f'<path mask="url(#{prefix}gap)" fill="{ink}" fill-rule="evenodd" d="{ring_d(variant)}"/>\n'
            f'<circle{badge_mask} cx="{bx}" cy="{by}" r="{br}" fill="{tally}"/>')
    return out


BRAND = REPO / 'build/brand'
HEAD = ('<?xml version="1.0" encoding="UTF-8"?>\n'
        '<!-- PlexCord symbol, "the live badge". See docs/brand.md. -->\n')
INK_HEX, TALLY_HEX, TILE_HEX = '#F2EEE8', '#FF4D8D', '#171514'


def indent(text, n):
    return ('\n' + ' ' * n).join(text.split('\n'))


def mark_svg(variant, side, note=''):
    return (HEAD + note + f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" '
            f'width="{side}" height="{side}" role="img" aria-label="PlexCord">\n  '
            + indent(symbol_svg_body(variant, 'currentColor', TALLY_HEX, 'm-'), 2) + '\n</svg>\n')


(BRAND / 'plexcord-mark.svg').write_text(
    HEAD.replace('-->', '-->\n<!-- Ring in currentColor; the badge is always the tally. -->')
    + mark_svg('regular', 64).split('\n', 2)[2], encoding='utf-8')
(BRAND / 'plexcord-mark-small.svg').write_text(
    mark_svg('small', 32, '<!-- Optical size: use below 32px. -->\n'), encoding='utf-8')


def icon_svg(variant, side, update=False):
    g = GEOM[variant]
    box = 64 * g['box']
    off = (64 - box) / 2.0
    scale = box / 64.0
    body = (f'<rect width="64" height="64" rx="{64 * RADIUS_R}" fill="{TILE_HEX}"/>\n'
            f'<g transform="translate({off:.3f} {off:.3f}) scale({scale:.5f})">\n  '
            + indent(symbol_svg_body(variant, INK_HEX, TALLY_HEX, 'i-'), 2) + '\n</g>')
    label = 'PlexCord — update ready' if update else 'PlexCord'
    note = '<!-- Tray icon while an update waits to be applied. -->\n' if update else ''
    svg = (HEAD + note + f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" '
           f'width="{side}" height="{side}" role="img" aria-label="{label}">\n')
    if not update:
        return svg + '  ' + indent(body, 2) + '\n</svg>\n'
    u = UPDATE[variant]
    cx, cy = 64 * UPDATE_C[0], 64 * UPDATE_C[1]
    return (svg + f'  <mask id="u-gap"><rect width="64" height="64" fill="#fff"/>'
            f'<circle cx="{cx:.2f}" cy="{cy:.2f}" r="{64 * u["gap"]:.2f}" fill="#000"/></mask>\n'
            f'  <g mask="url(#u-gap)">\n    ' + indent(body, 4) + '\n  </g>\n'
            f'  <circle cx="{cx:.2f}" cy="{cy:.2f}" r="{64 * u["r"]:.2f}" fill="#F5B944"/>\n</svg>\n')


(BRAND / 'plexcord-icon.svg').write_text(icon_svg('regular', 512), encoding='utf-8')
(BRAND / 'plexcord-icon-small.svg').write_text(icon_svg('small', 32), encoding='utf-8')
(BRAND / 'plexcord-icon-update.svg').write_text(icon_svg('regular', 512, update=True), encoding='utf-8')

FONT_DISPLAY = ("'Bricolage Grotesque Variable','Bricolage Grotesque',-apple-system,"
                "BlinkMacSystemFont,'Segoe UI Variable','Segoe UI',system-ui,Roboto,Arial,sans-serif")

(BRAND / 'plexcord-lockup.svg').write_text(
    HEAD + '<!-- Horizontal lockup: symbol, a gap of 40% of its height, the wordmark in\n'
    '     Bricolage Grotesque 700 at -0.025em (bundled in the app; falls back to\n'
    '     the system UI face where it is not installed). -->\n'
    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 208 48" width="208" height="48" '
    'role="img" aria-label="PlexCord">\n'
    '  <g transform="translate(0 4) scale(0.625)">\n    '
    + indent(symbol_svg_body('regular', INK_HEX, TALLY_HEX, 'l-'), 4) + '\n  </g>\n'
    f'  <text x="56" y="33" font-family="{FONT_DISPLAY}" font-size="28" font-weight="700" '
    f'letter-spacing="-0.7" fill="{INK_HEX}">PlexCord</text>\n</svg>\n', encoding='utf-8')

print('brand assets written')
