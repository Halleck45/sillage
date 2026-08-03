#!/usr/bin/env python3
"""Génère les icônes du site à partir du logo Sillage, redessiné en pixel art.

Le logo d'origine (logo-sillage.png) est une image générée : sa grille de pixels
est irrégulière et elle porte une frange magenta. On le redessine donc ici sur
une grille de 16x16, ce qui donne un rendu net à toutes les tailles et reste
lisible dans un onglet.

    python3 tools/make-favicon.py assets
"""

import sys
from pathlib import Path

from PIL import Image, ImageDraw

DARK = (18, 34, 32)       # #122220, le carré de fond
CREAM = (240, 234, 225)   # #f0eae1, la voile et la coque
TEAL = (120, 153, 154)    # #78999a, le sillage

RADIUS_RATIO = 0.20       # rayon des coins, en fraction du côté

# '.' fond, '#' bateau, '~' sillage
ART = [
    "................",
    "................",
    ".......##.......",
    ".......###......",
    ".......####.....",
    ".......#####....",
    ".......######...",
    ".......#........",
    "~~.....#......~~",
    ".~~.########.~~.",
    "..~~.######.~~..",
    "...~~......~~...",
    "....~~....~~....",
    ".....~~..~~.....",
    "......~~~~......",
    "................",
]

N = len(ART)
assert all(len(row) == N for row in ART), "la grille doit être carrée"

COLOR = {"#": CREAM, "~": TEAL}


def hexa(rgb):
    return "#%02x%02x%02x" % rgb


def runs():
    """Suites horizontales de cellules de même couleur, pour limiter les formes."""
    for y, row in enumerate(ART):
        x = 0
        while x < N:
            c = row[x]
            if c == ".":
                x += 1
                continue
            x0 = x
            while x < N and row[x] == c:
                x += 1
            yield x0, y, x - x0, c


def to_svg():
    # Le PNG est découpé par un masque arrondi : on découpe le SVG pareil, pour
    # que les deux formats restent identiques même si le dessin touche un coin.
    out = [
        '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" '
        'shape-rendering="crispEdges">' % (N, N),
        '  <defs><clipPath id="corners"><rect width="%d" height="%d" rx="%.2f"/>'
        "</clipPath></defs>" % (N, N, RADIUS_RATIO * N),
        '  <g clip-path="url(#corners)">',
        '    <rect width="%d" height="%d" fill="%s"/>' % (N, N, hexa(DARK)),
    ]
    for x, y, w, c in runs():
        out.append(
            '    <rect x="%d" y="%d" width="%d" height="1" fill="%s"/>'
            % (x, y, w, hexa(COLOR[c]))
        )
    out += ["  </g>", "</svg>", ""]
    return "\n".join(out)


def render(size, supersample=8):
    """Aplats nets, coins arrondis lissés par suréchantillonnage."""
    big = size * supersample
    cell = big / N
    im = Image.new("RGBA", (big, big), DARK + (255,))
    draw = ImageDraw.Draw(im)
    for x, y, w, c in runs():
        draw.rectangle(
            [round(x * cell), round(y * cell),
             round((x + w) * cell) - 1, round((y + 1) * cell) - 1],
            fill=COLOR[c] + (255,),
        )
    mask = Image.new("L", (big, big), 0)
    ImageDraw.Draw(mask).rounded_rectangle(
        [0, 0, big - 1, big - 1], radius=round(RADIUS_RATIO * big), fill=255
    )
    im.putalpha(mask)
    return im.resize((size, size), Image.LANCZOS)


def main():
    dest = Path(sys.argv[1] if len(sys.argv) > 1 else "assets")
    dest.mkdir(parents=True, exist_ok=True)

    (dest / "favicon.svg").write_text(to_svg())
    render(180).save(dest / "apple-touch-icon.png")
    render(512).save(dest / "icon-512.png")

    ico_sizes = [16, 32, 48]
    images = [render(s) for s in ico_sizes]
    images[-1].save(
        dest / "favicon.ico",
        format="ICO",
        sizes=[(s, s) for s in ico_sizes],
        append_images=images[:-1],
    )

    for name in ("favicon.svg", "favicon.ico", "apple-touch-icon.png", "icon-512.png"):
        print("%-22s %6d o" % (name, (dest / name).stat().st_size))


if __name__ == "__main__":
    main()
