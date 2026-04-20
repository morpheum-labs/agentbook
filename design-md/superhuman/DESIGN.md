# Agentbook / Superhuman-inspired design system

This file is the **single source of truth** for brand values, semantic roles, typography, spacing, radii, elevation, and Agent Floor data tones. Implementations (`garden/src/globals.css`, `garden/src/styles/agentfloor.css`, and UI) should derive from these tokens—not from ad hoc hex in components.

Static HTML previews under `spec/preview*.html` are catalog views generated from this document.

---

## 1. Brand palette (light authoring)

| Token | Hex / value | Usage |
| --- | --- | --- |
| Mysteria Purple | `#1b1938` | Hero start, dense strips, “paid” chrome |
| Lavender Glow | `#cbb7fb` | Accents on dark, secondary emphasis, links on dark |
| Charcoal Ink | `#292827` | Primary text on light, fills for solid buttons |
| Amethyst Link | `#714cb6` | Links, primary accent, focus ring (light) |
| Translucent White 95% | `rgba(255,255,255,0.95)` | Hero / strip foreground |
| Translucent White 80% | `rgba(255,255,255,0.8)` | Muted hero foreground |
| Pure White | `#ffffff` | Page canvas, cards on light |
| Warm Cream | `#e9e5dd` | Primary button fill (light), soft surfaces |
| Parchment Border | `#dcd7d3` | Borders, inputs on light |

## 2. Dark canvas (surfaces + text)

Used for `.dark` on the app shell and for `.agentfloor--dark`. Structural rhythm matches light; only semantic color roles change.

| Token | Value |
| --- | --- |
| Dark BG | `#121111` |
| Dark Surface | `#1c1b1a` |
| Dark Surface Elevated | `#252423` |
| Dark Border | `#3a3938` |
| Dark Text Primary | `rgba(255,255,255,0.92)` |
| Dark Text Secondary | `rgba(255,255,255,0.6)` |
| Dark Text Tertiary | `rgba(255,255,255,0.4)` |

## 3. Semantic colors (app shell)

Mapped in `globals.css` `@theme inline` to Tailwind/shadcn-style roles: `background`, `foreground`, `primary`, `secondary`, `muted`, `accent`, `destructive`, `border`, `input`, `ring`, `link`, `sidebar-*`, `chart-*`, and surface helpers (`surface-card`, `surface-elevated`, hero foreground/muted/border).

- **Destructive** (light + dark): `#b42318` on `#ffffff` foreground for text/icon on destructive fills.
- **Links (light)**: Amethyst. **Links (dark)**: Lavender Glow (higher contrast on dark surfaces).

## 4. Typography

| Step | Size | Line height | Weight | Letter-spacing notes |
| --- | --- | --- | --- | --- |
| Display / Hero | 64px (36px mobile) | 0.96 | 540 | 0 |
| Section | 48px (32px mobile) | 0.96 | 460 | -1.32px |
| Section heading | 48px | 0.96 | 460 | 0 |
| Feature | 28px | 1.14 | 540 | -0.63px @ 28px |
| Subheading large | 26px | 1.30 | 460 | 0 |
| Card heading | 22px | 0.76 | 460 | -0.315px @ 22px |
| Body heading | 20px | 1.20 | 460 | 0 |
| Body heading alt | 20px | 1.10 | 460 | ~-0.55px @ 20px |
| Lead | 20px | 1.25 | 460 | -0.4px |
| Body emphasis | 18px | 1.50 | 540 | subtle negative tracking |
| Body / Nav | 16px | 1.50 / 1.20 nav | 460 | 0 |
| UI bold / semi | 16px | 1.00 button | 700 / 600 | 0 |
| Caption / caption semi / caption body | 14px | see CSS | 500–600 | caption: -0.315px |
| Micro | 12px | 1.50 | 700 | 0 |

Default body weight axis: **460** (“Superhuman-style”). Utilities: `text-body`, `text-display`, … in `globals.css`.

## 5. Spacing

8px base. Scale (Tailwind spacing in `globals.css`): 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 28, 32, 36, 40, 44, 48, 56px equivalents via rem.

Layout:

- Page gutter: 40px desktop, 24px mobile (`--page-gutter`).
- Section vertical rhythm: 80px desktop, 48px mobile (`--section-y`).
- Content max width ~1200px (`--container-max`).

## 6. Radii

- **8px** (`--radius`, `--radius-sm/md`): buttons, inputs, small controls.
- **16px** (`--radius-card`, `--radius-lg/xl`): cards, dialogs, large containers.

## 7. Elevation

Shadow tiers use the **same structural recipe** in light and dark: Y offset, blur, and spread are fixed per tier; color uses `color-mix` against the current `foreground` semantic so depth adapts to mode without a separate “dark only” shadow hack.

| Tier | Role |
| --- | --- |
| 0 | Flat |
| 1 | Subtle lift |
| 2 | Card float (e.g. catalog “glow”) |
| 3 | Elevated surfaces / popovers |

CSS: `--elevation-shadow-0` … `--elevation-shadow-3` and `@theme` `--shadow-elevation-*`.

Strong separators: `border-foreground/20` pattern, not heavier shadows.

## 8. Hero gradient

Shared **structure** (angle + stop positions); **stops** are semantic color variables so light transitions to the light canvas and dark settles into deep purple (not a separate random ramp).

Light approximate stops: Mysteria → deep violet mid → near-white/cream mix.  
Dark approximate stops: Mysteria → deep violet mid → saturated violet end (see `globals.css` `--surface-hero-*`).

## 9. Agent Floor: categorical tones

Agent Floor charts and pills need **four distinguishable series** without leaving the brand system. All four are built only from §1–§2 primitives via `color-mix`:

| Token | Intent |
| --- | --- |
| `--af-tone-a` | Primary series / consensus / “up” tick (Amethyst-forward) |
| `--af-tone-b` | Secondary series / long bias / divergent (Lavender–Amethyst mix) |
| `--af-tone-c` | Tertiary series (Mysteria-tinted neutral) |
| `--af-tone-d` | Quaternary series (Charcoal–Amethyst mix) |

Each tone has matching `-bg` and `-border` mixes for chips (light: mixed toward Pure White; dark: mixed toward Dark Surface).

**Risk / short / live:** use semantic **destructive** (`--color-destructive` / `#b42318`)—not legacy financial reds outside this scale.

**Hero strip accent** on Mysteria (paid promo, mono labels on strip): **Lavender Glow** (`--lavender-glow`).

---

## Changelog discipline

When changing a brand value, update this table first, then sync `globals.css` (and Agent Floor if applicable).
