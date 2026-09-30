$ErrorActionPreference = 'Stop'
if ($env:OS -ne 'Windows_NT') { throw 'Use install.sh on Linux or macOS.' }
$installDir = Join-Path $env:LOCALAPPDATA 'Programs\Phi'
$legacyDir = Join-Path $env:LOCALAPPDATA 'Phi'
$marker = Join-Path $installDir '.phi-install'
if (Test-Path $installDir) {
    if (((Get-Item $installDir).Attributes -band [IO.FileAttributes]::ReparsePoint) -or !(Test-Path $marker)) {
        throw "Refusing to overwrite unmanaged installation: $installDir"
    }
}
$arch = $env:PROCESSOR_ARCHITECTURE
if ($env:PROCESSOR_ARCHITEW6432) { $arch = $env:PROCESSOR_ARCHITEW6432 }
$goArch = switch ($arch) { 'AMD64' { 'amd64' } 'ARM64' { 'arm64' } default { throw 'Supported architectures: x86-64 and ARM64.' } }
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
$stage = Join-Path ([IO.Path]::GetTempPath()) ([Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $stage | Out-Null
try {
    $goLine = Get-Content (Join-Path $PSScriptRoot 'go.mod') | Where-Object { $_ -match '^go\s+(\d+\.\d+(?:\.\d+)?)\s*$' } | Select-Object -First 1
    if (!$goLine) { throw 'Cannot read the required Go version from go.mod.' }
    $minimumGo = [version](($goLine -split '\s+')[1])
    if ($minimumGo.Build -lt 0) { $minimumGo = [version]"$minimumGo.0" }
    $goCommand = $null
    $existingGo = Get-Command go -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1
    $candidates = @()
    if ($existingGo) { $candidates += $existingGo.Source }
    $candidates += Join-Path $installDir 'go\bin\go.exe'
    if (Test-Path (Join-Path $legacyDir '.phi-install')) { $candidates += Join-Path $legacyDir 'go\bin\go.exe' }
    $savedToolchain = $env:GOTOOLCHAIN
    try {
        $env:GOTOOLCHAIN = 'local'
        foreach ($candidate in $candidates) {
            if (!(Test-Path $candidate)) { continue }
            $versionOutput = & $candidate version 2>$null
            if ($LASTEXITCODE -eq 0 -and "$versionOutput" -match '\bgo(\d+\.\d+(?:\.\d+)?)\s') {
                $installedVersion = [version]$Matches[1]
                if ($installedVersion.Build -lt 0) { $installedVersion = [version]"$installedVersion.0" }
                if ($installedVersion -ge $minimumGo) {
                    $goCommand = $candidate
                    Write-Host "Using existing Go $installedVersion`: $goCommand"
                    break
                }
            }
        }
    } finally { $env:GOTOOLCHAIN = $savedToolchain }
    if (!$goCommand) {
    Write-Host "Go $minimumGo or newer was not found; installing a private SDK."
    $releases = Invoke-RestMethod 'https://go.dev/dl/?mode=json' -TimeoutSec 60
    $release = $releases | Where-Object stable | Select-Object -First 1
    $archive = $release.files | Where-Object { $_.os -eq 'windows' -and $_.arch -eq $goArch -and $_.kind -eq 'archive' } | Select-Object -First 1
    if (!$archive) { throw "No official Go archive for Windows/$goArch." }
    Write-Host "Downloading $($release.version) for Windows/$goArch..."
    $zip = Join-Path $stage 'go.zip'
    Invoke-WebRequest "https://go.dev/dl/$($archive.filename)" -OutFile $zip -UseBasicParsing -TimeoutSec 600
    if ((Get-FileHash $zip -Algorithm SHA256).Hash -ne $archive.sha256) { throw 'Go checksum verification failed.' }
    Expand-Archive $zip -DestinationPath $stage
    $goCommand = Join-Path $stage 'go\bin\go.exe'
    }
    $oldGoRoot = $env:GOROOT
    $oldToolchain = $env:GOTOOLCHAIN
    $oldCgo = $env:CGO_ENABLED
    Push-Location $PSScriptRoot
    try {
        if (Test-Path (Join-Path $stage 'go')) { $env:GOROOT = Join-Path $stage 'go' }
        $env:GOTOOLCHAIN = 'local'
        $env:CGO_ENABLED = '0'
        & $goCommand build -trimpath -o (Join-Path $stage 'phi.exe') .
        if ($LASTEXITCODE -ne 0) { throw 'Building phi failed.' }
    } finally {
        Pop-Location
        $env:GOROOT = $oldGoRoot
        $env:GOTOOLCHAIN = $oldToolchain
        $env:CGO_ENABLED = $oldCgo
    }
    if (!(Test-Path $installDir) -and (Test-Path (Join-Path $legacyDir '.phi-install'))) {
        if ((Get-Item $legacyDir).Attributes -band [IO.FileAttributes]::ReparsePoint) { throw 'Refusing to migrate a linked legacy installation.' }
        New-Item -ItemType Directory -Force -Path (Split-Path $installDir) | Out-Null
        Move-Item $legacyDir $installDir
    }
    New-Item -ItemType Directory -Force -Path (Join-Path $installDir 'bin') | Out-Null
    New-Item -ItemType File -Force -Path $marker | Out-Null
    Set-Content -Encoding UTF8 -Path (Join-Path $installDir 'source-dir') -Value $PSScriptRoot
    $sdk = Join-Path $installDir 'go'
    if (Test-Path (Join-Path $stage 'go')) {
        if (Test-Path $sdk) { Remove-Item -Recurse -Force $sdk }
        Move-Item (Join-Path $stage 'go') $sdk
    }
    $binDir = Join-Path $installDir 'bin'
    # Windows can rename a running executable, but cannot overwrite it.
    Get-ChildItem $binDir -Filter 'phi-backup-*.exe' | Remove-Item -Force -ErrorAction SilentlyContinue
    $target = Join-Path $binDir 'phi.exe'
    $backup = Join-Path $binDir ("phi-backup-" + [Guid]::NewGuid().ToString() + '.exe')
    if (Test-Path $target) { Move-Item $target $backup }
    try {
        Move-Item (Join-Path $stage 'phi.exe') $target
    } catch {
        if (Test-Path $backup) { Move-Item $backup $target }
        throw
    }
    if (Test-Path $backup) { Remove-Item -Force $backup -ErrorAction SilentlyContinue }
    $tracking = Join-Path $installDir 'path-added.json'
    $added = @()
    if (Test-Path $tracking) { $added = @(Get-Content $tracking -Raw | ConvertFrom-Json) }
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $entries = @($userPath -split ';' | Where-Object { $_ })
    $legacyPaths = @((Join-Path $legacyDir 'bin'), (Join-Path $legacyDir 'go\bin'))
    $removePaths = @($added | Where-Object { $legacyPaths -contains $_ })
    $entries = @($entries | Where-Object { $removePaths -notcontains $_ })
    $env:Path = ($env:Path -split ';' | Where-Object { $removePaths -notcontains $_ }) -join ';'
    $added = @($added | Where-Object { $removePaths -notcontains $_ })
    $pathDirectories = @((Join-Path $installDir 'bin'))
    if (Test-Path $sdk) { $pathDirectories += Join-Path $sdk 'bin' }
    foreach ($directory in $pathDirectories) {
        if ($entries -notcontains $directory) { $entries += $directory; $added += $directory }
        if (($env:Path -split ';') -notcontains $directory) { $env:Path = "$directory;$env:Path" }
    }
    ConvertTo-Json -InputObject @($added | Select-Object -Unique) | Set-Content $tracking
    [Environment]::SetEnvironmentVariable('Path', ($entries -join ';'), 'User')
    & (Join-Path $installDir 'bin\phi.exe') version
    Write-Host 'Installed phi. Run phi doctor. Restart other terminals to refresh PATH.'
} finally {
    Remove-Item -Recurse -Force $stage
}
