# Composition

## Goal

Composition models a “has-a” relationship: an object owns a fixed set of other objects and delegates appropriate behavior to them.

## Files to work on

Starter source folders: `src/lib/dice/`, `src/lib/employee/`, `src/main/`. Follow the tasks below for the classes to edit or create. The copied Word handout contains the full lab specification.

## Tasks

1. Study `Employee`, including its `Name` and `Date` fields, constructors, equality, and nested method calls.
2. Study `PairOfDice`, which owns two `Die` objects and delegates `roll()` and score calculation.
3. Trace both examples through `EmployeeDemo` and `RollableDemo`.
4. Explain why using `Rollable` lets a `Die` and a `PairOfDice` be handled uniformly.

## Expected behavior

Employee delegates name and date behavior to its contained objects. PairOfDice rolls both dice and reports their combined score through Rollable.

## Automated check requirements

Checks verify that `PairOfDice` retains supplied dice, rolls both objects, and sums their scores. Employee checks cover supplied Name/Date references, setters, and value equality. Test doubles keep dice checks deterministic; you do not need to produce a particular random roll.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). The embedded behavioral and boundary tests must pass. Compilation is only the first stage; a compiling but incorrect implementation fails.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint: open if you get stuck</summary>

Follow a method call from the outer object to the contained object. Delegate instead of duplicating the contained object’s calculations.

</details>
