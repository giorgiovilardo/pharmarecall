# Meter

Measurement display using native `<meter>` element. CSS-only, no JS required.

## Attributes

| Attribute | Type | Purpose |
|-----------|------|---------|
| `value` | number | Current value |
| `min` | number | Minimum value |
| `max` | number | Maximum value |
| `low` | number | Low threshold |
| `high` | number | High threshold |
| `optimum` | number | Optimal value (determines color interpretation) |

## Color Behavior

The browser determines color based on where `value` falls relative to `low`, `high`, and `optimum`:

- **Optimum range:** `var(--success)` (green)
- **Suboptimal:** `var(--warning)` (orange)
- **Poor:** `var(--danger)` (red)

## Styling

- Full width, height: `var(--bar-height)`
- Border radius: full (pill-shaped)

## Examples

### Good / warning / poor
```html
<meter value="0.8" min="0" max="1" low="0.3" high="0.7" optimum="1"></meter>
<meter value="0.5" min="0" max="1" low="0.3" high="0.7" optimum="1"></meter>
<meter value="0.2" min="0" max="1" low="0.3" high="0.7" optimum="1"></meter>
```
