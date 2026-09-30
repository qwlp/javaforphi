# Portfolio A: Polymorphic Player

## Goal

Copy your earlier `Player` into this project, generalize it from `PairOfDice` to any `Rollable`, and finish the portfolio behavior.

## Files to work on

Starter source folders: `src/lib/`, `src/main/`. Follow the tasks below for the classes to edit or create. The copied Word handout contains the full lab specification.

## Required behavior

- Store exactly three private fields: `Rollable`, `Name`, and `String`.
- Support the constructors and methods used by `PlayerTest`, including `getRollable`.
- `setFullPlayerName` parses two names and normalizes their capitalization.
- `generateGameTag(1..100)` reverses the lowercase, whitespace-free full name and appends the number; invalid values do nothing.
- `PlayerApp.execute` filters dynamically and returns names in the required `FIRSTNAME, surname` format.
- Implement `Comparable<Player>` by name, then gamer tag.

## Expected behavior

Player works with any Rollable, normalizes full names, validates gamer-tag numbers, and orders players by name then gamer tag. The app formats filtered names as specified.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). The behavioral tests must pass. Use failures to identify the method or boundary case to revisit.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint — open if you get stuck</summary>

Use the Rollable interface for the field and constructor parameter. Validate the gamer-tag number before changing any state.

</details>
