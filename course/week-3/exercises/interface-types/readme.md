# Interface Types

## Goal

Use interfaces to express shared capabilities independently of a class hierarchy.

## Files to work on

Starter source folders: `src/lib/iterable_comparable/`, `src/lib/measurable/`, `src/lib/polymorphism/`, `src/main/`. Follow the tasks below for the classes to edit or create. The copied Word handout contains the full lab specification.

## Tasks

1. Study `Rollable` polymorphism.
2. Make `PlayList` and `MultipleDice` implement `Iterable` by delegating `iterator()` to their internal lists.
3. Make `Song` implement `Comparable<Song>` by title, artist, then duration; add `PlayList.sortPlaylist()`.
4. Create `lib.measurable.Measurable` with `int getMeasure()`.
5. Implement it in measurable `Die`, `Song`, and `Name`, returning score, duration, and full-name length.
6. Ensure `DataAnalysis` is in `lib.measurable`; add `avg()`, `min()`, and `max()`, returning `-1` for an empty analysis.

## Expected behavior

Collections support for-each iteration; songs sort by title, artist, then duration; measurable objects report their specified measures; empty analyses return -1.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). The behavioral tests must pass. Use failures to identify the method or boundary case to revisit.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint — open if you get stuck</summary>

Implement one interface at a time. For compareTo, compare the next field only when the previous comparison returns zero.

</details>
