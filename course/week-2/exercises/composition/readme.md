# Composition

Composition models a “has-a” relationship: an object owns a fixed set of other objects and delegates appropriate behavior to them.

## Tasks

1. Study `Employee`, including its `Name` and `Date` fields, constructors, equality, and nested method calls.
2. Study `PairOfDice`, which owns two `Die` objects and delegates `roll()` and score calculation.
3. Trace both examples through `EmployeeDemo` and `RollableDemo`.
4. Explain why using `Rollable` lets a `Die` and a `PairOfDice` be handled uniformly.

Run `phi check` to compile the complete project.
