# Form

Styled native form elements. CSS-only, no JS required (except validation state toggling in your own code).

## Field Container

Wrap inputs in `[data-field]` containers for consistent spacing.

| Attribute | Purpose |
|-----------|---------|
| `data-field` | Field wrapper (adds bottom margin) |
| `data-field="error"` | Shows `.error` message and styles input as invalid |
| `data-hint` | Hint/helper text below input |

## Auto-styled Elements

These are styled automatically with no classes needed:

- `<input>` (text, email, password, number, url, tel, search, date, datetime-local, file)
- `<textarea>`
- `<select>` (custom dropdown arrow)
- `<input type="checkbox">`
- `<input type="radio">`
- `<input type="range">`
- `<fieldset>` and `<legend>`
- `<label>`

## Validation

| Attribute | Element | Purpose |
|-----------|---------|---------|
| `aria-invalid="true"` | input | Marks input as invalid (red border + focus ring) |
| `data-field="error"` | container | Reveals `.error` message inside |
| `aria-describedby` | input | Links input to hint or error message by ID |

The browser's `:user-invalid` pseudo-class also triggers error styling automatically.

## Input Group

Use `<fieldset class="group">` to join inputs, selects, and buttons side by side. Use `<legend>` inside for an inline prefix label.

## Examples

### Standard form
```html
<form>
  <label data-field>
    Name
    <input type="text" placeholder="Enter your name" />
  </label>

  <label data-field>
    Email
    <input type="email" placeholder="you@example.com" />
  </label>

  <label data-field>
    Password
    <input type="password" placeholder="Password" aria-describedby="pw-hint" />
    <small id="pw-hint" data-hint>Must be at least 8 characters</small>
  </label>

  <div data-field>
    <label>Country</label>
    <select>
      <option value="">Select a country</option>
      <option value="us">United States</option>
      <option value="uk">United Kingdom</option>
    </select>
  </div>

  <label data-field>
    Message
    <textarea placeholder="Your message..."></textarea>
  </label>

  <label data-field>
    <input type="checkbox" /> I agree to the terms
  </label>

  <button type="submit">Submit</button>
</form>
```

### Radio group
```html
<fieldset class="hstack">
  <legend>Preference</legend>
  <label><input type="radio" name="pref"> Option A</label>
  <label><input type="radio" name="pref"> Option B</label>
  <label><input type="radio" name="pref"> Option C</label>
</fieldset>
```

### Input group
```html
<fieldset class="group">
  <legend>https://</legend>
  <input type="url" placeholder="subdomain">
  <select aria-label="Select a domain">
    <option>.example.com</option>
    <option>.example.net</option>
  </select>
  <button>Go</button>
</fieldset>

<fieldset class="group">
  <input type="text" placeholder="Search" />
  <button>Go</button>
</fieldset>
```

### Validation error
```html
<div data-field="error">
  <label for="email-input">Email</label>
  <input type="email" aria-invalid="true" aria-describedby="email-err" id="email-input" value="invalid" />
  <div id="email-err" class="error" role="status">Please enter a valid email address.</div>
</div>
```

### Range slider
```html
<label data-field>
  Volume
  <input type="range" min="0" max="100" value="50" />
</label>
```
