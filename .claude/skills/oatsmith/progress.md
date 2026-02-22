# Progress

Progress bar using native `<progress>` element. CSS-only, no JS required.

## Attributes

| Attribute | Type | Purpose |
|-----------|------|---------|
| `value` | number | Current progress value |
| `max` | number | Maximum value (default 1) |

Omitting `value` renders an indeterminate (animated) state.

## Styling

- Full width, height: `var(--bar-height)`
- Track: `var(--muted)`, fill: `var(--primary)`
- Border radius: full (pill-shaped)

## Examples

### Basic progress bars
```html
<progress value="30" max="100"></progress>
<progress value="60" max="100"></progress>
<progress value="90" max="100"></progress>
```

### Indeterminate
```html
<progress></progress>
```
