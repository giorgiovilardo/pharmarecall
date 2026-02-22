# Card

Container component using the `.card` class. CSS-only, no JS required.

## Classes

| Class | Purpose |
|-------|---------|
| `.card` | Applies card styling (background, border, shadow, padding, radius) |

Typically used on `<article>` elements. Internal structure uses semantic elements: `<header>`, `<footer>`, `<p>`.

## Styling

- Background: `var(--card)`, color: `var(--card-foreground)`
- Border: 1px solid `var(--border)`
- Border radius: `var(--radius-medium)`
- Shadow: `var(--shadow-small)`
- Padding: `var(--space-6)`

## Examples

### Basic card
```html
<article class="card">
  <header>
    <h3>Card Title</h3>
    <p>Card description goes here.</p>
  </header>
  <p>This is the card content. It can contain any HTML.</p>
  <footer class="flex gap-2 mt-4">
    <button class="outline">Cancel</button>
    <button>Save</button>
  </footer>
</article>
```

### Card grid
```html
<div class="container">
  <div class="row">
    <div class="col-4">
      <article class="card">
        <h4>Feature One</h4>
        <p>Description of the first feature.</p>
      </article>
    </div>
    <div class="col-4">
      <article class="card">
        <h4>Feature Two</h4>
        <p>Description of the second feature.</p>
      </article>
    </div>
    <div class="col-4">
      <article class="card">
        <h4>Feature Three</h4>
        <p>Description of the third feature.</p>
      </article>
    </div>
  </div>
</div>
```

### Simple content card
```html
<article class="card">
  <p>A simple card with just text content.</p>
</article>
```
