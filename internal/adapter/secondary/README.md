# Secondary Adapters

Concrete implementations of secondary ports.

These are the infrastructure components the application uses.

## Examples

- Database repositories (PostgreSQL, MySQL, etc.)
- External API clients
- Cache implementations (Redis)
- File system access
- Email / notification senders

They implement the interfaces defined in `core/ports/secondary`. Every secondary-adapter package must include an explicit compile-time assertion, for example:

```go
var _ secondaryport.UserRepository = (*PostgresUserRepository)(nil)
```

The architecture guardrail requires the assertion, and Go compilation verifies that the implementation satisfies the port.
