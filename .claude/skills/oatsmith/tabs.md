# Tabs

Tabbed interface using `<ot-tabs>` WebComponent with ARIA roles.

**Requires JS:** Yes - `<ot-tabs>` is a custom element.

## Structure

```html
<ot-tabs>
  <div role="tablist">
    <button role="tab">Tab 1</button>
    <button role="tab">Tab 2</button>
  </div>
  <div role="tabpanel">Panel 1 content</div>
  <div role="tabpanel">Panel 2 content</div>
</ot-tabs>
```

Tabs and panels are matched by order (first tab maps to first panel, etc.).

## ARIA (auto-managed)

The WebComponent automatically sets:
- `aria-selected` on the active tab
- `aria-controls` / `aria-labelledby` linking tabs to panels
- `tabindex` for keyboard navigation
- Unique IDs if not provided

## Events

| Event | Detail | Fires when |
|-------|--------|------------|
| `ot-tab-change` | `{ index, tab }` | Active tab changes |

## Properties

| Property | Type | Purpose |
|----------|------|---------|
| `activeIndex` | number | Get or set the active tab index |

## Keyboard Navigation

- `ArrowLeft` / `ArrowRight` - cycle through tabs (wraps around)

## Examples

### Basic tabs
```html
<ot-tabs>
  <div role="tablist">
    <button role="tab">Account</button>
    <button role="tab">Password</button>
    <button role="tab">Notifications</button>
  </div>
  <div role="tabpanel">
    <h3>Account Settings</h3>
    <p>Manage your account information here.</p>
  </div>
  <div role="tabpanel">
    <h3>Password Settings</h3>
    <p>Change your password here.</p>
  </div>
  <div role="tabpanel">
    <h3>Notification Settings</h3>
    <p>Configure your notification preferences.</p>
  </div>
</ot-tabs>
```

### Listening to tab change
```js
document.querySelector("ot-tabs").addEventListener("ot-tab-change", (e) => {
  console.log("Switched to tab", e.detail.index);
});
```
