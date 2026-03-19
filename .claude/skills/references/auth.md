# Authentication & Base URL

Shared setup used by all Lab Nocturne image skills. Follow these steps before making any API call.

## Resolve the API key

Check if the user has an API key configured:

```bash
printenv LABNOCTURNE_API_KEY
```

- If the command outputs a key value, use it. Do NOT echo or print the key back to the user — treat it as sensitive.
- If the output is empty (variable is unset), generate a test key automatically:
  ```bash
  curl -s https://images.labnocturne.com/key
  ```
  The response is JSON: `{"api_key": "ln_test_..."}`. Extract the `api_key` value and use it. Tell the user a temporary test key was generated (7-day file retention, 10MB limit).

## Resolve the base URL

```bash
printenv LABNOCTURNE_BASE_URL
```

Use the output if set, otherwise default to `https://images.labnocturne.com`.

## Error responses

All API errors return this shape:
```json
{
  "error": {
    "message": "Human-readable message",
    "type": "error_category",
    "code": "machine_readable_code"
  }
}
```

Common error codes across all endpoints:

| `code` | Suggested fix |
|---|---|
| `missing_api_key` | Set `$LABNOCTURNE_API_KEY` or let the skill generate a test key |
| `invalid_api_key` | Check that `$LABNOCTURNE_API_KEY` is correct, or unset it to auto-generate a test key |
| `invalid_auth_format` | The key should be passed as `Bearer <key>` — this is handled automatically by the skill |
