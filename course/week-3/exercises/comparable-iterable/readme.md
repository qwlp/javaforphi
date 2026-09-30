# Comparable and Iterable Example

## Goal

Study a completed employee register that participates in standard Java iteration and natural ordering.

## Files to work on

Starter source folders: `src/lib/employeeRegister/`, `src/main/`. Follow the tasks below for the classes to edit or create. The copied Word handout contains the full lab specification.

## Tasks

1. Find the class that implements `Iterable` and trace its `iterator()` delegation.
2. Find each `Comparable` implementation and write down its field comparison order.
3. Follow the register sorting call from the demo into the collection API.
4. Add two employees with shared fields to verify the comparison tie-breakers.

## Expected behavior

The register supports for-each iteration and natural-order sorting. Employees sharing a primary field are ordered using the subsequent comparison fields.

## Automated check requirements

Checks verify family-name/first-name ordering, year/month/day date ordering, Employee name/date/salary tie-breakers, and iteration before and after sorting and removal. Comparison assertions use the sign of `compareTo`, not a particular integer value.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). The embedded behavioral and boundary tests must pass. Compilation is only the first stage; a compiling but incorrect implementation fails.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint: open if you get stuck</summary>

Delegate iterator() to the collection. Use matching primary fields to exercise secondary comparison fields.

</details>
