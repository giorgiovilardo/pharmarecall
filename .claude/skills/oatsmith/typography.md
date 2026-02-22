# Typography

Styled native text elements. CSS-only, no JS required. All elements are styled automatically.

## Headings

| Element | Size | Top Margin | Bottom Margin |
|---------|------|------------|---------------|
| `<h1>` | `--text-1` | `--space-10` | `--space-6` |
| `<h2>` | `--text-2` | `--space-8` | `--space-5` |
| `<h3>` | `--text-3` | `--space-6` | `--space-4` |
| `<h4>` | `--text-4` | `--space-5` | `--space-3` |
| `<h5>` | `--text-5` | `--space-4` | `--space-2` |
| `<h6>` | `--text-regular` | `--space-4` | `--space-2` |

All headings: `font-weight: var(--font-semibold)`, `line-height: 1.25`

## Text Elements

| Element | Styling |
|---------|---------|
| `<p>` | Bottom margin `--space-4` (0 if last child) |
| `<a>` | Primary color, underline, hover darkens |
| `<strong>`, `<b>` | Semibold weight |
| `<em>`, `<i>` | Italic |
| `<small>` | Font size `--text-7` |
| `<code>` | Monospace, faint background, small padding |
| `<pre>` | Monospace, faint background, padding, horizontal scroll |
| `<blockquote>` | Left border 4px, italic, muted color |
| `<mark>` | Warning color at 30% opacity background |
| `<hr>` | Top border, vertical margin `--space-8` |
| `<ul>`, `<ol>` | Left padding, bottom margin. Use `.unstyled` to remove bullets/padding. |

## Examples

### Mixed content
```html
<h2>Article Title</h2>
<p>A paragraph with <strong>bold</strong>, <em>italic</em>, <a href="#">a link</a>, and <code>inline code</code>.</p>

<blockquote>This is a blockquote styled automatically.</blockquote>

<pre><code>const greeting = "Hello, world!";</code></pre>

<ul>
  <li>List item one</li>
  <li>List item two</li>
</ul>
```
