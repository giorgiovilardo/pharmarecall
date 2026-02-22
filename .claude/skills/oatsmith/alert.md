# Alert

Static notification messages using `role="alert"`. CSS-only, no JS required.

## Attributes

| Attribute | Values | Purpose |
|-----------|--------|---------|
| `role="alert"` | - | Required. Styles the element and announces to screen readers. |
| `data-variant` | `danger` (or `error`), `success`, `warning` | Color variant. Default is neutral gray. |

`data-variant="error"` and `data-variant="danger"` are interchangeable.

## Styling

- Padding, border-radius, colored left border for variants
- Variants apply background tint via `color-mix()`
- Use `<strong>` for alert titles

## Examples

### All variants
```html
<div role="alert">
  <strong>Note.</strong> This is a default alert message.
</div>

<div role="alert" data-variant="success">
  <strong>Success!</strong> Your changes have been saved.
</div>

<div role="alert" data-variant="warning">
  <strong>Warning!</strong> Please review before continuing.
</div>

<div role="alert" data-variant="danger">
  <strong>Error!</strong> Something went wrong.
</div>
```
