# Aggregation

## Goal

Aggregation models a “has-many” relationship. A domain class owns a collection and exposes only the operations its clients need.

## Files to work on

Starter source folders: `src/lib/employeeRegister/`, `src/lib/playlist/`, `src/main/`. Follow the tasks below for the classes to edit or create. The copied Word handout contains the full lab specification.

## Tasks

1. Study `PlayList` as an aggregate of `Song` objects and trace its delegated add/remove/get/size operations.
2. Study `EmployeeRegister` as an aggregate of `Employee` objects.
3. Run `PlayListApp` and `EmployeeRegisterDemo`; follow the total and average calculations.
4. Compare the aggregate's focused public API with exposing an `ArrayList` directly.

## Expected behavior

Collection operations change the aggregate as expected. The demos report totals and averages derived from the current songs and employees.

## Automated check requirements

Checks exercise collection mutation, totals, averages, search, and boundary operations. Complete the playlist boundary behavior promised in its Javadoc: invalid `removeSong`, `moveUp`, and `moveDown` indices must do nothing. In the starter, movement guards have stray semicolons and removal lacks its guard; fix those defects. An empty employee register reports zero total and average salary.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). The embedded behavioral and boundary tests must pass. Compilation is only the first stage; a compiling but incorrect implementation fails.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint: open if you get stuck</summary>

Trace which collection operation each public method delegates to. Consider what happens when the collection is empty.

</details>
