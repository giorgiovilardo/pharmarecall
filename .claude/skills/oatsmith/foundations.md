# Oat - Foundations Reference

## Setup

**CDN:**
```html
<link rel="stylesheet" href="https://unpkg.com/@knadh/oat/oat.min.css">
<script src="https://unpkg.com/@knadh/oat/oat.min.js"></script>
```

**npm:**
```bash
npm install @knadh/oat
```
Then import `@knadh/oat/oat.min.css` and `@knadh/oat/oat.min.js` in your project.

JS is only needed for: tabs, dropdown, toast, tooltip, sidebar, and the dialog `commandfor` polyfill.

## Philosophy

- **Semantic HTML first** - styles native elements (`<button>`, `<dialog>`, `<details>`, `<table>`, `<article>`) rather than generic divs
- **Zero dependencies** - vanilla JS, no build step required
- **Data attributes for color variants** - use `data-variant` (but NOT for sizes or badge variants - see exceptions below)
- **CSS cascade layers** - `theme → base → components → animations → utilities`
- Custom elements use `ot-` prefix (`<ot-tabs>`, `<ot-dropdown>`, `<ot-sidebar>`)
- Custom events use `ot-` prefix (`ot-tab-change`, `ot-toast-close`)

## Dark Mode

Add `data-theme="dark"` to `<body>`. All components adapt automatically.

## CSS Custom Properties

### Colors
| Property | Light | Dark |
|----------|-------|------|
| `--background` | `#fff` | `#09090b` |
| `--foreground` | `#09090b` | `#fafafa` |
| `--card` / `--card-foreground` | `#fff` / `#09090b` | `#18181b` / `#fafafa` |
| `--primary` / `--primary-foreground` | `#574747` / `#fafafa` | `#fafafa` / `#18181b` |
| `--secondary` / `--secondary-foreground` | `#f4f4f5` / `#574747` | `#27272a` / `#fafafa` |
| `--muted` / `--muted-foreground` | `#f4f4f5` / `#71717a` | `#27272a` / `#a1a1aa` |
| `--faint` / `--faint-foreground` | `#fafafa` / `#a1a1aa` | `#1e1e21` / `#71717a` |
| `--accent` | `#f4f4f5` | `#27272a` |
| `--danger` / `--danger-foreground` | `#d32f2f` / `#fafafa` | `#f4807b` / `#18181b` |
| `--success` / `--success-foreground` | `#008032` / `#fafafa` | `#6cc070` / `#18181b` |
| `--warning` / `--warning-foreground` | `#a65b00` / `#09090b` | `#f0a030` / `#09090b` |
| `--border` | `#d4d4d8` | `#52525b` |
| `--input` | `#d4d4d8` | `#52525b` |
| `--ring` | `#574747` | `#d4d4d8` |

### Spacing
`--space-1` (0.25rem), `--space-2` (0.5rem), `--space-3` (0.75rem), `--space-4` (1rem), `--space-5` (1.25rem), `--space-6` (1.5rem), `--space-8` (2rem), `--space-10` (2.5rem), `--space-12` (3rem), `--space-14` (3.5rem), `--space-16` (4rem), `--space-18` (4.5rem)

### Border Radius
`--radius-small` (0.125rem), `--radius-medium` (0.375rem), `--radius-large` (0.75rem), `--radius-full` (9999px)

### Typography
- Fonts: `--font-sans` (system-ui), `--font-mono` (ui-monospace)
- Sizes: `--text-1` (largest, ~2.25rem) through `--text-8` (smallest, 0.75rem). `--text-regular` = `--text-6` (1rem)
- Weights: `--font-normal` (400), `--font-medium` (500), `--font-semibold` (600), `--font-bold` (600, same as semibold)

### Shadows
`--shadow-small`, `--shadow-medium`, `--shadow-large`

### Transitions
`--transition-fast` (120ms), `--transition` (200ms)

### Z-index
`--z-dropdown` (50), `--z-modal` (200)

## Variant Pattern

Components that support color variants use `data-variant="value"`:
- **Buttons**: `secondary`, `danger` (default is primary)
- **Alerts**: `danger` (or `error`), `success`, `warning`
- **Toasts**: `danger`, `success`, `warning`

**Exceptions - these use CSS classes instead of `data-variant`:**
- **Badges**: `.secondary`, `.outline`, `.success`, `.warning`, `.danger`
- **Button appearance**: `.outline`, `.ghost` (combinable with `data-variant` for color)

## Size Pattern

Components that support sizes use CSS classes: `.small`, `.large`

Buttons also support `.icon` for square icon-only buttons.

## Utility Classes

**Layout:** `.flex`, `.flex-col`, `.items-center`, `.justify-center`, `.justify-between`
**Stacks:** `.hstack` (horizontal, centered, gap), `.vstack` (vertical, gap)
**Gap:** `.gap-1`, `.gap-2`, `.gap-4`
**Margin:** `.mt-2`, `.mt-4`, `.mt-6`, `.mb-2`, `.mb-4`, `.mb-6`
**Padding:** `.p-4`
**Width:** `.w-100`
**Text:** `.text-left`, `.text-center`, `.text-right`, `.text-light`, `.text-lighter`
**Accessibility:** `.sr-only`
**Lists:** `ul.unstyled` / `ol.unstyled` (removes bullets/numbers and padding)

## JS Components

| Component | Element | Key API |
|-----------|---------|---------|
| Tabs | `<ot-tabs>` | `ot-tab-change` event |
| Dropdown | `<ot-dropdown>` | Popover API, keyboard nav |
| Sidebar | `<ot-sidebar>` | `data-sidebar-layout`, mobile toggle |
| Toast | - | `ot.toast(message, options)`, `ot.toastEl(element)` |
| Tooltip | - | Auto-enhances elements with `title` attribute |
| Dialog | `<dialog>` | `commandfor`/`command` polyfill for Safari |

## Component Directory

| Component | Description | JS? |
|-----------|-------------|-----|
| Accordion | Collapsible sections via `<details>`/`<summary>` | No |
| Alert | Status messages with `role="alert"` and variants | No |
| Badge | Inline labels/pills with class-based variants | No |
| Button | Styled native `<button>` with variants, sizes, groups | No |
| Card | Container using `.card` class | No |
| Dialog | Modal via native `<dialog>` with commandfor | Polyfill |
| Dropdown | Menu via `<ot-dropdown>` and popover API | Yes |
| Form | Inputs, selects, textareas, checkboxes, radios, fieldsets | No |
| Grid | 12-column responsive CSS grid | No |
| Meter | Measurement display via `<meter>` | No |
| Progress | Progress bar via `<progress>` | No |
| Sidebar | Dashboard layout with sticky sidebar and mobile overlay | Yes |
| Skeleton | Shimmer loading placeholders | No |
| Spinner | Animated loading indicator | No |
| Switch | Toggle via checkbox `role="switch"` | No |
| Table | Styled `<table>` with thead/tbody | No |
| Tabs | Tabbed interface via `<ot-tabs>` WebComponent | Yes |
| Toast | Dynamic notifications via `ot.toast()` | Yes |
| Tooltip | Enhanced `title` tooltips | Yes |
| Typography | Headings, paragraphs, lists, code, blockquote, hr | No |
| Utilities | Helper classes for layout, spacing, text | No |
