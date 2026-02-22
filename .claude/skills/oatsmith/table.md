# Table

Styled native `<table>` element. CSS-only, no JS required.

## Structure

Use standard `<table>` with `<thead>` and `<tbody>`. No classes needed.

## Styling

- Full width, fixed layout, collapsed borders
- `<thead>`: bottom border, medium font weight, muted-foreground color
- `<tbody tr>`: hover background highlight
- Cell padding: `var(--space-3)` vertical

## Examples

### Basic table
```html
<table>
  <thead>
    <tr>
      <th>Name</th>
      <th>Email</th>
      <th>Role</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>Alice Johnson</td>
      <td>alice@example.com</td>
      <td>Admin</td>
    </tr>
    <tr>
      <td>Bob Smith</td>
      <td>bob@example.com</td>
      <td>Editor</td>
    </tr>
    <tr>
      <td>Carol White</td>
      <td>carol@example.com</td>
      <td>Viewer</td>
    </tr>
  </tbody>
</table>
```

### Table with badges
```html
<table>
  <thead>
    <tr>
      <th>Name</th>
      <th>Status</th>
      <th>Role</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>Alice Johnson</td>
      <td><span class="badge success">Active</span></td>
      <td>Admin</td>
    </tr>
    <tr>
      <td>Bob Smith</td>
      <td><span class="badge warning">Pending</span></td>
      <td>Editor</td>
    </tr>
  </tbody>
</table>
```
