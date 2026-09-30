param([string]$Workspace = '')
$ErrorActionPreference = 'Stop'
if ($env:OS -ne 'Windows_NT') { throw 'Use uninstall.sh on Linux or macOS.' }
$installDirs = @((Join-Path $env:LOCALAPPDATA 'Programs\Phi'), (Join-Path $env:LOCALAPPDATA 'Phi'))
# Validate every target before removing either installation.
foreach ($installDir in $installDirs) {
    if (!(Test-Path $installDir)) { continue }
    if (((Get-Item $installDir).Attributes -band [IO.FileAttributes]::ReparsePoint) -or !(Test-Path (Join-Path $installDir '.phi-install'))) {
        throw "Refusing to remove unmanaged installation: $installDir"
    }
}
$workspaces = @((Join-Path $env:USERPROFILE 'phi-lessons'))
if ($env:PHI_WORKSPACE) { $workspaces += $env:PHI_WORKSPACE }
if ($Workspace) { $workspaces += $Workspace }
$labs = @{}
foreach ($root in $workspaces) {
    if (!(Test-Path -LiteralPath $root -PathType Container)) { continue }
    if ((Get-Item -LiteralPath $root).Attributes -band [IO.FileAttributes]::ReparsePoint) { continue }
    foreach ($lab in Get-ChildItem -LiteralPath $root -Directory) {
        if ($lab.Attributes -band [IO.FileAttributes]::ReparsePoint) { continue }
        if ($lab.FullName -eq $env:USERPROFILE -or $lab.FullName -eq $PSScriptRoot) { continue }
        foreach ($name in @('.phi.json', '.javaforphi.json')) {
            $marker = Join-Path $lab.FullName $name
            if (!(Test-Path -LiteralPath $marker -PathType Leaf)) { continue }
            if ((Get-Item -LiteralPath $marker).Attributes -band [IO.FileAttributes]::ReparsePoint) { continue }
            try { $metadata = Get-Content -LiteralPath $marker -Raw | ConvertFrom-Json } catch { continue }
            if ($metadata.lesson -is [string] -and $metadata.lesson) { $labs[$lab.FullName] = $lab.FullName; break }
        }
    }
}
$recoveryDir = $null
if ($labs.Count -gt 0) {
    Write-Host 'Generated lab folders found (including your work and documents):'
    $labPaths = @($labs.Values | Sort-Object)
    foreach ($lab in $labPaths) { Write-Host "  $lab" }
    Write-Host 'Removing them will move them to a recovery folder so you can restore them.'
    $answer = ''
    try { $answer = Read-Host 'Remove these generated lab folders? [y/N]' } catch { $answer = '' }
    if ($answer -match '^(y|yes)$') {
        $recoveryBase = Join-Path $env:LOCALAPPDATA 'PhiLabBackups'
        if ((Test-Path -LiteralPath $recoveryBase) -and ((Get-Item -LiteralPath $recoveryBase).Attributes -band [IO.FileAttributes]::ReparsePoint)) {
            throw "Refusing to use linked recovery directory: $recoveryBase"
        }
        $recoveryDir = Join-Path $recoveryBase ([Guid]::NewGuid().ToString())
        New-Item -ItemType Directory -Path $recoveryDir -Force | Out-Null
        $index = 0
        foreach ($lab in $labPaths) {
            $item = Get-Item -LiteralPath $lab
            if (!$item.PSIsContainer -or ($item.Attributes -band [IO.FileAttributes]::ReparsePoint)) { throw "Lab folder changed: $lab" }
            $index++
            $destination = Join-Path $recoveryDir $index
            New-Item -ItemType Directory -Path $destination | Out-Null
            Set-Content -LiteralPath (Join-Path $destination 'original-path.txt') -Value $lab
            Move-Item -LiteralPath $lab -Destination $destination
        }
        Write-Host "Lab folders removed from their workspaces. Restore them from: $recoveryDir"
    } else { Write-Host 'Generated lab folders were kept.' }
}
foreach ($installDir in $installDirs) {
if (!(Test-Path $installDir)) { continue }
if (((Get-Item $installDir).Attributes -band [IO.FileAttributes]::ReparsePoint) -or !(Test-Path (Join-Path $installDir '.phi-install'))) {
    throw "Refusing to remove unmanaged installation: $installDir"
}
$tracking = Join-Path $installDir 'path-added.json'
if (Test-Path $tracking) {
    $added = @(Get-Content $tracking -Raw | ConvertFrom-Json)
    # Restrict removal even if the tracking file was modified.
    $owned = @((Join-Path $installDir 'bin'), (Join-Path $installDir 'go\bin'))
    $added = @($added | Where-Object { $owned -contains $_ })
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    [Environment]::SetEnvironmentVariable('Path', (($userPath -split ';' | Where-Object { $added -notcontains $_ }) -join ';'), 'User')
    $env:Path = ($env:Path -split ';' | Where-Object { $added -notcontains $_ }) -join ';'
}
Remove-Item -Recurse -Force $installDir
}
Write-Host 'Removed installer-owned phi, Go, and PATH entries. Restart other terminals.'
if (!$recoveryDir) { Write-Host 'Your lessons were kept.' }
Write-Host 'Your dependency cache, settings, and other Go installations were kept.'
