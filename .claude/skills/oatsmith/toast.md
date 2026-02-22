# Toast

Dynamic notification system. Global JS API - no HTML element needed.

**Requires JS:** Yes - uses `window.ot.toast()`.

## API

### `ot.toast(message, title?, options?)`

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `message` | string | - | Main message text |
| `title` | string | `undefined` | Optional title |
| `options.variant` | string | `""` | `"success"`, `"danger"`, `"warning"` |
| `options.placement` | string | `"top-right"` | See placements below |
| `options.duration` | number | `4000` | Auto-close in ms. `0` = persistent. |

### `ot.toastEl(element, options?)`

Show a custom HTML element as a toast. Accepts an `HTMLElement`, `HTMLTemplateElement`, or CSS selector string. Options: `placement`, `duration`.

### `ot.toast.clear(placement?)`

Clear all toasts, or only those in a specific placement.

## Placements

`top-left`, `top-center`, `top-right`, `bottom-left`, `bottom-center`, `bottom-right`

## Variants

| Variant | Color |
|---------|-------|
| (default) | Neutral |
| `success` | Green left border |
| `danger` | Red left border |
| `warning` | Orange left border |

## Examples

### Basic toasts
```html
<button onclick="ot.toast('File uploaded successfully', 'Done', { variant: 'success' })">
  Success toast
</button>

<button onclick="ot.toast('Something went wrong', 'Error', { variant: 'danger' })">
  Error toast
</button>

<button onclick="ot.toast('Check your input', 'Warning', { variant: 'warning' })">
  Warning toast
</button>
```

### Custom placement and duration
```html
<button onclick="ot.toast('Saved', '', { placement: 'bottom-center', duration: 2000 })">
  Bottom center, 2s
</button>
```

### Custom HTML toast
```html
<template id="undo-toast">
  <output class="toast" data-variant="success">
    <h6 class="toast-title">Changes saved</h6>
    <p>Your document has been updated.</p>
    <button data-variant="secondary" class="small" onclick="this.closest('.toast').remove()">Okay</button>
  </output>
</template>

<button onclick="ot.toastEl('#undo-toast', { duration: 8000 })">Toast with action</button>
```

### Clear all toasts
```js
ot.toast.clear();            // all placements
ot.toast.clear("top-right"); // specific placement
```
