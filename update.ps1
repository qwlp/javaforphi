$ErrorActionPreference = 'Stop'
if ($env:OS -ne 'Windows_NT') { throw 'Use update.sh on Linux or macOS.' }
if (!(Get-Command git -ErrorAction SilentlyContinue)) { throw 'Updating requires Git and a cloned checkout.' }
$root = & git -C $PSScriptRoot rev-parse --show-toplevel 2>$null
if ($LASTEXITCODE -ne 0 -or !$root -or [IO.Path]::GetFullPath($root) -ne [IO.Path]::GetFullPath($PSScriptRoot)) {
    throw 'Updating requires a cloned checkout. For a ZIP download, download the new source and run install.ps1.'
}
$changes = & git -C $PSScriptRoot status --porcelain
if ($LASTEXITCODE -ne 0) { throw 'Could not inspect the checkout.' }
if ($changes) { throw 'Commit or stash your checkout changes before updating.' }
& git -C $PSScriptRoot rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'Your current branch needs an upstream before updating.' }
Write-Host 'Updating the checkout from its upstream...'
& git -C $PSScriptRoot pull --ff-only
if ($LASTEXITCODE -ne 0) { throw 'Updating the checkout failed; phi was not reinstalled.' }
& (Join-Path $PSScriptRoot 'install.ps1')
Write-Host 'phi is updated.'
