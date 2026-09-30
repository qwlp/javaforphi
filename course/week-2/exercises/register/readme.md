# Portfolio B: Register

Implement `lib.Register` as a capacity-limited aggregation of `Name` objects, then complete `RegisterApp.execute`.

## Required behavior

- Store exactly two private instance fields: `ArrayList<Name>` and room capacity.
- Default capacity is 25; the custom constructor accepts a capacity.
- Add, bulk-add, retrieve, remove, size, clear, and empty operations. An addition that would breach capacity does nothing.
- Case-insensitive family-initial search and first-name occurrence count.
- Implement `Iterable<Name>` and `sortRegister()` using `Name`'s natural order.
- `RegisterApp.execute` performs the specified removal/addition and builds lowercase email addresses dynamically.

Run `phi check`. All Register and application tests run together, including the later iterable/sorting portfolio steps.
