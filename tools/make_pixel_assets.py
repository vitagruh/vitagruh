# -*- coding: utf-8 -*-
"""Генератор пиксельных SVG-ассетов для профиля vitagruh."""
import json
import os
import random
import urllib.request

OUT = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "assets")

FONT = {
    "A": [".###.", "#...#", "#...#", "#####", "#...#", "#...#", "#...#"],
    "B": ["####.", "#...#", "#...#", "####.", "#...#", "#...#", "####."],
    "C": [".###.", "#...#", "#....", "#....", "#....", "#...#", ".###."],
    "D": ["####.", "#...#", "#...#", "#...#", "#...#", "#...#", "####."],
    "E": ["#####", "#....", "#....", "####.", "#....", "#....", "#####"],
    "F": ["#####", "#....", "#....", "####.", "#....", "#....", "#...."],
    "G": [".###.", "#...#", "#....", "#.###", "#...#", "#...#", ".###."],
    "H": ["#...#", "#...#", "#...#", "#####", "#...#", "#...#", "#...#"],
    "I": [".###.", "..#..", "..#..", "..#..", "..#..", "..#..", ".###."],
    "J": ["..###", "...#.", "...#.", "...#.", "...#.", "#..#.", ".##.."],
    "K": ["#...#", "#..#.", "#.#..", "##...", "#.#..", "#..#.", "#...#"],
    "L": ["#....", "#....", "#....", "#....", "#....", "#....", "#####"],
    "M": ["#...#", "##.##", "#.#.#", "#...#", "#...#", "#...#", "#...#"],
    "N": ["#...#", "##..#", "#.#.#", "#..##", "#...#", "#...#", "#...#"],
    "O": [".###.", "#...#", "#...#", "#...#", "#...#", "#...#", ".###."],
    "P": ["####.", "#...#", "#...#", "####.", "#....", "#....", "#...."],
    "Q": [".###.", "#...#", "#...#", "#...#", "#.#.#", "#..#.", ".##.#"],
    "R": ["####.", "#...#", "#...#", "####.", "#.#..", "#..#.", "#...#"],
    "S": [".####", "#....", "#....", ".###.", "....#", "....#", "####."],
    "T": ["#####", "..#..", "..#..", "..#..", "..#..", "..#..", "..#.."],
    "U": ["#...#", "#...#", "#...#", "#...#", "#...#", "#...#", ".###."],
    "V": ["#...#", "#...#", "#...#", "#...#", "#...#", ".#.#.", "..#.."],
    "W": ["#...#", "#...#", "#...#", "#...#", "#.#.#", "##.##", "#...#"],
    "X": ["#...#", "#...#", ".#.#.", "..#..", ".#.#.", "#...#", "#...#"],
    "Y": ["#...#", "#...#", ".#.#.", "..#..", "..#..", "..#..", "..#.."],
    "Z": ["#####", "....#", "...#.", "..#..", ".#...", "#....", "#####"],
    "0": [".###.", "#...#", "#..##", "#.#.#", "##..#", "#...#", ".###."],
    "1": ["..#..", ".##..", "..#..", "..#..", "..#..", "..#..", ".###."],
    "2": [".###.", "#...#", "....#", "...#.", "..#..", ".#...", "#####"],
    "3": ["####.", "....#", "....#", ".###.", "....#", "....#", "####."],
    "4": ["...#.", "..##.", ".#.#.", "#..#.", "#####", "...#.", "...#."],
    "5": ["#####", "#....", "####.", "....#", "....#", "#...#", ".###."],
    "6": [".###.", "#....", "#....", "####.", "#...#", "#...#", ".###."],
    "7": ["#####", "....#", "...#.", "..#..", ".#...", ".#...", ".#..."],
    "8": [".###.", "#...#", "#...#", ".###.", "#...#", "#...#", ".###."],
    "9": [".###.", "#...#", "#...#", ".####", "....#", "....#", ".###."],
    " ": [".....", ".....", ".....", ".....", ".....", ".....", "....."],
    ".": [".....", ".....", ".....", ".....", ".....", ".....", "..#.."],
    ":": [".....", "..#..", ".....", ".....", "..#..", ".....", "....."],
    "/": ["....#", "....#", "...#.", "..#..", ".#...", "#....", "#...."],
    "-": [".....", ".....", ".....", ".###.", ".....", ".....", "....."],
    "!": ["..#..", "..#..", "..#..", "..#..", "..#..", ".....", "..#.."],
    "?": [".###.", "#...#", "....#", "...#.", "..#..", ".....", "..#.."],
    "+": [".....", "..#..", "..#..", "#####", "..#..", "..#..", "....."],
    "(": ["...#.", "..#..", ".#...", ".#...", ".#...", "..#..", "...#."],
    ")": [".#...", "..#..", "...#.", "...#.", "...#.", "..#..", ".#..."],
    ",": [".....", ".....", ".....", ".....", "..#..", "..#..", ".#..."],
    "'": ["..#..", "..#..", ".....", ".....", ".....", ".....", "....."],
    "<": ["...#.", "..#..", ".#...", "#....", ".#...", "..#..", "...#."],
    ">": [".#...", "..#..", "...#.", "....#", "...#.", "..#..", ".#..."],
    "=": [".....", ".....", "#####", ".....", "#####", ".....", "....."],
    "_": [".....", ".....", ".....", ".....", ".....", ".....", "#####"],
    "%": ["##..#", "##.#.", "...#.", "..#..", ".#...", ".#.##", "#..##"],
    "#": [".#.#.", ".#.#.", "#####", ".#.#.", "#####", ".#.#.", ".#.#."],
    "*": [".....", "#.#.#", ".###.", "#####", ".###.", "#.#.#", "....."],
}

HEART = [
    ".##.##.",
    "#######",
    "#######",
    ".#####.",
    "..###..",
    "...#...",
]

SKULL = [
    ".....######.....",
    "...##########...",
    "..############..",
    ".##############.",
    ".###..####..###.",
    ".###..####..###.",
    ".##############.",
    ".##############.",
    "..#####..#####..",
    "...##########...",
    "....##.##.##....",
    "....########....",
    ".....#.##.#.....",
    "......####......",
    ".......##.......",
    "................",
]


def text_width(text, s):
    return (len(text) * 6 - 1) * s


def text_path(text, x, y, s):
    parts = []
    cx = x
    for ch in text.upper():
        glyph = FONT.get(ch, FONT[" "])
        for r, row in enumerate(glyph):
            for c, v in enumerate(row):
                if v == "#":
                    parts.append("M{} {}h{}v{}h-{}z".format(cx + c * s, y + r * s, s, s, s))
        cx += 6 * s
    return "".join(parts)


def text(parts, value, x, y, s, fill, outline=None, off=None, extra=""):
    if outline:
        o = off if off is not None else max(1, s // 4)
        for dx, dy in ((-o, 0), (o, 0), (0, -o), (0, o)):
            parts.append('<path d="{}" fill="{}"/>'.format(text_path(value, x + dx, y + dy, s), outline))
    parts.append('<path d="{}" fill="{}"{} />'.format(text_path(value, x, y, s), fill, extra))


def ctext(parts, value, cx, y, s, fill, outline=None, off=None):
    text(parts, value, cx - text_width(value, s) // 2, y, s, fill, outline, off)


def rect(x, y, w, h, fill, extra=""):
    return '<rect x="{}" y="{}" width="{}" height="{}" fill="{}"{} />'.format(x, y, w, h, fill, extra)


def frame(w, h, parts, stroke="#3a4048", bg="#06080b"):
    parts.append(rect(0, 0, w, h, bg))
    parts.append(rect(4, 4, w - 8, h - 8, "none", ' stroke="#1b2127" stroke-width="4"'))
    parts.append(rect(10, 10, w - 20, h - 20, "none", ' stroke="{}" stroke-width="4"'.format(stroke)))


def star_field(parts, w, h, count=70, seed=1337):
    rnd = random.Random(seed)
    for _ in range(count):
        x = rnd.randrange(10, w - 10)
        y = rnd.randrange(10, h - 60)
        size = rnd.choice([2, 2, 2, 3])
        base = round(rnd.uniform(0.08, 0.5), 2)
        dur = round(rnd.uniform(3.2, 9.0), 1)
        begin = -round(rnd.uniform(0, 6), 1)
        parts.append(
            '<rect x="{0}" y="{1}" width="{2}" height="{3}" fill="#8b949e" opacity="{4}">'
            '<animate attributeName="opacity" values="0.05;{4};0.14;{4};0.05" dur="{5}s" begin="{6}s" repeatCount="indefinite"/>'
            "</rect>".format(x, y, size, size, base, dur, begin)
        )


def city(parts, w, h, floor_y, seed=7):
    rnd = random.Random(seed)
    x = 0
    while x < w:
        bw = rnd.choice([24, 28, 32, 40])
        bh = rnd.randrange(20, 54)
        parts.append(rect(x, floor_y - bh, bw, bh, "#10151a"))
        parts.append(rect(x, floor_y - bh, bw, 3, "#2b343d"))
        if rnd.random() < 0.35:
            ax = x + bw // 2
            parts.append(rect(ax, floor_y - bh - 14, 2, 14, "#2b343d"))
            parts.append(
                '<rect x="{}" y="{}" width="4" height="4" fill="#7a2b2b">'
                '<animate attributeName="opacity" values="1;0.1;1" dur="3.6s" begin="-{}s" repeatCount="indefinite"/>'
                "</rect>".format(ax - 1, floor_y - bh - 18, round(rnd.uniform(0, 2), 1))
            )
        for wy in range(floor_y - bh + 12, floor_y - 10, 12):
            for wx in range(x + 6, x + bw - 6, 11):
                if rnd.random() < 0.42:
                    dur = round(rnd.uniform(4, 11), 1)
                    parts.append(
                        '<rect x="{0}" y="{1}" width="4" height="4" fill="#9a7838" opacity="0.45">'
                        '<animate attributeName="opacity" values="0.06;0.45;0.06" dur="{2}s" begin="-{3}s" repeatCount="indefinite"/>'
                        "</rect>".format(wx, wy, dur, round(rnd.uniform(0, 5), 1))
                    )
        x += bw + rnd.choice([2, 4, 6])


def planet(x, y, s=5):
    body = ["..###..", ".#####.", "###.###", "#######", "##.####", ".#####.", "..###.."]
    parts = [
        '<g><animateTransform attributeName="transform" type="translate" values="0 0;0 -4;0 0" dur="9s" repeatCount="indefinite"/>',
        '<rect x="{}" y="{}" width="{}" height="6" fill="#2b3138"/>'.format(x - 10, y + s * 4, s * 7 + 20),
    ]
    for r, row in enumerate(body):
        for c, v in enumerate(row):
            if v == "#":
                color = "#8f8d82" if (r, c) in ((2, 2), (4, 4)) else "#c2beb0"
                parts.append(rect(x + c * s, y + r * s, s, s, color))
    parts.append(rect(x + 2 * s, y + 1 * s, s, s, "#d8d4c6"))
    parts.append("</g>")
    return "".join(parts)


def scanlines(w, h):
    return (
        '<pattern id="scan" width="4" height="4" patternUnits="userSpaceOnUse">'
        '<rect width="4" height="1" fill="#000000" opacity="0.3"/></pattern>'
        '<rect width="{}" height="{}" fill="url(#scan)" opacity="0.6"/>'.format(w, h)
    )


def fog(parts, w, seed=5):
    rnd = random.Random(seed)
    bands = [(108, 30, 0.14, 21, -4), (168, 26, 0.11, 26, -12), (232, 40, 0.09, 33, -8)]
    for y, hh, op, dur, beg in bands:
        parts.append('<g><animateTransform attributeName="transform" type="translate" values="-90 0;90 0;-90 0" dur="{}s" begin="{}s" repeatCount="indefinite"/>'.format(dur, beg))
        for _ in range(110):
            fx = rnd.randrange(0, w)
            fy = y + rnd.randrange(0, hh)
            fw = rnd.choice([2, 2, 4, 6, 8])
            parts.append(rect(fx, fy, fw, 2, "#5a6670", ' opacity="{}"'.format(round(op * rnd.uniform(0.45, 1.0), 2))))
        parts.append("</g>")


def make_banner():
    w, h = 1200, 340
    p = []
    p.append('<svg width="{0}" height="{1}" viewBox="0 0 {0} {1}" xmlns="http://www.w3.org/2000/svg" role="img" aria-label="vitagruh — offensive security">'.format(w, h))
    p.append(
        "<defs>"
        '<linearGradient id="sky" x1="0" y1="0" x2="0" y2="1">'
        '<stop offset="0%" stop-color="#020305"/><stop offset="60%" stop-color="#06080b"/><stop offset="100%" stop-color="#0a0d11"/>'
        "</linearGradient>"
        '<pattern id="dots" width="6" height="6" patternUnits="userSpaceOnUse"><rect width="1" height="1" fill="#141a20" opacity="0.35"/></pattern>'
        '<linearGradient id="dim" x1="0" y1="0" x2="0" y2="1">'
        '<stop offset="0%" stop-color="#020305" stop-opacity="0"/>'
        '<stop offset="18%" stop-color="#020305" stop-opacity="0.55"/>'
        '<stop offset="82%" stop-color="#020305" stop-opacity="0.55"/>'
        '<stop offset="100%" stop-color="#020305" stop-opacity="0"/>'
        "</linearGradient>"
        + "</defs>"
    )
    p.append(rect(0, 0, w, h, "url(#sky)"))
    star_field(p, w, h, 80)
    p.append(planet(1082, 96, 7))
    city(p, w, h, h - 26)
    fog(p, w)
    p.append(rect(0, 182, w, 124, "url(#dim)"))
    p.append(rect(0, h - 26, w, 26, "#05070a"))
    p.append(rect(0, h - 26, w, 2, "#1a2129"))
    for gx in range(0, w, 30):
        p.append(rect(gx, h - 24, 14, 2, "#141a20", ' opacity="0.7"'))

    # HUD
    text(p, "1UP", 40, 34, 3, "#8b8f8a")
    text(p, "01337", 40, 68, 3, "#7f8a72")
    text(p, "HI 09000", w - 40 - text_width("HI 09000", 3), 34, 3, "#6d7275")

    # title
    tw = text_width("VITAGRUH", 12)
    x0 = (w - tw) // 2
    p.append('<path d="{}" fill="#241111"/>'.format(text_path("VITAGRUH", x0 + 8, 128, 12)))
    text(p, "VITAGRUH", x0, 120, 12, "#d9d6c7", outline="#040404", off=5)
    # subtitle
    ctext(p, "OFFENSIVE SECURITY // RED TEAM", w // 2, 222, 3, "#8d8f86")

    # press start
    p.append('<g opacity="1"><animate attributeName="opacity" values="1;1;0;0" keyTimes="0;0.58;0.59;1" dur="1.6s" repeatCount="indefinite"/>')
    text(p, "PRESS START", w // 2 - text_width("PRESS START", 3) // 2, 258, 3, "#b08a4a")
    p.append("</g>")

    # bobbing skulls
    for i, x in enumerate((1114, 1046, 978)):
        p.append(
            '<g><animateTransform attributeName="transform" type="translate" values="0 0;0 -10;0 0" dur="{}s" begin="-{}s" repeatCount="indefinite"/>'
            '<animate attributeName="opacity" values="0.85;1;0.8;1;0.85" dur="2.4s" begin="-{}s" repeatCount="indefinite"/>'.format(
                3.0 + i * 0.4, i * 0.7, i * 0.5
            )
        )
        for r, row in enumerate(SKULL):
            for c, v in enumerate(row):
                if v == "#":
                    p.append(rect(x + c * 2, 236 + r * 2, 2, 2, "#a8a498"))
        for ex in (4, 10):
            p.append(rect(x + ex * 2, 236 + 4 * 2, 4, 4, "#a63a3a"))
        p.append("</g>")

    # hearts
    for i in range(3):
        hx, hy = 40 + i * 34, 266
        hs = 4
        for r, row in enumerate(HEART):
            for c, v in enumerate(row):
                if v == "#":
                    p.append(rect(hx + c * hs, hy + r * hs, hs, hs, "#7e2b2b"))

    p.append(scanlines(w, h))
    p.append(
        '<rect x="0" y="-30" width="{}" height="30" fill="#9fb2a6" opacity="0.03">'
        '<animateTransform attributeName="transform" type="translate" from="0 -30" to="0 {}" dur="9s" repeatCount="indefinite"/>'
        "</rect>".format(w, h + 30)
    )
    p.append("</svg>")
    return "".join(p)


def make_player(repos, stars, followers, top_lang):
    w, h = 1200, 400
    p = []
    p.append('<svg width="{0}" height="{1}" viewBox="0 0 {0} {1}" xmlns="http://www.w3.org/2000/svg" role="img" aria-label="player card">'.format(w, h))
    frame(w, h, p)
    p.append(rect(18, 18, w - 36, h - 36, "none", ' stroke="#141a20" stroke-width="2"'))

    text(p, "PLAYER 1", 50, 38, 3, "#a8894f")

    # skull sprite
    s = 8
    sx, sy = 44, 84
    p.append(rect(sx - 12, sy - 12, 16 * s + 24, 16 * s + 24, "#0b0d10", ' stroke="#2a3038" stroke-width="2"'))
    skull = []
    for r, row in enumerate(SKULL):
        for c, v in enumerate(row):
            if v == "#":
                skull.append(rect(sx + c * s, sy + r * s, s, s, "#cfcbbd"))
    skull.append('<g><animate attributeName="opacity" values="1;1;0;1;1" keyTimes="0;0.45;0.5;0.55;1" dur="5.5s" repeatCount="indefinite"/>')
    for ex, ey in ((4, 4), (10, 4)):
        skull.append(rect(sx + ex * s, sy + ey * s, 2 * s, 2 * s, "#a63a3a", ' opacity="0.85"'))
    skull.append("</g>")
    p.append("".join(skull))

    # identity
    text(p, "VITAGRUH", 240, 76, 8, "#e2ded0", outline="#050505", off=3)
    text(p, "CLASS: OFFENSIVE SEC", 240, 168, 3, "#8d8f86")
    text(p, "LOADOUT: PYTHON / JS", 240, 200, 3, "#8d8f86")
    text(p, "GUILD: SOLO", 240, 232, 3, "#8d8f86")

    # bars
    def bar(x, y, wpx, ratio, color, label):
        text(p, label, 240, y - 6, 3, "#9aa0a0")
        bx = 300
        p.append(rect(bx, y, wpx, 18, "#090b0e", ' stroke="#2a3038" stroke-width="2"'))
        seg = 12
        gap = 3
        count = max(1, round(ratio * (wpx - 8 + gap) / (seg + gap)))
        for i in range(count):
            xx = bx + 4 + i * (seg + gap)
            if xx + seg <= bx + wpx - 4:
                p.append(rect(xx, y + 4, seg, 10, color))

    bar(240, 272, 300, 0.86, "#8c3a3a", "HP")
    bar(240, 314, 300, 0.72, "#4a6b7a", "MP")

    # score panel
    p.append(rect(700, 60, 460, 300, "#090b0e", ' stroke="#2a3038" stroke-width="2"'))
    text(p, "SCORE", 724, 80, 4, "#a8894f")
    rows = [
        ("REPOS", str(repos)),
        ("STARS", str(stars)),
        ("FANS", str(followers)),
        ("LANG", top_lang),
    ]
    y = 150
    for i, (label, value) in enumerate(rows):
        text(p, ">", 724, y, 3, "#7f8a72")
        text(p, label, 756, y, 3, "#6f7577")
        vx = 1132 - text_width(value, 4 if len(value) < 8 else 3)
        text(p, value, vx, y - 2, 4 if len(value) < 8 else 3, "#e2ded0")
        y += 56

    text(p, "SELECT = USE    BACK = CANCEL", 44, 346, 2, "#4a514f")
    text(p, "199X VITAGRUH", 44, 366, 2, "#333a38")
    p.append("</svg>")
    return "".join(p)


def chip(x, y, label, pips, total=5):
    w, h = 126, 72
    s = []
    s.append(rect(x, y, w, h, "#0f1216", ' stroke="#454c55" stroke-width="3"'))
    s.append(rect(x + 6, y + 6, w - 12, h - 12, "none", ' stroke="#1b2127" stroke-width="1"'))
    size = 3 if text_width(label, 3) + 16 < w else 2
    ctext(s, label, x + w // 2, y + 18, size, "#d5d2c6")
    pip_w = 10
    gap = 4
    start = x + (w - (total * pip_w + (total - 1) * gap)) // 2
    for i in range(total):
        color = "#5c6b57" if i < pips else "#1a1f24"
        s.append(rect(start + i * (pip_w + gap), y + 50, pip_w, 8, color))
    return "".join(s)


def make_skills():
    w, h = 1200, 220
    p = []
    p.append('<svg width="{0}" height="{1}" viewBox="0 0 {0} {1}" xmlns="http://www.w3.org/2000/svg" role="img" aria-label="skill tree">'.format(w, h))
    frame(w, h, p)
    text(p, "SKILL TREE", 44, 40, 4, "#a8894f")
    skills = [
        ("PY", 5), ("JS", 4), ("SQL", 4), ("FLASK", 4),
        ("DOCKER", 3), ("LINUX", 4), ("GIT", 4), ("TG", 5),
    ]
    x = 44
    for i, (label, pips) in enumerate(skills):
        p.append(chip(x, 106, label, pips))
        x += 140 if i < 7 else 0
        if i == 2 or i == 5:
            x += 4
    p.append("</svg>")
    return "".join(p)


def make_heading(title, subtitle=None):
    size = 4
    w = 1200
    h = 76
    p = []
    p.append('<svg width="{0}" height="{1}" viewBox="0 0 {0} {1}" xmlns="http://www.w3.org/2000/svg" role="img">'.format(w, h))
    p.append(rect(0, 0, w, h, "#05070a"))
    p.append(rect(0, 0, w, 2, "#252b31"))
    p.append(rect(0, h - 2, w, 2, "#252b31"))
    tw = text_width(title, size)
    x0 = (w - tw) // 2
    p.append(rect(x0 - 44, h // 2 - 2, 24, 4, "#5c6b57"))
    p.append(rect(x0 + tw + 20, h // 2 - 2, 24, 4, "#5c6b57"))
    text(p, title, x0, 24, size, "#d5d2c6")
    p.append("</svg>")
    return "".join(p)


def make_footer():
    w, h = 1200, 200
    p = []
    p.append('<svg width="{0}" height="{1}" viewBox="0 0 {0} {1}" xmlns="http://www.w3.org/2000/svg" role="img" aria-label="continue?">'.format(w, h))
    p.append(rect(0, 0, w, h, "#040507"))
    star_field(p, w, h, 30, seed=99)
    fog(p, w)

    label = "CONTINUE?"
    size = 5
    total = text_width(label, size) + 24 + text_width("0", 8)
    start = (w - total) // 2
    text(p, label, start, 40, size, "#d5d2c6", outline="#050505", off=2)

    # countdown 9..0
    dx = start + text_width(label, size) + 24
    for i in range(10):
        digit = str(9 - i)
        p.append('<g opacity="0">')
        p.append(
            '<animate attributeName="opacity" values="0;0;1;1;0;0" keyTimes="0;0.001;0.01;0.09;0.091;1" dur="10s" begin="{}s" repeatCount="indefinite"/>'.format(i)
        )
        text(p, digit, dx, 30, 8, "#a63a3a")
        p.append("</g>")

    p.append('<g><animate attributeName="opacity" values="1;1;0;0" keyTimes="0;0.6;0.61;1" dur="1.6s" repeatCount="indefinite"/>')
    ctext(p, "INSERT COIN", w // 2, 104, 3, "#a8894f")
    p.append("</g>")
    ctext(p, "(C) 199X VITAGRUH", w // 2, 146, 2, "#4a514f")
    ctext(p, "UP UP DOWN DOWN LEFT RIGHT LEFT RIGHT B A", w // 2, 174, 2, "#2a3136")
    p.append("</svg>")
    return "".join(p)


def fetch_stats():
    req = urllib.request.Request("https://api.github.com/users/vitagruh", headers={"User-Agent": "pixel-assets"})
    with urllib.request.urlopen(req, timeout=15) as r:
        user = json.load(r)
    req = urllib.request.Request("https://api.github.com/users/vitagruh/repos?per_page=100", headers={"User-Agent": "pixel-assets"})
    with urllib.request.urlopen(req, timeout=15) as r:
        repos = json.load(r)
    stars = sum(x.get("stargazers_count", 0) for x in repos)
    langs = {}
    for x in repos:
        if x.get("language"):
            langs[x["language"]] = langs.get(x["language"], 0) + 1
    top = max(langs, key=langs.get) if langs else "PYTHON"
    return user.get("public_repos", len(repos)), stars, user.get("followers", 0), top.upper()


def main():
    os.makedirs(OUT, exist_ok=True)
    try:
        repos, stars, followers, top = fetch_stats()
    except Exception as exc:  # offline fallback
        print("api fallback:", exc)
        repos, stars, followers, top = 10, 2, 0, "PYTHON"
    files = {
        "pixel-banner.svg": make_banner(),
        "pixel-player.svg": make_player(repos, stars, followers, top),
        "pixel-skills.svg": make_skills(),
        "pixel-footer.svg": make_footer(),
        "pixel-heading-quests.svg": make_heading("QUEST LOG"),
        "pixel-heading-objective.svg": make_heading("OBJECTIVE"),
        "pixel-heading-highscore.svg": make_heading("HIGH SCORE"),
    }
    for name, data in files.items():
        path = os.path.join(OUT, name)
        with open(path, "w", encoding="utf-8", newline="\n") as f:
            f.write(data)
        print(name, len(data))


if __name__ == "__main__":
    main()
