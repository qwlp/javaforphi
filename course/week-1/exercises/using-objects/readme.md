# Using Objects

## Goal

Practice using Java API classes, especially `String` and `Scanner`, and distinguish object references from primitive values.

## Files to work on

Starter source folders: `src/scanner/`, `src/strings/`. Follow the tasks below for the classes to edit or create. The copied Word handout contains the full lab specification.

## Tasks

1. Extend `StringDemo` with `toUpperCase`, `replace`, `substring`, and `indexOf`.
2. Create `strings/Initials.java`; turn a two-part name into initials and a lowercase `initials@email.dmu.ac.uk` address.
3. Create `StringArrayDemo` and print six fruit names in uppercase.
4. Create `ImmutableDemo` and explain why ignoring the value returned by `toLowerCase()` leaves the original string unchanged.
5. Extend `ScannerDemo` to read and process keyboard input.
6. Study the input pitfall demonstrated by `ScannerSurprise`.

## Expected behavior

String transformations produce the requested initials, email address, and case conversions. Keyboard input is read correctly even when numeric and line input are mixed.

## Automated check requirements

For automated checks, `Initials.main` accepts a two-part full name in `args[0]` (or one standard-input line), then prints uppercase initials and the lowercase email address. `ImmutableDemo.main` accepts a string the same way and prints both the unchanged original and its returned lowercase copy. `StringArrayDemo` prints exactly six uppercase fruit names separated by whitespace. Fix `ScannerSurprise` to consume the newline after `nextDouble()` and print the last sentence or its correct length.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). The embedded behavioral and boundary tests must pass. Compilation is only the first stage; a compiling but incorrect implementation fails.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint — open if you get stuck</summary>

String methods return a new value. After reading a number with Scanner, check whether the trailing newline still needs to be consumed.

</details>
