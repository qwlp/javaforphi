# Java for Phi

Java for Phi turns the material in `material/` into a local, Boot.dev-style course: each lab has a short lesson, a starter project, and an automated check. The checker is a single Go program and works on Windows, macOS, and Linux.

## Requirements

- JDK 17 or newer (`java` and `javac` on `PATH`)
- Go 1.22 or newer only when building the CLI from source
- An internet connection on the first graded check, so the CLI can cache JUnit 4 and Hamcrest

No Maven, Gradle, or Eclipse installation is required. Eclipse projects remain supported.

## Build and use

```text
go build -o phi .              # Windows: go build -o phi.exe .
./phi doctor
./phi list
./phi 3
```

`phi 3` creates lesson 3 under the default `phi-lessons` folder in your home directory, copies the instructions to `LESSON.md`, extracts the starter exercise, and opens a subshell inside the exercise. From there:

```text
phi show       # display the current lesson
phi check      # check the current exercise
exit           # return to the previous shell
```

The marker in `.phi.json` lets these commands find the lesson root even when you run them from a nested folder such as `src/lib`. Set `PHI_WORKSPACE` to choose a different default workspace, or run `phi start 3 ./my-coursework`. Use `--no-shell` (`-n`) when you only want to create the folder.

In PowerShell, use `./phi.exe` (or `.\phi.exe` on older PowerShell) in the same commands.

Every frequent command has a shorthand:

| Command | Shorthand |
| --- | --- |
| `phi start 3` | `phi 3`, `phi go 3`, or `phi g 3` |
| `phi list` | `phi l` or `phi ls` |
| `phi show` | `phi s` |
| `phi init` | `phi i` |
| `phi check` | `phi c` |
| `phi check-all` | `phi ca` |
| `phi doctor` | `phi d` |
| `phi version` | `phi v` |
| `phi help` | `phi h` |

Initialize and check the complete course:

```text
./phi init --all
./phi check-all
```

These commands use the same default workspace and numbered folders, such as `01-basic-java-programs` and `03-custom-classes`.

The starter archives, lesson text, and pristine grading tests are embedded in the executable. A learner cannot make a failing check pass by editing the copied test files. Dependencies are downloaded from Maven Central, pinned by version, and checked against their SHA-256 digest before use. Set `PHI_CACHE` to override the dependency cache directory. The old `JAVAFORPHI_CACHE` variable and `.javaforphi.json` project markers are still recognized for migration from v0.1.0.

## Checker coverage

The original JUnit suites grade the custom-class, Register, and Player portfolios. Additional tests grade Player composition, inheritance, and interface-type exercises. Demonstration-oriented projects currently receive compile checks. See [course/README.md](course/README.md) for the lesson format and extension points.

## Cross-compile releases

From any Go-supported host:

```text
GOOS=windows GOARCH=amd64 go build -o dist/phi-windows-amd64.exe .
GOOS=linux   GOARCH=amd64 go build -o dist/phi-linux-amd64 .
GOOS=darwin  GOARCH=arm64 go build -o dist/phi-darwin-arm64 .
```

The included GitHub Actions workflow builds Windows, Linux, and macOS artifacts for version tags.
