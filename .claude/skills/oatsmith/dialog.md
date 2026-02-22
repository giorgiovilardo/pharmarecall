# Dialog

Modal dialog using native `<dialog>` element with `commandfor`/`command` attributes for zero-JS opening/closing.

**Requires JS:** Polyfill only (for Safari `commandfor` support).

## Structure

```
<dialog>
  <form method="dialog">     ← auto-closes on submit
    <header>                  ← title + description
    <div>                     ← scrollable content
    <footer>                  ← action buttons
  </form>
</dialog>
```

The `<form method="dialog">` wrapper is required for proper close behavior. Omit it only for non-interactive content.

## Attributes

| Attribute | Element | Purpose |
|-----------|---------|---------|
| `commandfor="dialog-id"` | button | Opens or closes the target dialog |
| `command="show-modal"` | button | Opens as modal (with backdrop) |
| `command="close"` | button | Closes the dialog |
| `closedby="any"` | dialog | Allows close by clicking backdrop or pressing Escape |
| `method="dialog"` | form | Submit closes dialog, sets `returnValue` to button's `value` |

## Sections

| Element | Location | Styling |
|---------|----------|---------|
| `<header>` | Direct child of dialog or form | Padding, no bottom padding. `<h*>` and `<p>` styled. |
| `<div>`, `<section>`, `<p>` | Direct child | Padding, scrollable overflow |
| `<footer>` | Direct child | Flex end-aligned, gap between buttons, no top padding |

## Sizing

Default max-width is `32rem`, max-height `85vh`. Override with inline style:

```html
<dialog style="max-width: 48rem;">
```

## Examples

### Basic dialog
```html
<button commandfor="my-dialog" command="show-modal">Open dialog</button>
<dialog id="my-dialog" closedby="any">
  <form method="dialog">
    <header>
      <h3>Title</h3>
      <p>This is a dialog description.</p>
    </header>
    <div>
      <p>Dialog content goes here.</p>
    </div>
    <footer>
      <button type="button" commandfor="my-dialog" command="close" class="outline">Cancel</button>
      <button value="confirm">Confirm</button>
    </footer>
  </form>
</dialog>
```

### Form dialog
```html
<button commandfor="edit-dialog" command="show-modal">Edit</button>
<dialog id="edit-dialog">
  <form method="dialog">
    <header>
      <h3>Edit Profile</h3>
    </header>
    <div>
      <label data-field>Name <input name="name" required></label>
      <label data-field>Email <input name="email" type="email"></label>
    </div>
    <footer>
      <button type="button" commandfor="edit-dialog" command="close" class="outline">Cancel</button>
      <button value="save">Save</button>
    </footer>
  </form>
</dialog>
```

### Handling return value
```js
document.querySelector("#my-dialog").addEventListener("close", (e) => {
  console.log(e.target.returnValue); // "confirm" or "save"
});
```
