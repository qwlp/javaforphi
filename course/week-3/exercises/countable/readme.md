# Countable Interface

## Goal

Study a compact example of polymorphism through the `Countable` interface.

## Files to work on

Starter source folders: `src/lib/`, `src/main/`. Follow the tasks below for the classes to edit or create. The copied Word handout contains the full lab specification.

## Tasks

1. Identify the methods guaranteed by `Countable`.
2. Compare how `Counter` and `ModuloCounter` implement those methods.
3. Trace the interface-typed variables in `CountableDemo` and the runtime dispatch of each call.
4. Add another interaction that treats both implementations uniformly.

## Expected behavior

Calls through Countable dispatch to the appropriate Counter or ModuloCounter behavior without requiring implementation-specific code in the caller.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). All sources must compile. Run the relevant demo classes in your IDE and compare their output with the tasks; compilation alone does not verify behavior.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint — open if you get stuck</summary>

Read the interface before the implementations. The declared type determines available methods; the runtime object determines their behavior.

</details>
