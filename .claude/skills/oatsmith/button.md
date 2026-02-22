# Button

Styled native `<button>` element with semantic variants, appearance styles, sizes, and button groups. CSS-only, no JS required.

## Variants

Color variants use `data-variant`. Appearance styles and sizes use CSS classes.

| Type | Attribute/Class | Values |
|------|----------------|--------|
| Color | `data-variant` | `secondary`, `danger` (default is primary) |
| Appearance | class | `.outline`, `.ghost` |
| Size | class | `.small`, `.large` |
| Icon-only | class | `.icon` (square button, combine with `.small` or `.large`) |

**Important:** `.outline` and `.ghost` are CSS classes, NOT `data-variant` values. They can combine with `data-variant`:

```html
<button data-variant="danger" class="outline">Danger outline</button>
```

## Selectors

Oat styles these automatically: `<button>`, `[type=submit]`, `[type=reset]`, `[type=button]`, `a.button`, `::file-selector-button`

## Button Group

Wrap buttons in `<menu class="buttons">` for connected/joined buttons.

## Examples

### All variants
```html
<button>Primary</button>
<button data-variant="secondary">Secondary</button>
<button data-variant="danger">Danger</button>
<button class="outline">Outline</button>
<button data-variant="danger" class="outline">Danger outline</button>
<button data-variant="secondary" class="outline">Secondary outline</button>
<button class="ghost">Ghost</button>
<button data-variant="danger" class="ghost">Danger ghost</button>
<button disabled>Disabled</button>
```

### Sizes
```html
<button class="small">Small</button>
<button>Default</button>
<button class="large">Large</button>
```

### Icon button
```html
<button class="icon" aria-label="Settings">&#9881;</button>
<button class="icon small" aria-label="Close">&#10005;</button>
<button class="icon large" aria-label="Menu">&#9776;</button>
```

### Link as button
```html
<a href="/signup" class="button">Sign Up</a>
```

### Button group
```html
<menu class="buttons">
  <button class="outline">Left</button>
  <button class="outline">Center</button>
  <button class="outline">Right</button>
</menu>
```
