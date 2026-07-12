---
name: BegireX
colors:
  surface: '#0b1326'
  surface-dim: '#0b1326'
  surface-bright: '#31394d'
  surface-container-lowest: '#060e20'
  surface-container-low: '#131b2e'
  surface-container: '#171f33'
  surface-container-high: '#222a3d'
  surface-container-highest: '#2d3449'
  on-surface: '#dae2fd'
  on-surface-variant: '#c7c4d7'
  inverse-surface: '#dae2fd'
  inverse-on-surface: '#283044'
  outline: '#908fa0'
  outline-variant: '#464554'
  surface-tint: '#c0c1ff'
  primary: '#c0c1ff'
  on-primary: '#1000a9'
  primary-container: '#8083ff'
  on-primary-container: '#0d0096'
  inverse-primary: '#494bd6'
  secondary: '#b9c8de'
  on-secondary: '#233143'
  secondary-container: '#39485a'
  on-secondary-container: '#a7b6cc'
  tertiary: '#ffb783'
  on-tertiary: '#4f2500'
  tertiary-container: '#d97721'
  on-tertiary-container: '#452000'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#e1e0ff'
  primary-fixed-dim: '#c0c1ff'
  on-primary-fixed: '#07006c'
  on-primary-fixed-variant: '#2f2ebe'
  secondary-fixed: '#d4e4fa'
  secondary-fixed-dim: '#b9c8de'
  on-secondary-fixed: '#0d1c2d'
  on-secondary-fixed-variant: '#39485a'
  tertiary-fixed: '#ffdcc5'
  tertiary-fixed-dim: '#ffb783'
  on-tertiary-fixed: '#301400'
  on-tertiary-fixed-variant: '#703700'
  background: '#0b1326'
  on-background: '#dae2fd'
  surface-variant: '#2d3449'
typography:
  display:
    fontFamily: Instrument Sans
    fontSize: 24px
    fontWeight: '600'
    lineHeight: 32px
    letterSpacing: -0.02em
  headline:
    fontFamily: Instrument Sans
    fontSize: 18px
    fontWeight: '600'
    lineHeight: 24px
  body-lg:
    fontFamily: Instrument Sans
    fontSize: 16px
    fontWeight: '400'
    lineHeight: 24px
  body-sm:
    fontFamily: Instrument Sans
    fontSize: 14px
    fontWeight: '400'
    lineHeight: 20px
  label-mono:
    fontFamily: JetBrains Mono
    fontSize: 12px
    fontWeight: '500'
    lineHeight: 16px
  caption:
    fontFamily: Instrument Sans
    fontSize: 12px
    fontWeight: '500'
    lineHeight: 16px
    letterSpacing: 0.01em
rounded:
  sm: 0.125rem
  DEFAULT: 0.25rem
  md: 0.375rem
  lg: 0.5rem
  xl: 0.75rem
  full: 9999px
spacing:
  unit: 4px
  space-xs: 4px
  space-sm: 8px
  space-md: 16px
  space-lg: 24px
  space-xl: 48px
  layout-margin: 32px
  layout-gutter: 16px
---

## Brand & Style
The design system focuses on utility and surgical precision. It follows a **Minimalist Modern** aesthetic, stripping away all unnecessary visual noise to prioritize the content—specifically the media download queue and technical configuration. The goal is to make a powerful CLI tool feel approachable yet professional.

The UI relies on structural integrity rather than decorative elements. It uses a **Dark-first** approach to reduce eye strain during long-running background tasks. The emotional response is one of efficiency and control; the interface should feel like a high-end terminal translated into a clean, graphical space.

## Colors
The palette is rooted in a "Dark Graphite" scale. The background uses a deep ink-blue to provide more depth than pure black. 

- **Primary**: A pale periwinkle (`primary` `#c0c1ff`) used sparingly for action buttons and active progress indicators.
- **Secondary**: A muted slate blue-gray (`secondary` `#b9c8de`) for secondary text and non-critical icons.
- **Neutral/Surface**: Layered `surface-container-*` shades from ink-blue to slate to define hierarchy.
- **Light Mode**: Deferred — token set above is dark-only. If added later, mirror the same role names (surface, on-surface, primary, etc.) rather than introducing new ones.

Ensure all color combinations meet WCAG AA contrast ratios for accessibility.

## Typography
The system uses **Instrument Sans** for all UI elements — a grotesk with more character than the default system stack, still built for dense legibility at small sizes. It scales from high-contrast headlines for view titles to dense, readable body text for file paths and metadata.

A secondary monospaced font, **JetBrains Mono**, is utilized for technical labels, version numbers, and the log output, providing a clear visual distinction between "UI labels" and "system data." 

**Formatting Rules:**
- Use `display` for main view titles.
- Use `body-sm` for the majority of the application interface.
- Use `label-mono` for file sizes, formats (e.g., MP4, MKV), and CLI flags.

## Layout & Spacing
This design system uses a **Fluid Layout** with a logical property focus (`margin-inline`, `padding-block`) to support RTL languages out of the box. 

- **Grid**: A flexible 12-column system is used for settings pages, while the main queue uses a single-column list layout.
- **Rhythm**: Spacing follows a 4px baseline. Most interactive elements are separated by `space-md` (16px).
- **Desktop Strategy**: The sidebar is fixed at 240px, while the main content area expands. On narrower windows, the sidebar collapses into a rail.
- **Safe Zones**: Ensure a minimum `layout-margin` of 32px on large displays to prevent content from feeling cramped against the window edges.

## Elevation & Depth
In alignment with the WebKitGTK-safe constraint, the design system avoids heavy backdrop filters or complex blurs. Instead, it utilizes **Tonal Layers** to establish hierarchy.

- **Base Layer**: The darkest surface (`surface-container-lowest` `#060e20`).
- **Mid Layer**: Input fields and secondary containers use a slightly lighter shade (`surface-container` `#171f33`) to appear inset or "carved" into the base.
- **Top Layer**: Active cards, modals, and popovers use the lightest surface (`surface-container-high` `#222a3d`) with a 1px solid `outline-variant` border for definition.
- **Focus**: High-visibility 2px solid rings in `primary` (`#c0c1ff`) are used for keyboard navigation and active focus states.

## Shapes
The design system adopts a **Soft** shape language. This provides a modern touch without appearing overly "bubbly" or consumer-grade.

- **Buttons & Inputs**: 0.25rem (4px) corner radius.
- **Cards & Queue Items**: 0.5rem (8px) corner radius.
- **Progress Bars**: Fully rounded (pill-shaped) to represent "fluid" movement.

## Components

### Progressive Disclosure
- **Expandable Inputs**: The main URL input is large and centered. Upon pasting a link, it shrinks and moves to the top, revealing the format selection and download options below.
- **Advanced Settings**: Technical yt-dlp flags are hidden behind a "chevron" toggle to keep the main view clean.

### Download Queue
- **List Items**: Each item displays a thumbnail (fixed aspect ratio), title, and a compact progress bar.
- **Status Indicators**: Use small, high-contrast labels (e.g., "Downloading", "Merging", "Completed") using the `label-mono` typography level.

### Buttons
- **Primary**: Solid `primary` background (`#c0c1ff`) with `on-primary` text (`#1000a9`).
- **Secondary**: Ghost style with a 1px `outline-variant` border (`#464554`).
- **States**: Hover states should slightly lighten the background color; active states should depress the element by 1px via transform.

### Input Fields
- **Search/URL**: High-contrast borders that transition from `outline-variant` to `primary` on focus.
- **Checkboxes**: Custom-styled squares with a checkmark, avoiding native browser/OS styling to maintain the dark-mode aesthetic.

### Progress Bars
- Background: `surface-container`.
- Fill: `primary`.
- Height: 4px for standard items, 8px for the "active" download.