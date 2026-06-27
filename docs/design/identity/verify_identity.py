#!/usr/bin/env python3
"""Heimdall visual-identity audit.

Recomputes every foreground/background contrast pair from palette.json, checks
ANSI-256 fallbacks are present, resolves all W3C DTCG references in tokens.json,
and verifies palette <-> tokens colour agreement for the brand/border roles.

Run from the repo root:
    python3 docs/design/identity/verify_identity.py

Exit status 0 = all checks pass; 1 = at least one failure.
"""
import json
import os
import re
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
PALETTE = os.path.join(HERE, "palette.json")
TOKENS = os.path.join(HERE, "tokens.json")
TYPO = os.path.join(HERE, "typography.json")

REF_RE = re.compile(r"^\{([A-Za-z0-9_.]+)\}$")


# ---- WCAG contrast -------------------------------------------------------
def relative_luminance(hex_color):
    h = hex_color.lstrip("#")
    r, g, b = (int(h[i:i + 2], 16) / 255 for i in (0, 2, 4))
    lin = lambda c: c / 12.92 if c <= 0.04045 else ((c + 0.055) / 1.055) ** 2.4
    r, g, b = lin(r), lin(g), lin(b)
    return 0.2126 * r + 0.7152 * g + 0.0722 * b


def contrast_ratio(fg, bg):
    l1, l2 = relative_luminance(fg), relative_luminance(bg)
    hi, lo = max(l1, l2), min(l1, l2)
    return (hi + 0.05) / (lo + 0.05)


def floor_for(aa):
    return 4.5 if aa == "text" else 3.0


# ---- checks --------------------------------------------------------------
def check_contrast(palette):
    fails, total = [], 0
    for tname, theme in palette["themes"].items():
        surf = {k: v["hex"] for k, v in theme["surface"].items()}
        for gname in ("structure", "text", "brand", "semantic"):
            for role, spec in theme.get(gname, {}).items():
                if not isinstance(spec, dict) or "hex" not in spec or "on" not in spec:
                    continue
                fl = floor_for(spec.get("aa", "text"))
                for bgn in spec["on"]:
                    total += 1
                    ratio = contrast_ratio(spec["hex"], surf[bgn])
                    if ratio < fl:
                        fails.append(f"{tname} {gname}.{role} {spec['hex']} on {bgn} "
                                     f"= {ratio:.2f}:1 < {fl:g}:1")
        sev = theme.get("severity", {})
        for stop in (sev.get("ramp", []) if isinstance(sev, dict) else []):
            fl = floor_for(stop.get("aa", "large"))
            for bgn in stop.get("on", ["background"]):
                total += 1
                ratio = contrast_ratio(stop["hex"], surf[bgn])
                if ratio < fl:
                    fails.append(f"{tname} severity.{stop.get('name')} {stop['hex']} "
                                 f"on {bgn} = {ratio:.2f}:1 < {fl:g}:1")
        # focus uses reverse video on the signature fill
        total += 1
        ratio = contrast_ratio(surf["background"], theme["brand"]["signature"]["hex"])
        if ratio < 4.5:
            fails.append(f"{tname} focus(reverse) = {ratio:.2f}:1 < 4.5:1")
    return total, fails


def check_ansi(palette):
    missing = []

    def walk(node, path):
        if isinstance(node, dict):
            if "hex" in node and "ansi256" not in node:
                missing.append(path)
            for k, v in node.items():
                walk(v, f"{path}.{k}")
        elif isinstance(node, list):
            for i, v in enumerate(node):
                walk(v, f"{path}[{i}]")

    walk(palette["themes"], "themes")
    return missing


def resolve(path, tokens, seen=None):
    seen = seen or set()
    if path in seen:
        return None  # cycle
    seen.add(path)
    node = tokens
    for part in path.split("."):
        if not isinstance(node, dict) or part not in node:
            return None
        node = node[part]
    val = node.get("$value") if isinstance(node, dict) else None
    if isinstance(val, str):
        m = REF_RE.match(val)
        if m:
            return resolve(m.group(1), tokens, seen)
        return val
    if isinstance(val, dict) and isinstance(val.get("color"), str):
        m = REF_RE.match(val["color"])
        if m:
            return resolve(m.group(1), tokens, seen) or val
    return val if val is not None else (node if isinstance(node, dict) else None)


def check_refs(tokens):
    dangling = []

    def walk(node):
        if isinstance(node, dict):
            for v in node.values():
                walk(v)
        elif isinstance(node, list):
            for v in node:
                walk(v)
        elif isinstance(node, str):
            m = REF_RE.match(node)
            if m and resolve(m.group(1), tokens) is None:
                dangling.append(node)

    walk(tokens)
    return sorted(set(dangling))


def check_drift(palette, tokens):
    """Brand + border colours must agree between palette.json and tokens.json."""
    d = palette["themes"]["dark"]
    l = palette["themes"]["light"]
    checks = [
        ("color.signature", d["brand"]["signature"]["hex"]),
        ("color.steel", d["brand"]["steel"]["hex"]),
        ("color.border", d["structure"]["border"]["hex"]),
        ("color.light.signature", l["brand"]["signature"]["hex"]),
        ("color.light.steel", l["brand"]["steel"]["hex"]),
        ("color.light.border", l["structure"]["border"]["hex"]),
    ]
    out = []
    for ref, expect in checks:
        got = resolve(ref, tokens)
        got = got.lower() if isinstance(got, str) else got
        if got != expect.lower():
            out.append(f"{ref}: tokens={got} palette={expect.lower()}")
    return out


def statuses():
    out = {}
    for name, path in (("palette", PALETTE), ("tokens", TOKENS), ("typography", TYPO)):
        out[name] = json.load(open(path)).get("_meta", {}).get("status")
    return out


def main():
    palette = json.load(open(PALETTE))
    tokens = json.load(open(TOKENS))
    json.load(open(TYPO))  # parse check

    total, cfails = check_contrast(palette)
    missing = check_ansi(palette)
    dangling = check_refs(tokens)
    drift = check_drift(palette, tokens)
    st = statuses()

    ok = not (cfails or missing or dangling or drift)
    print(f"contrast pairs : {total} checked, {len(cfails)} fail")
    for f in cfails:
        print("  FAIL", f)
    print(f"ansi256 fallback: {'all present' if not missing else str(len(missing)) + ' missing'}")
    for m in missing:
        print("  MISSING", m)
    print(f"dtcg refs      : {'all resolve' if not dangling else str(len(dangling)) + ' dangling'}")
    for r in dangling:
        print("  DANGLING", r)
    print(f"palette<->tokens: {'aligned' if not drift else str(len(drift)) + ' mismatch'}")
    for r in drift:
        print("  DRIFT", r)
    print(f"status         : palette={st['palette']} tokens={st['tokens']} typography={st['typography']}")
    print("RESULT:", "PASS" if ok else "FAIL")
    sys.exit(0 if ok else 1)


if __name__ == "__main__":
    main()
