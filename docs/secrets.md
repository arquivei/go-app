# Secrets in the configuration log

When the application boots, `app.Bootstrap` logs the whole configuration struct
using `logger.Flatten`:

```text
[app] Configuration loaded and global logger configured. config="App.Log.Level: 'info', Database.Password: '***', ..."
```

To keep credentials out of the logs, `logger.Flatten` masks values that look
like secrets. Values are masked automatically, so a new field such as `APIKey`
is safe even if nobody remembers to tag it. The `secret` struct tag overrides
the automatic behavior in either direction.

## Rules

For every value in the configuration, the first matching rule wins:

| # | Condition | Logged as |
|---|-----------|-----------|
| 1 | The field is tagged `secret:"false"` | The original value, untouched |
| 2 | The field is tagged `secret:"true"` (any type) | `'***'` when set, `''` when empty |
| 3 | A string field (`string`, `*string`, `[]string`) whose name ends with a sensitive suffix | `'***'` when set, `''` when empty |
| 4 | A string that contains `://` and is a valid URL | The URL with its password and sensitive query parameters replaced by `***` |
| 5 | None of the above | The original value |

Details:

- **Sensitive suffixes** are `Key`, `Token`, `Password` and `Secret` (and their
  plurals), `Pass`, `Pwd`, `DSN`, `Salt`, `Credential`, `Credentials`, `Auth`,
  `Authorization`, `Cert` and `PEM`. The match is case-insensitive and also
  applies to the whole name, so `Key`, `APIKey`, `ApiKey` and `DBPassword` all
  match.
- **Only string fields are matched by name.** Fields such as
  `PartitionKey bool` or `ShardKey int` are not secrets, so they are logged as
  usual. To hide a non-string value, tag it with
  `secret:"true"`.
- **Map keys** follow the same name rule as struct fields. In a
  `map[string]string`, the values of the keys `api_key` and `Authorization` are
  masked.
- **File paths stay visible.** `CertFile` or `PrivateKeyPath` do not end with a
  sensitive suffix, so the path is logged. Only the content is a secret.
- **Tags apply to nested fields.** A struct field tagged `secret:"true"` masks
  every field inside it, and `secret:"false"` reveals every field inside it.
  The nearest tag wins, so a field inside a hidden struct can still be revealed
  with its own `secret:"false"`.
- **URLs:** the password of the user info (`postgres://user:***@host/db`) and
  the value of every query parameter whose name matches the sensitive suffixes
  (`?token=***`) are redacted. Everything else in the URL is kept, which
  helps when debugging. Each element of a `[]string` is checked on its own.
- **Malformed URLs** (a string that contains `://` but cannot be parsed) are
  masked as `'***'`, since they may still carry credentials.
- A secret tag with any value other than `true` or `false` is ignored.

## Examples

```go
var cfg struct {
	app.Config

	Database struct {
		URL      string // "postgres://app:s3cr3t@db:5432/app" -> "postgres://app:***@db:5432/app"
		Password string // masked: ends with "Password"
	}

	Webhook struct {
		URL string // "https://example.com/hook?token=abc" -> "https://example.com/hook?token=***"
	}

	// Not a secret: logged as is.
	PartitionKey string `secret:"false"`

	// A secret whose name does not look like one.
	SigningMaterial string `secret:"true"`
}
```

## Limitations

The automatic rules are best-effort. They do **not** catch:

- Secrets in a URL path, such as `https://hooks.example.com/services/T000/B000/XXXX`.
- Secrets in fields with unusual names, such as `Signature` or `Seed`.
- Secrets in values that are not strings, such as `[]byte`.
- Secrets inside slices of structs or maps, such as `[]BrokerConfig`. A slice
  is logged as a single value, so the rules are not applied to its elements
  (except for `[]string`). Tag the slice field with `secret:"true"`.
- Secrets logged by your own code. Only `logger.Flatten` applies these rules.

Tag those fields with `secret:"true"`.
