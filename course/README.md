# Course authoring

The layout mirrors the useful parts of the reference Boot.dev assets without copying their content:

```text
course/
  catalog.json
  week-N/exercises/lesson-name/
    readme.md
    tests/                 # optional instructor-owned tests
material/
  week_N/labs/Starter.zip  # embedded starter project
```

`catalog.json` is the source of truth. A lesson selects a starter archive and either a `compile` check or a `junit4` check. JUnit tests may come from the original archive (`test_source: archive`) or the lesson's private test tree (`test_source: course`). Test sources are compiled from the executable's embedded copy, not the learner's directory. All shipped lessons use behavioral checks. Additional suites live under `grading/` and are not copied into labs; `support_path` optionally adds shared Java test utilities.

Lesson numbers follow the order of entries in `catalog.json`. Keep that order stable after learners have begun the course. `phi start <number>` creates a numbered folder, copies the lesson readme to `LESSON.md`, extracts the starter, and records the stable lesson ID in `.phi.json`. The optional `document` field names an embedded handout under `word_doc/`; it is copied to the lab folder using its original filename. Several lessons can share a handout. Interactive starts open it in a word processor unless `--no-open` is specified.

To add a lesson:

1. Add its starter ZIP beneath `material/`.
2. Add a concise `readme.md` with the goal, task, constraints, and check command.
3. Add the lesson to `catalog.json`.
4. For behavioral grading, add JUnit tests and list their fully qualified class names.
5. Run `go test ./...`, then initialize a clean starter and verify that its expected failures are understandable.

Compile checks prove only that the source is valid Java. Use behavioral tests for assessed work. Avoid tests that depend on random output, wall-clock time, network access, Eclipse metadata, or a particular operating system path separator.

The bundled JUnit runner reports individual case results to the terminal UI.
A suite with zero executed cases, ignored cases, or skipped assumptions cannot
establish completion. Tests should use descriptive method names and helpful
assertion messages. Hidden means absent from learner folders, not secret from
readers of this open-source repository.

To run integration checks for the shipped suites, cache the pinned JUnit and
Hamcrest JARs with one normal graded check, then run
`PHI_TEST_CACHE=/path/to/dependency-cache go test ./...`. These checks also mutate
otherwise compiling programs and verify that incorrect behavior fails and
learner-edited tests cannot bypass grading.
