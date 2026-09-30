# Portfolio B: Register

## Goal

Implement `lib.Register` as a capacity-limited aggregation of `Name` objects, then complete `RegisterApp.execute`.

## Files to work on

Starter source folders: `src/lib/`, `src/main/`. Follow the tasks below for the classes to edit or create. The copied Word handout contains the full lab specification.

## Required behavior

- Store exactly two private instance fields: `ArrayList<Name>` and room capacity.
- Default capacity is 25; the custom constructor accepts a capacity.
- Add, bulk-add, retrieve, remove, size, clear, and empty operations. An addition that would breach capacity does nothing.
- Case-insensitive family-initial search and first-name occurrence count.
- Implement `Iterable<Name>` and `sortRegister()` using `Name`'s natural order.
- `RegisterApp.execute` performs the specified removal/addition and builds lowercase email addresses dynamically.

## Expected behavior

Register respects its capacity, searches names case-insensitively where specified, supports iteration and sorting, and generates email addresses from current data.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). The behavioral tests must pass. Use failures to identify the method or boundary case to revisit.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint — open if you get stuck</summary>

Check capacity before mutating the list, including bulk additions. Delegate iteration and sorting to the internal collection.

</details>
