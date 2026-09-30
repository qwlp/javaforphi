# Portfolio A: Polymorphic Player

Copy your earlier `Player` into this project, generalize it from `PairOfDice` to any `Rollable`, and finish the portfolio behavior.

## Required behavior

- Store exactly three private fields: `Rollable`, `Name`, and `String`.
- Support the constructors and methods used by `PlayerTest`, including `getRollable`.
- `setFullPlayerName` parses two names and normalizes their capitalization.
- `generateGameTag(1..100)` reverses the lowercase, whitespace-free full name and appends the number; invalid values do nothing.
- `PlayerApp.execute` filters dynamically and returns names in the required `FIRSTNAME, surname` format.
- Implement `Comparable<Player>` by name, then gamer tag.

Run `phi check` to execute the complete embedded portfolio suite.
