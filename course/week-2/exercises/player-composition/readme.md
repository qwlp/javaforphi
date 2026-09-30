# Portfolio A.1: Player Composition

## Goal

Implement `lib.Player` as a composition of `Name`, `PairOfDice`, and a gamer-tag `String`.

## Files to work on

Starter source folders: `src/lib/`. Follow the tasks below for the classes to edit or create. The copied Word handout contains the full lab specification.

## Required API

- A default constructor, `Player(Name, String)`, and `Player(PairOfDice, Name, String)`.
- Get/set methods for the name and gamer tag; a getter for the pair of dice.
- `rollDice()` and `getDiceScore()` methods that delegate to `PairOfDice`.
- A conventional `toString()` containing the three fields.
- Complete Javadoc for the class, constructors, and public methods.

## Expected behavior

Player preserves supplied object references, exposes its name and gamer tag, and delegates rolling and scoring to its PairOfDice.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). The behavioral tests must pass. Use failures to identify the method or boundary case to revisit.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint: open if you get stuck</summary>

Keep the supplied Name and PairOfDice references. Delegate rolling and score retrieval to the stored dice object.

</details>
