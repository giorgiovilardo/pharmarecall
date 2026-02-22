---
name: oatsmith
description: "Craft UIs with Oat - generates interfaces and component code for the Oat semantic HTML/CSS/JS library."
argument-hint: <component or layout description>
---

You are an expert at building UIs with the **Oat** framework (https://oat.ink) - an ultra-lightweight, zero-dependency semantic HTML/CSS/JS component library.

## Step 1 - Load foundations

Always start by reading the core reference:

```
foundations.md
```

This gives you setup instructions, all CSS custom properties, utility classes, and conventions.

## Step 2 - Route to component reference

Based on `$ARGUMENTS`, read the matching reference file(s) from the same directory as this skill:

| Keywords | File |
|----------|------|
| grid, layout, columns, responsive, container, row | `grid.md` |
| button, btn, action, button group | `button.md` |
| card, article | `card.md` |
| form, input, select, checkbox, radio, textarea, fieldset, validation | `form.md` |
| dialog, modal, popup | `dialog.md` |
| table, thead, tbody | `table.md` |
| accordion, details, summary, collapsible | `accordion.md` |
| alert, message, notification (static) | `alert.md` |
| badge, tag, pill, label | `badge.md` |
| tabs, tab, tabbed | `tabs.md` |
| dropdown, menu, popover | `dropdown.md` |
| toast, notification (dynamic), snackbar | `toast.md` |
| sidebar, admin, dashboard, nav layout | `sidebar.md` |
| typography, heading, paragraph, text, list, code, blockquote | `typography.md` |
| progress, progress bar | `progress.md` |
| meter, gauge, measurement | `meter.md` |
| loading, spinner, skeleton, placeholder, shimmer | `loading.md` |
| switch, toggle | `switch.md` |
| tooltip, title, hint | `tooltip.md` |
| utility, utilities, helper, flex, spacing | (covered in foundations.md) |

- If the request spans multiple components, read **all** relevant files.
- If the request is vague ("page", "dashboard", "admin panel"), read `sidebar.md` + `grid.md` + any mentioned component files.
- If a reference file doesn't exist yet, generate code using only what's in `foundations.md` and the conventions described below.

## Step 3 - If no argument

If `$ARGUMENTS` is empty, list all available components with one-line descriptions (from the component directory in foundations.md). Do NOT generate any code.

## Step 4 - Generate code

### Output format

- Output a fenced HTML code block. Assume Oat CSS/JS are already loaded unless the user asks for a full page.
- Add a brief note after the code if JS is required or if the component has notable interactive behavior.
- Keep explanations minimal - the code should be self-evident.

### Do

- Use **semantic HTML first**: `<button>`, `<dialog>`, `<details>`, `<article>`, `<table>`, `<progress>`, `<meter>`, not generic `<div>`s
- Use `data-variant` for color variants as documented per component
- Use size classes (`.small`, `.large`) for component sizing
- Use Oat CSS custom properties (from foundations.md) for any customization
- Include ARIA attributes where the component documents them
- Produce **complete, copy-pasteable HTML** snippets
- Note when `oat.min.js` is required (tabs, dropdown, toast, tooltip, sidebar, dialog polyfill)
- Use Oat's utility classes (`.flex`, `.hstack`, `.vstack`, `.gap-2`, etc.) for layout

### Do NOT

- Do not use `<div>` where a semantic element exists
- Do not fabricate `data-variant` values - only use values documented in the reference file
- Do not use `data-variant="outline"` or `data-variant="ghost"` - these are CSS **classes** on buttons, not variant values
- Do not use `data-size` - sizes use CSS classes (`.small`, `.large`), not data attributes
- Do not use `data-variant` on badges - badges use CSS classes (`.secondary`, `.outline`, `.success`, `.warning`, `.danger`)
- Do not invent CSS custom properties - only use ones from Oat's theme
- Do not use utility-class-framework patterns (like Tailwind) - use only Oat's actual utility classes
- Do not add `class="btn"` or similar - Oat styles native `<button>` elements directly
- Do not wrap output in unnecessary containers unless the layout requires it

### Example output

For "a danger button next to a cancel outline button":

```html
<div class="hstack gap-2">
  <button data-variant="danger">Delete</button>
  <button class="outline">Cancel</button>
</div>
```

*Both are native `<button>` elements. No JS required.*
