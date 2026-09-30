# More Custom Classes

Use tests and UML-style specifications to implement stronger domain classes.

## Tasks

1. Study `Die` and its constructor delegation and validation.
2. Create `lib.OrderLine` with `id`, `unitPrice`, and `quantity`; add constructors, accessors, `getCost()`, `toString()`, and value-based `equals()`.
3. Create `lib.Module` with `code`, `name`, `examWeight`, and `cwkWeight`. Its two weightings must always total 100; changing one adjusts the other.
4. Add small main-method demonstrations after the library tests pass.

Run `phi check` to execute the embedded `OrderLineTest` and `ModuleTest` suites.
