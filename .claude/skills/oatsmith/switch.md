# Switch

Toggle switch using native checkbox with `role="switch"`. CSS-only, no JS required.

## Structure

```html
<label>
  <input type="checkbox" role="switch"> Label text
</label>
```

The `role="switch"` attribute on an `<input type="checkbox">` transforms it into a toggle switch.

## Styling

- Pill-shaped track, circular thumb
- Off: `var(--input)` background
- On: `var(--primary)` background
- Thumb slides left/right with smooth transition
- Disabled: 50% opacity

## Examples

### Basic switches
```html
<label>
  <input type="checkbox" role="switch"> Notifications
</label>
<label>
  <input type="checkbox" role="switch" checked> Dark mode
</label>
```

### Disabled
```html
<label>
  <input type="checkbox" role="switch" disabled> Disabled off
</label>
<label>
  <input type="checkbox" role="switch" checked disabled> Disabled on
</label>
```
