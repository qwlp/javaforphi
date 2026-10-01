# Java for Phi

Java for Phi turns the material in `material/` into a local, Boot.dev-style course: each lab has a short lesson, a starter project, and an automated check. The checker is a single Go program and works on Windows, macOS, and Linux.

## Quick install

Linux and macOS (requires Git and Bash):

```sh
git clone https://github.com/qwlp/javaforphi.git "$HOME/javaforphi" && bash "$HOME/javaforphi/install.sh"
```

Windows (PowerShell, requires Git):

```powershell
git clone https://github.com/qwlp/javaforphi.git "$HOME\javaforphi"; if ($LASTEXITCODE -eq 0) { powershell -NoProfile -ExecutionPolicy Bypass -File "$HOME\javaforphi\install.ps1" }
```

These commands create a checkout at `~/javaforphi`, then run the installer.
Keep it for `phi update`. If that folder already exists, run its installer
instead. The installer reuses compatible Go or downloads a verified private
SDK. Open a new terminal afterward and run `phi setup`. JDK 17+ is required
for the exercises.

## Terminal interface

In an interactive terminal, `phi` and `phi list` open a Bubble Tea course
browser styled with Charm Lip Gloss: progress, lesson statuses, instructions,
and optional hints are available without remembering commands.

Use **↑/↓** or **j/k** to choose a lesson, **Enter** to open it, **l** to read,
**h** for a hint, **w** to open the Word handout, **c** to check, and **s** to submit. **n** starts the next
incomplete lesson, **r** resumes, **o** runs setup, **?** shows help, and **q**
quits. Instructions scroll with arrow keys or Page Up/Page Down; **Esc** returns.
Actions close the course browser before opening apps; check and submit open
their own live checking screen.

Piped output stays plain. Set `PHI_PLAIN=1` to disable the browser, or
`NO_COLOR=1` to disable its colors. Existing CLI commands remain available.

## Guided learning

Run `phi` to see your progress and the next action. Start with `phi setup`:
it asks for your editor and workflow in an interactive terminal, saves your
preferences, and checks that Java and javac are version 17 or newer.
For scripted setup, use `phi setup --editor intellij --workflow editor`.

- `phi resume` reopens an unfinished lesson; inside a lab it prefers that lesson.
- `phi next` opens the first lesson without a successful submission, in course order.
- `phi list` shows Not started, In progress, or Completed and the check type.
- `phi show` displays the lesson; `phi hint` reveals its optional hint.
- `phi check` reports actionable failures before full compiler or test diagnostics.
- `phi submit` records completion after checks pass, then points you to `phi next`.

`phi next` and `phi resume` accept an optional workspace and the same
`--no-shell`, `--no-open`, and `--no-editor` flags as `phi start`.
Inside a lab, the guided commands use its parent as the workspace; elsewhere
they use `PHI_WORKSPACE` or `~/phi-lessons`. Completion reflects a valid local
receipt from the last successful submission, not automatic checking after edits.

Choose a workflow during setup or with `phi settings set workflow <name>`:

| Workflow | What opens |
| --- | --- |
| `editor` | Your selected IDE, Word handout, and lesson shell |
| `terminal` | Word handout and a shell in the lesson folder |
| `full` | Same as `editor` (supported for compatibility) |

Starting a lesson enters a shell in its folder in every workflow. Use
`--no-shell` to skip it; type `exit` to return to the previous shell.
Older saved `no_shell` and `no_open` preferences are ignored. Interactive
`start`, `go`, `next`, and `resume` open the handout by default; use `--no-open`
to skip it for a particular start.
Setup defaults to the editor workflow; choosing no editor defaults to terminal.
The handout is copied in every workflow. Each lesson includes its goal, source
folders, tasks, expected behavior, checking instructions, and an optional hint.

## Requirements

- JDK 17 or newer (`java` and `javac` on `PATH`)
- Go 1.26 or newer only when building the CLI from source
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
required by `go.mod` (currently 1.26+). They also check for a previously installed
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
phi open       # open its Word handout
phi check      # check the current exercise
phi submit     # check and record a successful local submission
exit           # return to the previous shell
```

The marker in `.phi.json` lets these commands find the lesson root even when you run them from a nested folder such as `src/lib`. Set `PHI_WORKSPACE` to choose a different default workspace, or run `phi start 3 ./my-coursework`. Use `--no-shell` (`-n`) when you only want to create the folder.

Use `phi open` inside a lesson folder (including nested source folders) to
open its handout explicitly. From elsewhere, use `phi open 3` or
`phi open 3 /path/to/workspace`; Phi prepares the lesson if it is missing.
This opens the handout without launching an IDE or entering a shell, and
opens it even when the lesson was started with `--no-open`. Edited handouts are
preserved; missing handouts are restored from the embedded copy.

Use `--no-open` to skip opening the Word document automatically. `--no-shell` controls only
the subshell; use both flags to just prepare the lab. Piped/noninteractive
lesson-start commands never open GUI applications. Explicit `phi open`
requests do open the document, including with redirected input/output. Handouts from `word_doc/` are embedded
in the executable and copied, so the original documents stay available.
Exercises that share a handout open the same file in Phi's configuration folder
under `handouts/`. Your word processor can reuse its existing document window
as you move between these exercises. The shared file preserves your edits;
its first opening uses the current lab's copy, including existing annotations.
Lab folders also retain their own copies for portability. If opening fails,
the CLI prints the document path and you can open it manually.

Eclipse is the default IDE for `phi start` in interactive terminals. You can
change or disable it:

```sh
phi settings set editor intellij
phi settings set editor eclipse
phi settings set editor none       # disable IDE opening
phi settings                      # view settings and their file location
```

Install and verify the Phi plugin for an existing Eclipse installation:

```sh
phi install eclipse
```

If Eclipse is not on PATH, provide its launcher or macOS application bundle:

```sh
phi install eclipse "/path/to/eclipse"
# Alias: phi plugin install eclipse "/path/to/eclipse"
```

On Windows, pass the path to `eclipse.exe`. This command builds Phi's bundled
import plugin, verifies it in a temporary headless workspace, then selects
Eclipse as your editor and enables IDE opening. Existing settings are saved
only after verification succeeds. A custom `eclipse-workspace` is preserved.
Rerunning the command reuses the matching cached plugin configuration.
It requires Eclipse to be installed already; it does not download the IDE.

On Windows, `phi update` also installs and verifies this plugin automatically
when Eclipse is selected and IDE opening is enabled. It searches `PATH` and
standard Eclipse Installer folders, or uses your saved `editor-path`. It uses
Eclipse's console launcher for verification and can find `javac` through
`JAVA_HOME` when it is missing from `PATH`. Other editor selections are preserved.
Eclipse Installer's shared `.p2` bundle pool is supported; plugin locations are
resolved from Eclipse's installed bundle index.
If automatic discovery fails, run `phi install eclipse "C:\path\to\eclipse.exe"`.

The plugin lives in Phi's private configuration. Eclipse's installed files
remain untouched, so no administrator access is required. Start Eclipse
through `phi next` or `phi go <number>` to load this configuration. Close an
existing Eclipse window once if it uses an earlier Phi configuration.

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
standard Eclipse `plugins/` and `configuration/` layout (including NixOS packages
with a `bin/eclipse` launcher and an `eclipse/` installation) and a compatible
Java compiler. Phi uses Eclipse's bundled `javac` when available, otherwise your
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
All shipped lessons require their behavioral tests to pass.

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

All 15 shipped lessons run behavioral JUnit checks after compilation.
`phi check` and `phi submit` show a live Bubble Tea screen in a terminal:
real preparation/compilation/test stages, a spinner, individual case results,
and passed/failed counts. Press **d** after completion to expand full diagnostics,
**↑/↓** to scroll, and **Enter** to close. **Ctrl+C** or **q** cancels a running
check. Scripted and piped commands retain plain output and exit codes.

The eight formerly compile-only lessons now have instructor-owned suites for
console output, input boundaries, object delegation, collection mutation,
ordering, wrapping counters, and pricing policies. Console input contracts and
small testable calculation methods are described in each lesson's
`Automated check requirements` section. Some worked examples already pass;
unimplemented tasks and the playlist starter's boundary bugs fail.

Grading cases run from pristine embedded copies in a temporary directory.
Additional suites under `grading/` are never copied into learner folders;
edits to a lab's test directory do not change grading. This keeps cases hidden
from the lab workflow, but an offline open-source checker cannot keep test
sources secret from someone inspecting the repository or executable.

`phi submit` always checks current files and records completion only after
checks pass. Failed or cancelled attempts preserve an earlier receipt.
Legacy compile-only receipts for newly graded lessons no longer count as
completion; submit again to run the behavioral checks.
See [course/README.md](course/README.md) for authoring and verification.

## Cross-compile releases

From any Go-supported host:

```text
GOOS=windows GOARCH=amd64 go build -o dist/phi-windows-amd64.exe .
GOOS=linux   GOARCH=amd64 go build -o dist/phi-linux-amd64 .
GOOS=darwin  GOARCH=arm64 go build -o dist/phi-darwin-arm64 .
```

The included GitHub Actions workflow builds Windows, Linux, and macOS artifacts for version tags.
