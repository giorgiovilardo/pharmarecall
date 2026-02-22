# Loading

Spinner and skeleton placeholder components. CSS-only, no JS required.

## Spinner

Animated rotating circle.

| Class | Purpose |
|-------|---------|
| `.spinner` | Required. Base spinner. |
| `.small` | Smaller (1rem) |
| `.large` | Larger (2rem) |

Always include `role="status"` for accessibility.

### Examples
```html
<div role="status" class="spinner small"></div>
<div role="status" class="spinner"></div>
<div role="status" class="spinner large"></div>
```

---

## Skeleton

Shimmer loading placeholders.

| Class | Purpose |
|-------|---------|
| `.skeleton` | Required. Base shimmer animation. |
| `.line` | Text-line placeholder (full width, 1rem height) |
| `.box` | Square placeholder (4rem x 4rem) |

Always include `role="status"` for accessibility.

### Examples
```html
<div role="status" class="skeleton line"></div>
<div role="status" class="skeleton line" style="width: 60%"></div>
<div role="status" class="skeleton box"></div>
```

### Skeleton card pattern
```html
<article class="card" style="display: flex; gap: var(--space-3);">
  <div role="status" class="skeleton box"></div>
  <div style="flex: 1; display: flex; flex-direction: column; gap: var(--space-1);">
    <div role="status" class="skeleton line"></div>
    <div role="status" class="skeleton line" style="width: 60%"></div>
  </div>
</article>
```
