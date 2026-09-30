# Portfolio A.1: Player Composition

Implement `lib.Player` as a composition of `Name`, `PairOfDice`, and a gamer-tag `String`.

## Required API

- A default constructor, `Player(Name, String)`, and `Player(PairOfDice, Name, String)`.
- Get/set methods for the name and gamer tag; a getter for the pair of dice.
- `rollDice()` and `getDiceScore()` methods that delegate to `PairOfDice`.
- A conventional `toString()` containing the three fields.
- Complete Javadoc for the class, constructors, and public methods.

Run `phi check`. Instructor-owned tests verify constructor identity, delegation, accessors, and the string representation.
