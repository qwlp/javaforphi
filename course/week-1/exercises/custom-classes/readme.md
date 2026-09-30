# Your First Custom Classes

## Goal

Build classes with private state, constructors, accessors, behavior, and useful `toString`/`equals` implementations.

## Files to work on

Starter source folders: `src/lib/`, `src/main/`. Follow the tasks below for the classes to edit or create. The copied Word handout contains the full lab specification.

## Tasks

1. Add `decrement()` to `Counter` and use it from `CounterApp`.
2. Study `Name`, `NameDemo`, and their tests.
3. Create `src/lib/CDTrack.java` with private `title`, `artist`, and `duration` fields.
4. Supply default and three-argument constructors, getters, setters, and a conventional `toString()`.
5. Keep duration non-negative: a negative constructor/setter value becomes or leaves it at zero.

## Expected behavior

Counters decrement correctly. CDTrack exposes its title, artist, and duration through constructors and accessors, and duration never becomes negative.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). The behavioral tests must pass. Use failures to identify the method or boundary case to revisit.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint — open if you get stuck</summary>

Start with constructors and getters so tests can inspect state. Validate duration in both the constructor and setter.

</details>
