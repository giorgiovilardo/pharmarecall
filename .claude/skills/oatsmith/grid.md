# Grid

12-column responsive CSS grid system. CSS-only, no JS required.

## CSS Custom Properties

Override these on `.row` or `.container` to customize:

| Property | Default | Purpose |
|----------|---------|---------|
| `--grid-cols` | `12` | Number of columns |
| `--grid-gap` | `1.5rem` | Gap between columns |
| `--container-max` | `1280px` | Max container width |
| `--container-pad` | `1rem` | Container inline padding |

## Classes

| Class | Purpose |
|-------|---------|
| `.container` | Centered max-width wrapper with inline padding |
| `.row` | CSS grid with 12 columns |
| `.col` | Full-width column (spans all columns) |
| `.col-{1-12}` | Column spanning N of 12 columns |
| `.offset-{1-6}` | Push column start by N positions |
| `.col-end` | Align column to the end of the row |

## Responsive Behavior

At `max-width: 768px`:
- Grid switches to **4 columns** with `1rem` gap
- All `.col-*` spans become full width
- All `.offset-*` are ignored (reset to `auto`)

## Examples

### Equal columns
```html
<div class="container">
  <div class="row">
    <div class="col-4">One</div>
    <div class="col-4">Two</div>
    <div class="col-4">Three</div>
  </div>
</div>
```

### Two-column (halves)
```html
<div class="row">
  <div class="col-6">Left</div>
  <div class="col-6">Right</div>
</div>
```

### Sidebar layout
```html
<div class="container">
  <div class="row">
    <aside class="col-3">Sidebar</aside>
    <main class="col-9">Main content</main>
  </div>
</div>
```

### Offset
```html
<div class="row">
  <div class="col-6 offset-3">Centered 6-column block</div>
</div>
```

### End-aligned column
```html
<div class="row">
  <div class="col-3">Left</div>
  <div class="col-4 col-end">Pushed to end</div>
</div>
```

### Nested rows
```html
<div class="row">
  <div class="col-8">
    <div class="row">
      <div class="col-6">Nested left</div>
      <div class="col-6">Nested right</div>
    </div>
  </div>
  <div class="col-4">Sidebar</div>
</div>
```

### Card grid
```html
<div class="container">
  <div class="row">
    <div class="col-4"><article class="card">Card 1</article></div>
    <div class="col-4"><article class="card">Card 2</article></div>
    <div class="col-4"><article class="card">Card 3</article></div>
  </div>
</div>
```

### Custom gap and max-width
```html
<div class="container" style="--container-max: 960px;">
  <div class="row" style="--grid-gap: 2rem;">
    <div class="col-6">Left</div>
    <div class="col-6">Right</div>
  </div>
</div>
```
