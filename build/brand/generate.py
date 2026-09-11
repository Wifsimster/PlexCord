# -*- coding: utf-8 -*-
"""Regenerates PlexCord's brand assets from the canonical geometry.

Run from the repository root: python3 build/brand/generate.py
Documented in docs/brand.md. Requires Pillow.

Canonical geometry, 64x64 box:
  regular (>= 48 px): dash 15x9 r4.5 at x=3 and x=46, dot r8.5 centered
  small   (<= 32 px): dash 13x11 r5.5 at x=2 and x=49, dot r10.5
Small is an optical size, not a style: at 16 px the regular drawing's masses
fall below a pixel and the motif closes up into a smudge.
"""
import io, struct, pathlib
from PIL import Image, ImageDraw

REPO = pathlib.Path(__file__).resolve().parents[2]
TILE_BG  = (0x17, 0x18, 0x1D, 255)   # --pc-surface-850
INK      = (0xED, 0xED, 0xF0, 255)   # --pc-text (dark)
BADGE_INK = (0xF5, 0xB9, 0x44, 255)  # --pc-warn: "something wants your attention"
CLEAR    = (0, 0, 0, 0)

# Update badge, in fractions of the tile side. The dot sits far enough into the
# corner to clear the symbol's right dash (which ends at 0.574 vertically), and
# its transparent gap stays inside the tile's rounded silhouette, so the badge
# never leaves a crescent hanging off the corner.
#
# Like the symbol it is sized optically rather than proportionally: a badge that
# still registers at 16px would dominate the mark at 128px.
BADGE_C = 0.78
BADGE = {
    'regular': dict(gap=0.145, r=0.105),
    'small':   dict(gap=0.190, r=0.150),
}
RADIUS_R = 0.225                     # tile corner radius, as a fraction of its side
SS       = 8                         # supersampling factor

GEOM = {
    'regular': dict(dash=(3.0, 27.5, 15.0, 9.0, 4.5), dot=(32.0, 32.0, 8.5), box=0.74),
    'small':   dict(dash=(2.0, 26.5, 13.0, 11.0, 5.5), dot=(32.0, 32.0, 10.5), box=0.86),
}

def draw_mark(d, size, variant, color, offset=(0, 0)):
    """Draw the symbol centered in a square of side `size`."""
    g = GEOM[variant]
    box = size * g['box']
    s = box / 64.0
    ox = offset[0] + (size - box) / 2.0
    oy = offset[1] + (size - box) / 2.0
    x, y, w, h, r = g['dash']
    for dx in (x, 64.0 - x - w):
        d.rounded_rectangle([ox + dx * s, oy + y * s, ox + (dx + w) * s, oy + (y + h) * s],
                            radius=r * s, fill=color)
    cx, cy, rad = g['dot']
    d.ellipse([ox + (cx - rad) * s, oy + (cy - rad) * s,
               ox + (cx + rad) * s, oy + (cy + rad) * s], fill=color)

def icon(size, variant=None, badge=False):
    variant = variant or ('small' if size <= 32 else 'regular')
    big = size * SS
    img = Image.new('RGBA', (big, big), (0, 0, 0, 0))
    d = ImageDraw.Draw(img)
    d.rounded_rectangle([0, 0, big - 1, big - 1], radius=big * RADIUS_R, fill=TILE_BG)
    draw_mark(d, big, variant, INK)
    if badge:
        # The gap is erased rather than filled with the tile color, so the badge
        # reads the same whatever the icon is sitting on.
        b = BADGE[variant]
        cx = cy = big * BADGE_C
        for r, fill in ((big * b['gap'], CLEAR), (big * b['r'], BADGE_INK)):
            d.ellipse([cx - r, cy - r, cx + r, cy + r], fill=fill)
    return img.resize((size, size), Image.LANCZOS)

# ---------------------------------------------------------------- appicon.png
icon(1024).save(REPO / 'build/appicon.png', 'PNG', optimize=True)
# Tray variant shown while an update is waiting to be applied.
icon(1024, badge=True).save(REPO / 'build/appicon-update.png', 'PNG', optimize=True)

# ---------------------------------------------------------------- icon.ico
def ico(path, sizes, badge=False):
    """Write an .ico with PNG payloads — one image per size, optical variant included."""
    payloads = []
    for s in sizes:
        buf = io.BytesIO()
        icon(s, badge=badge).save(buf, 'PNG', optimize=True)
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
ico(REPO / 'build/windows/icon-update.ico', [16, 24, 32, 48, 64, 128, 256], badge=True)

# ---------------------------------------------------------------- SVG
def mark_paths(variant, color='currentColor'):
    g = GEOM[variant]
    x, y, w, h, r = g['dash']
    cx, cy, rad = g['dot']
    return (f'<rect x="{x}" y="{y}" width="{w}" height="{h}" rx="{r}" fill="{color}"/>\n  '
            f'<circle cx="{cx}" cy="{cy}" r="{rad}" fill="{color}"/>\n  '
            f'<rect x="{64 - x - w}" y="{y}" width="{w}" height="{h}" rx="{r}" fill="{color}"/>')

BRAND = REPO / 'build/brand'
HEAD = ('<?xml version="1.0" encoding="UTF-8"?>\n'
        '<!-- PlexCord symbol. See docs/brand.md. -->\n')

(BRAND / 'plexcord-mark.svg').write_text(
    HEAD + '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" width="64" height="64" '
    'role="img" aria-label="PlexCord">\n  ' + mark_paths('regular') + '\n</svg>\n', encoding='utf-8')

(BRAND / 'plexcord-mark-small.svg').write_text(
    HEAD + '<!-- Optical size: use at 32px and below. -->\n'
    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" width="32" height="32" '
    'role="img" aria-label="PlexCord">\n  ' + mark_paths('small') + '\n</svg>\n', encoding='utf-8')

def icon_svg(variant, side):
    g = GEOM[variant]
    box = 64 * g['box']
    off = (64 - box) / 2.0
    scale = box / 64.0
    inner = mark_paths(variant, '#EDEDF0')
    return (HEAD + f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" '
            f'width="{side}" height="{side}" role="img" aria-label="PlexCord">\n'
            f'  <rect width="64" height="64" rx="{64 * RADIUS_R}" fill="#17181D"/>\n'
            f'  <g transform="translate({off:.3f} {off:.3f}) scale({scale:.5f})">\n    '
            + inner.replace('\n  ', '\n    ') + '\n  </g>\n</svg>\n')

(BRAND / 'plexcord-icon.svg').write_text(icon_svg('regular', 512), encoding='utf-8')

def icon_update_svg(side):
    base = icon_svg('regular', side).split('\n', 2)[2].rstrip()
    body = base[base.index('>') + 1:base.rindex('</svg>')]
    c = 64 * BADGE_C
    gap, r = 64 * BADGE['regular']['gap'], 64 * BADGE['regular']['r']
    return (HEAD + '<!-- Tray icon while an update waits to be applied. -->\n'
            f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" width="{side}" '
            f'height="{side}" role="img" aria-label="PlexCord — update ready">\n'
            f'  <mask id="badge-gap">\n'
            f'    <rect width="64" height="64" fill="#fff"/>\n'
            f'    <circle cx="{c}" cy="{c}" r="{gap}" fill="#000"/>\n'
            f'  </mask>\n'
            f'  <g mask="url(#badge-gap)">{body}  </g>\n'
            f'  <circle cx="{c}" cy="{c}" r="{r}" fill="#F5B944"/>\n</svg>\n')

(BRAND / 'plexcord-icon-update.svg').write_text(icon_update_svg(512), encoding='utf-8')
(BRAND / 'plexcord-icon-small.svg').write_text(icon_svg('small', 32), encoding='utf-8')

FONT_UI = ("-apple-system,BlinkMacSystemFont,'Segoe UI Variable','Segoe UI',system-ui,"
           "Roboto,'Helvetica Neue',Arial,sans-serif")

(BRAND / 'plexcord-lockup.svg').write_text(
    HEAD + '<!-- Horizontal lockup. The wordmark is a system stack: it composes with the\n'
    '     font of the host, as in the app (no webfonts, see design-system.md). -->\n'
    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 220 48" width="220" height="48" '
    'role="img" aria-label="PlexCord">\n'
    '  <g transform="translate(0 8) scale(0.5)">\n    '
    + mark_paths('regular', '#EDEDF0').replace('\n  ', '\n    ') + '\n  </g>\n'
    f'  <text x="45" y="31" font-family="{FONT_UI}" font-size="26" font-weight="600" '
    'letter-spacing="-0.52" fill="#EDEDF0">PlexCord</text>\n</svg>\n', encoding='utf-8')

print('brand assets written')
