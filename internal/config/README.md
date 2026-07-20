# Config Layer

Dependency-free configuration structures, defaults, and standard-library validation.

## Best Practices

- Define configuration as structs with validation
- Keep configuration simple and immutable where possible
- Use only the Go standard library in this package
- Put vendor-specific or format-specific loading behind an adapter boundary
- Construct configuration in `cmd/` and pass it to primary adapters
