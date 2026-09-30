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

## Automated check requirements

For automated checks, `StringListDemo.main` uses the supplied command-line words when arguments are present and prints uppercase and lowercase forms of each; keep your collection mutation demonstration for runs without arguments. `NameListDemo` reads four first/family name pairs from standard input and prints each full name. Factor the order calculations into `public static double totalCost(List<OrderLine> orders)` and `public static double averageCost(List<OrderLine> orders)` on `main.OrderListDemo`; an empty list returns zero for both. Use these methods in your demo so checks can try different order contents.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). The embedded behavioral and boundary tests must pass. Compilation is only the first stage; a compiling but incorrect implementation fails.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint: open if you get stuck</summary>

Use a typed list such as `ArrayList<String>`. Guard an average calculation when the list is empty.

</details>
