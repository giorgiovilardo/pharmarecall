# Tooltip

Enhanced tooltips that replace the browser's native `title` attribute.

**Requires JS:** Yes - converts `title` attributes to `data-tooltip` on page load.

## How It Works

On `DOMContentLoaded`, JS finds all elements with `title` attribute and:
1. Copies value to `data-tooltip`
2. Sets `aria-label` (if not already present)
3. Removes the native `title` (prevents browser's built-in tooltip)

CSS renders the tooltip via `::before` (arrow) and `::after` (text) pseudo-elements.

## Usage

Just use the standard `title` attribute:

```html
<button title="Save your document">Save</button>
```

Or apply `data-tooltip` directly (skips JS conversion):

```html
<button data-tooltip="Save your document" aria-label="Save your document">Save</button>
```

## Styling

- Appears above element on hover/focus-visible
- Delay: 700ms before showing
- Background: `var(--foreground)`, text: `var(--background)` (inverted)
- Single line (`white-space: nowrap`)

## Examples

### On buttons
```html
<button title="Save changes">Save</button>
<button title="Delete item" data-variant="danger">Delete</button>
```

### On links
```html
<a href="/profile" title="View your profile">Profile</a>
```

### On any element
```html
<span title="This is additional context">Hover me</span>
```
