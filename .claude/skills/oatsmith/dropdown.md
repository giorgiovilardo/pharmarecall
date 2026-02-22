# Dropdown

Menu dropdown using `<ot-dropdown>` WebComponent and the Popover API.

**Requires JS:** Yes - `<ot-dropdown>` is a custom element.

## Structure

```html
<ot-dropdown>
  <button popovertarget="menu-id">Trigger</button>
  <menu popover id="menu-id">
    <button role="menuitem">Item</button>
  </menu>
</ot-dropdown>
```

## Attributes

| Attribute | Element | Purpose |
|-----------|---------|---------|
| `popovertarget="id"` | trigger button | Links to the popover menu |
| `popover` | menu element | Enables Popover API |
| `role="menuitem"` | menu buttons | Enables keyboard nav and hover styling |

## ARIA (auto-managed)

- `aria-expanded` set on trigger based on popover state

## Keyboard Navigation

- `ArrowDown` / `ArrowUp` - cycle through menu items
- `Escape` - close menu

## Styling

- Menu: min-width 12rem, card background, border, shadow
- Items: flex, padding, hover/focus to `var(--accent)`
- `<hr>` inside menu creates a visual separator

## Examples

### Menu dropdown
```html
<ot-dropdown>
  <button popovertarget="actions" class="outline">Options</button>
  <menu popover id="actions">
    <button role="menuitem">Profile</button>
    <button role="menuitem">Settings</button>
    <button role="menuitem">Help</button>
    <hr>
    <button role="menuitem">Logout</button>
  </menu>
</ot-dropdown>
```

### Confirmation popover
```html
<ot-dropdown>
  <button popovertarget="confirm-pop" data-variant="danger">Delete</button>
  <article class="card" popover id="confirm-pop">
    <header>
      <h4>Are you sure?</h4>
      <p>This action cannot be undone.</p>
    </header>
    <footer>
      <button class="outline small" popovertarget="confirm-pop">Cancel</button>
      <button data-variant="danger" class="small" popovertarget="confirm-pop">Delete</button>
    </footer>
  </article>
</ot-dropdown>
```
