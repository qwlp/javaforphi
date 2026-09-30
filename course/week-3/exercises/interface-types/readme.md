# Interface Types

Use interfaces to express shared capabilities independently of a class hierarchy.

## Tasks

1. Study `Rollable` polymorphism.
2. Make `PlayList` and `MultipleDice` implement `Iterable` by delegating `iterator()` to their internal lists.
3. Make `Song` implement `Comparable<Song>` by title, artist, then duration; add `PlayList.sortPlaylist()`.
4. Create `lib.measurable.Measurable` with `int getMeasure()`.
5. Implement it in measurable `Die`, `Song`, and `Name`, returning score, duration, and full-name length.
6. Ensure `DataAnalysis` is in `lib.measurable`; add `avg()`, `min()`, and `max()`, returning `-1` for an empty analysis.

Run `phi check` for behavioral checks of all four capabilities.
