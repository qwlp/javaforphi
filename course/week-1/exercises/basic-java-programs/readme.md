# Basic Java Programs

## Goal

Learn Java's primitive types and the three basic control-flow tools: selection, iteration, and arrays.

## Files to work on

Starter source folders: `src/arrays/`, `src/controlstructures/`, `src/primitives/`. Follow the tasks below for the classes to edit or create. The copied Word handout contains the full lab specification.

## Tasks

1. In `primitives/Converter.java`, use `double` for both temperatures and observe the difference from integer division.
2. Study `ConverterNew.java`, including how the boolean selects a conversion direction.
3. Create `controlstructures/GradeMark.java`. Print `Fail`, `Pass`, `Merit`, or `Distinction` for the ranges `<40`, `40–59`, `60–69`, and `70+`.
4. Create `controlstructures/DaysOfWeek.java` with switch statements for the day name and weekday/weekend classification.
5. Rewrite `TimesTable` with a `while` loop, then use nested loops for the tables 1 through 12.
6. Study the traversal and calculations in `arrays/NumberArrayDemo.java`.

## Expected behavior

Temperature conversions retain fractional values; grade labels change at 40, 60, and 70; day classification and multiplication tables match the selected inputs.

## Automated check requirements

For automated checks, `GradeMark.main` reads the mark from `args[0]` (or from standard input) and prints only the grade label. `DaysOfWeek.main` reads the day the same way and prints its name and classification; out-of-range days print `Unknown day`. Keep the 21°C example in `Converter` and preserve its fractional result. `TimesTable` prints all 144 products for tables 1–12 in table order, separated by whitespace, without numeric labels. These input extensions let the checker try boundary values rather than just your hardcoded example.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). The embedded behavioral and boundary tests must pass. Compilation is only the first stage; a compiling but incorrect implementation fails.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint: open if you get stuck</summary>

Use a decimal literal such as `5.0` when you need floating-point division. Test grade boundaries at 39, 40, 59, 60, 69, and 70.

</details>
