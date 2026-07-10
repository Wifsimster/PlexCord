#!/usr/bin/env python3
"""Generate the PlexCord app icon from the BrandMark glyph.

Renders the Plexamp polyline glyph (see frontend/src/components/BrandMark.vue)
in the Plex brand pigment (#e5a00d) on the app-canvas dark surface (#0b0c0f),
then writes build/appicon.png (1024x1024) and build/windows/icon.ico.

Supersamples 4x and downscales with LANCZOS for smooth, anti-aliased strokes.
"""

import os

from PIL import Image, ImageDraw

# --- Brand constants (kept in sync with frontend tokens) --------------------
PLEX = (0xE5, 0xA0, 0x0D, 0xFF)   # --pc-plex
CANVAS = (0x0B, 0x0C, 0x0F, 0xFF)  # --pc-surface-950 (window background)

# BrandMark glyph: viewBox 0 0 48 48, stroke-width 2.5, round caps/joins.
POINTS = [
    (4.5, 24), (23.444, 24), (12.808, 9.342), (16.883, 9.342),
    (27.519, 24), (16.883, 38.658), (20.957, 38.658), (31.594, 24),
    (20.957, 9.342), (25.032, 9.342), (35.668, 24), (25.032, 38.658),
    (29.107, 38.658), (39.743, 24), (43.5, 24),
]
VIEWBOX = 48.0
STROKE = 2.5

FINAL = 1024
SS = 4                       # supersample factor
S = FINAL * SS
GLYPH_FRACTION = 0.64        # glyph box relative to icon
CORNER_RADIUS = 0.185        # rounded-square radius relative to icon


def build(size, ss):
    work = size * ss
    img = Image.new("RGBA", (work, work), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)

    # Rounded-square canvas background.
    radius = int(CORNER_RADIUS * work)
    draw.rounded_rectangle([0, 0, work - 1, work - 1], radius=radius, fill=CANVAS)

    # Map the 48-unit viewBox into a centered box of GLYPH_FRACTION of the icon.
    box = GLYPH_FRACTION * work
    scale = box / VIEWBOX
    offset = (work - box) / 2.0
    pts = [(offset + x * scale, offset + y * scale) for x, y in POINTS]
    width = max(1, int(round(STROKE * scale)))

    # Rounded joins/caps: draw the stroke, then discs at every vertex so both
    # the segment joints and the two open ends read as round (SVG stroke-round).
    draw.line(pts, fill=PLEX, width=width, joint="curve")
    r = width / 2.0
    for x, y in pts:
        draw.ellipse([x - r, y - r, x + r, y + r], fill=PLEX)

    return img.resize((size, size), Image.LANCZOS)


def main():
    here = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    png = build(FINAL, SS)
    png_path = os.path.join(here, "build", "appicon.png")
    png.save(png_path, "PNG")
    print("wrote", png_path)

    ico_path = os.path.join(here, "build", "windows", "icon.ico")
    sizes = [16, 24, 32, 48, 64, 128, 256]
    png.save(ico_path, "ICO", sizes=[(s, s) for s in sizes])
    print("wrote", ico_path)


if __name__ == "__main__":
    main()
