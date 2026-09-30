# The ArrayList Collection

## Goal

Use generic collections, interface-typed variables, enhanced loops, lambdas, and stream pipelines.

## Files to work on

Starter source folders: `src/lib/`, `src/main/`. Follow the tasks below for the classes to edit or create. The copied Word handout contains the full lab specification.

## Tasks

1. Extend `IntegerArrayListDemo` with `remove`, `set`, `clear`, and `isEmpty`.
2. Create `StringListDemo`; mutate a list and print uppercase/lowercase forms using both loop styles.
3. Create `NameListDemo`; collect four names, filter them, and explore `contains` with `Name.equals`.
4. Study how `MultipleDice` delegates collection operations and implements `Rollable`.
5. Create `OrderListDemo`, calculating total and average costs with loops and then a stream.

## Expected behavior

List operations update contents as expected; name membership uses value equality; loop and stream calculations give the same order totals and averages.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). All sources must compile. Run the relevant demo classes in your IDE and compare their output with the tasks; compilation alone does not verify behavior.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint — open if you get stuck</summary>

Use a typed list such as `ArrayList<String>`. Guard an average calculation when the list is empty.

</details>
