# Badge

Inline label/pill element. CSS-only, no JS required.

## Classes

| Class | Purpose |
|-------|---------|
| `.badge` | Required. Base badge styling. |
| `.secondary` | Gray/muted variant |
| `.outline` | Bordered, transparent background |
| `.success` | Green variant |
| `.warning` | Orange variant |
| `.danger` | Red variant |

**Important:** Badges use CSS classes for variants, NOT `data-variant`. This is different from most other Oat components.

## Styling

- Pill-shaped (`border-radius: var(--radius-full)`)
- Inline-flex, small text (`--text-8`), medium font weight
- Default: primary color background

## Examples

### All variants
```html
<span class="badge">Default</span>
<span class="badge secondary">Secondary</span>
<span class="badge outline">Outline</span>
<span class="badge success">Success</span>
<span class="badge warning">Warning</span>
<span class="badge danger">Danger</span>
```

### In a table cell
```html
<td><span class="badge success">Active</span></td>
```

### In a heading
```html
<h3>Notifications <span class="badge">3</span></h3>
```
