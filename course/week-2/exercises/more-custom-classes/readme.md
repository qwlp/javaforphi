# More Custom Classes

## Goal

Use tests and UML-style specifications to implement stronger domain classes.

## Files to work on

Starter source folders: `src/lib/`, `src/main/`. Follow the tasks below for the classes to edit or create. The copied Word handout contains the full lab specification.

## Tasks

1. Study `Die` and its constructor delegation and validation.
2. Create `lib.OrderLine` with `id`, `unitPrice`, and `quantity`; add constructors, accessors, `getCost()`, `toString()`, and value-based `equals()`.
3. Create `lib.Module` with `code`, `name`, `examWeight`, and `cwkWeight`. Its two weightings must always total 100; changing one adjusts the other.
4. Add small main-method demonstrations after the library tests pass.

## Expected behavior

OrderLine reports quantity times unit price and compares by value. Module weightings remain complementary and always total 100.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). The behavioral tests must pass. Use failures to identify the method or boundary case to revisit.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint: open if you get stuck</summary>

Express the weighting invariant in one place: the other weight is `100 - newWeight`. Test equality using two distinct objects with equal values.

</details>
