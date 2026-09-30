# Java for Phi

Java for Phi turns the material in `material/` into a local, Boot.dev-style course: each lab has a short lesson, a starter project, and an automated check. The checker is a single Go program and works on Windows, macOS, and Linux.

## Requirements

- JDK 17 or newer (`java` and `javac` on `PATH`)
- Go 1.22 or newer only when building the CLI from source
- An internet connection on the first graded check, so the CLI can cache JUnit 4 and Hamcrest

No Maven, Gradle, or Eclipse installation is required. Eclipse projects remain supported.

## Install and uninstall

Download or clone this repository, then run the installer from the checkout.
Linux and macOS (Intel and Apple Silicon):

```sh
bash install.sh
```

Windows (x86-64 or ARM64), in PowerShell:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\install.ps1
```

The scripts first check for Go on PATH and reuse it if it meets the version
required by `go.mod` (currently 1.22+). They also check for a previously installed
private SDK. Only if neither is suitable do they download Go from go.dev and
verify its SHA-256 checksum. They build `phi` and add it to your user PATH;
a downloaded private SDK is also available on PATH. No administrator privileges are needed.
Linux/macOS require Bash; downloading Go additionally requires curl, tar, and
sha256sum or shasum. Windows requires
PowerShell 5.1 or newer. Open a new terminal after installation, then run
`phi doctor` or `phi 3`. JDK 17+ is still required for Java exercises.

On Linux, the command is installed as `~/.local/bin/phi` (a symlink to the
managed executable under `~/.local/share/phi/bin`). The installer automatically
adds `~/.local/bin` to PATH. On Windows, the executable is installed in
`%LOCALAPPDATA%\Programs\Phi\bin`, which is automatically added to your user
PATH. Reinstalling migrates the earlier `%LOCALAPPDATA%\Phi` layout.
macOS uses `~/.local/share/phi/bin`. The Unix installer configures Bash, Zsh,
POSIX login shells, and Fish. Re-running the installer rebuilds `phi` and
reuses compatible Go without adding duplicate PATH entries.

### Updating

After installation, update from any directory:

```sh
phi update    # shorthand: phi u
```

The installer records your checkout location. Keep that checkout available;
if you move it, rerun the installer. `phi update` updates the current branch
from its configured upstream and reinstalls `phi`. Your checkout must
be clean. Copies built manually must run the installer first to enable this command.

You can also run the update scripts directly from the checkout:

```sh
bash update.sh
```

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\update.ps1
```

Updates are manual. Both installers reuse compatible Go; if Go is missing or
too old, they install the latest stable release from go.dev and verify its
published checksum. The updater uses `git pull --ff-only`
and stops if the checkout has local changes or the branch has diverged.
For a ZIP download, download the new source and run the installer again.
You can also rerun the installer to rebuild your current source without pulling
source changes. Existing compatible Go versions are kept. Download and build failures leave the installed
`phi` and Go untouched; if reinstalling fails after a successful pull, the
checkout stays updated and you can rerun the installer. Lessons are preserved.

To uninstall from the checkout:

```sh
bash uninstall.sh
```

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\uninstall.ps1
```

Uninstalling removes the private Go SDK, `phi`, and installer-added PATH
configuration. It lists generated lab folders and asks whether to remove them,
defaulting to **No** (including on end-of-input). Only direct lab folders with
a `.phi.json` or legacy `.javaforphi.json` marker are selected; unrelated
folders and linked directories are kept. It searches `~/phi-lessons` and
`PHI_WORKSPACE`. For another workspace, use
`bash uninstall.sh --workspace /path/to/workspace` or append
`-Workspace 'C:\path\to\workspace'` to the
PowerShell uninstall command.

Choosing Yes moves the listed labs, including your work and Word documents,
to `~/.local/share/phi-lab-backups` on Linux/macOS or
`%LOCALAPPDATA%\PhiLabBackups` on Windows. The script prints the recovery
location; each numbered backup includes `original-path.txt` so you can move
the lab back. This removes labs from their workspaces without permanently
erasing them. Dependency caches, settings, and other Go installations are kept.
Restart your terminal afterward. Keep the install directory for installer-managed
files only, since uninstalling removes it entirely.

## Build and use

```text
go build -o phi .              # Windows: go build -o phi.exe .
./phi doctor
./phi list
./phi 3
```

`phi 3` creates lesson 3 under the default `phi-lessons` folder in your home directory,
copies the instructions to `LESSON.md` and the original Word lab handout into the
folder, extracts the starter exercise, and opens a subshell inside the exercise.
In an interactive terminal it also opens the handout in a word processor:
Microsoft Word when available on Windows/macOS, LibreOffice Writer or OpenOffice
on Linux, or the default document application as a fallback. No word processor
is installed automatically. From the lab shell:

```text
phi show       # display the current lesson
phi check      # check the current exercise
phi submit     # check and record a successful local submission
exit           # return to the previous shell
```

The marker in `.phi.json` lets these commands find the lesson root even when you run them from a nested folder such as `src/lib`. Set `PHI_WORKSPACE` to choose a different default workspace, or run `phi start 3 ./my-coursework`. Use `--no-shell` (`-n`) when you only want to create the folder.

Use `--no-open` to skip opening the Word document. `--no-shell` controls only
the subshell; use both flags to just prepare the lab. Piped/noninteractive
commands never open GUI applications. Handouts from `word_doc/` are embedded
in the executable and copied, so the original documents stay available.
Shared handouts are copied to each relevant lab. Reopening a lab preserves
document edits and restores the handout if it is missing. If opening fails,
the CLI prints the document path and you can open it manually.

Eclipse is the default IDE for `phi start` in interactive terminals. You can
change or disable it:

```sh
phi settings set editor intellij
phi settings set editor eclipse
phi settings set editor none       # disable IDE opening
phi settings                      # view settings and their file location
```

Select one editor. IntelliJ opens the lab folder directly. Eclipse automatically
imports the lab into its shared workspace **at the original lab path**. It does
not copy your source files. Eclipse, `phi check`, and `phi submit` therefore use
the same files. New lab imports are queued to an already-running Phi-managed
Eclipse instance; starting the same lab again reopens/refreshes the existing
project. Projects with matching names in different lab folders get distinct
workspace names so neither is replaced.

Phi builds a small bundled Eclipse helper on first use and caches it in a
private configuration next to your settings. The installed Eclipse files are
not modified, and no third-party plugin is downloaded. This requires the
standard Eclipse `plugins/` and `configuration/` layout and a compatible Java
compiler. Phi uses Eclipse's bundled `javac` when available, otherwise your
JDK. After updating Eclipse or switching from an older Phi setup, close its
existing workspace once and let `phi start` reopen it with the helper.

If the IDE launcher is not on PATH, set its full executable path after choosing
the editor (switching editors clears the previous custom launcher):

```sh
phi settings set editor-path "/path/to/idea/bin/idea.sh"
phi settings set eclipse-workspace "/path/to/eclipse-workspace"
```

On Windows, use the full path to `idea64.exe` or `eclipse.exe`; on macOS, you
can also select an `.app` bundle. Without a custom path, macOS tries the
IntelliJ IDEA or Eclipse application. Use `--no-editor` to skip the IDE for
one start. This is independent of `--no-open` (Word) and `--no-shell`.
IDEs must already be installed. Settings are saved in your OS user config
directory under `phi/settings.json`; `PHI_CONFIG` can override that file path.

`phi start` prints the lab folder and instructions for reading, testing, and
submitting, including when you use `--no-shell` or reopen an existing lab.
`phi submit` runs the same checks as `phi check` and saves a local receipt in
`.phi-submission.json` only after they pass. The receipt includes the lesson,
timestamp, check type, and CLI version. It records completion at that time;
rerun `phi submit` after changing your work. There is no remote upload or
submission server. You can also run `phi submit 3 /path/to/lab` from elsewhere.
Compile-only lessons require successful compilation; graded lessons require
their tests to pass.

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
| `phi update` | `phi u` |
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
