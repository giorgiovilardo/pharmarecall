# Sidebar

Dashboard-style layout with a sticky sidebar and optional top navigation. Mobile responsive with overlay.

**Requires JS:** Yes - for toggle behavior.

## Data Attributes

| Attribute | Element | Purpose |
|-----------|---------|---------|
| `data-sidebar-layout` | container | Enables sidebar grid layout |
| `data-sidebar-layout="always"` | container | Sidebar is collapsible at all screen sizes |
| `data-sidebar` | `<aside>` | Marks the sidebar element |
| `data-sidebar-toggle` | button | Toggles sidebar open/closed |
| `data-sidebar-open` | container | Applied automatically when sidebar is open |
| `data-topnav` | `<nav>` or `<header>` | Sticky top navigation bar |

## Structure

```html
<div data-sidebar-layout>
  <aside data-sidebar>
    <header>Logo</header>
    <nav>
      <ul>
        <li><a href="#" aria-current="page">Home</a></li>
        <li><a href="#">Settings</a></li>
      </ul>
    </nav>
    <footer>Bottom actions</footer>
  </aside>
  <main>Page content</main>
</div>
```

## Responsive Behavior

- **Desktop (> 768px):** Side-by-side grid layout, sidebar is sticky
- **Mobile (<= 768px):** Sidebar becomes fixed overlay, slides in from left
- **`always` mode:** Toggle button visible at all sizes, sidebar collapses to 0 width

## Navigation

- Use `aria-current="page"` on the active link for highlighted state
- Nested navigation uses `<details>` inside `<li>` for collapsible sections

## Examples

### Basic sidebar layout
```html
<div data-sidebar-layout>
  <aside data-sidebar>
    <nav>
      <ul>
        <li><a href="#" aria-current="page">Dashboard</a></li>
        <li><a href="#">Users</a></li>
        <li><a href="#">Settings</a></li>
      </ul>
    </nav>
  </aside>
  <main>
    <div style="padding: var(--space-4);">
      <h1>Dashboard</h1>
      <p>Main content area.</p>
    </div>
  </main>
</div>
```

### With top nav and collapsible sidebar
```html
<div data-sidebar-layout="always">
  <nav data-topnav>
    <button data-sidebar-toggle aria-label="Toggle menu" class="ghost icon">&#9776;</button>
    <strong>App Name</strong>
  </nav>
  <aside data-sidebar>
    <header><strong>Menu</strong></header>
    <nav>
      <ul>
        <li><a href="#" aria-current="page">Home</a></li>
        <li>
          <details open>
            <summary>Settings</summary>
            <ul>
              <li><a href="#">General</a></li>
              <li><a href="#">Security</a></li>
              <li><a href="#">Billing</a></li>
            </ul>
          </details>
        </li>
      </ul>
    </nav>
    <footer>
      <button class="outline" style="width: 100%;">Logout</button>
    </footer>
  </aside>
  <main>
    <div style="padding: var(--space-4);">Main content.</div>
  </main>
</div>
```
