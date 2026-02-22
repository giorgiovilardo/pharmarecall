# Accordion

Collapsible sections using native `<details>` and `<summary>` elements. CSS-only, no JS required.

## Structure

Stack multiple `<details>` elements. Adjacent details merge their borders automatically.

## Styling

- Border: 1px solid, radius on first/last, merged for adjacent
- Summary: flex, padding, pointer cursor, chevron icon (rotates on open)
- Summary hover: muted background
- Content padding: `var(--space-4)`

## Examples

### Stacked accordion
```html
<details>
  <summary>What is Oat?</summary>
  <p>Oat is a minimal, semantic-first UI component library with zero dependencies.</p>
</details>
<details>
  <summary>How do I install it?</summary>
  <p>Include the CSS and JS files via CDN or npm. Most elements are styled automatically.</p>
</details>
<details>
  <summary>Is it accessible?</summary>
  <p>Yes. It uses semantic HTML and ARIA attributes. Keyboard navigation works out of the box.</p>
</details>
```

### Single open by default
```html
<details open>
  <summary>Expanded by default</summary>
  <p>This section starts open.</p>
</details>
<details>
  <summary>Collapsed section</summary>
  <p>This section starts closed.</p>
</details>
```
