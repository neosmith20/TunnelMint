[CmdletBinding()]
param(
    [ValidateSet('Upgrade', 'UninstallReinstall')]
    [string] $Phase = 'Upgrade',
    [string] $MsiPath = 'C:\Dev\tunnelmint\installer\dist\wirehush-amd64-0.1.0.msi',
    [string] $ResultPath = 'C:\TunnelMint-Test\task018-release-candidate.json'
)

$ErrorActionPreference = 'Stop'

function Test-Administrator {
    $principal = [Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Get-WireHushInstallations {
    $paths = @(
        'HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*',
        'HKLM:\Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*'
    )
    return @(
        Get-ItemProperty -Path $paths -ErrorAction SilentlyContinue |
            Where-Object { $_.DisplayName -in @('TunnelMint Development', 'WireHush Development') } |
            Select-Object DisplayName, DisplayVersion, Publisher, PSChildName
    )
}

function Get-WireGuardMarker {
    $path = 'C:\Program Files\WireGuard\wireguard.exe'
    $file = Get-Item -LiteralPath $path -ErrorAction SilentlyContinue
    $service = Get-Service -Name 'WireGuardManager' -ErrorAction SilentlyContinue
    return [ordered]@{
        executablePresent = $null -ne $file
        executableLength = if ($file) { [Int64] $file.Length } else { $null }
        executableVersion = if ($file) { $file.VersionInfo.FileVersion } else { $null }
        managerServicePresent = $null -ne $service
        managerServiceStatus = if ($service) { $service.Status.ToString() } else { $null }
    }
}

function Invoke-Msi {
    param([string[]] $Arguments)

    $process = Start-Process -FilePath 'msiexec.exe' -ArgumentList $Arguments -Wait -PassThru
    return $process.ExitCode
}

$result = [ordered]@{
    startedAtUtc = [DateTime]::UtcNow.ToString('o')
    phase = $Phase
    administrator = $false
    wireGuardBefore = $null
    wireGuardAfter = $null
    productBefore = [ordered]@{ count = 0; names = @() }
    productAfter = [ordered]@{ count = 0; names = @() }
    upgrade = [ordered]@{ attempted = $false; exitCode = $null }
    uninstall = [ordered]@{ attempted = $false; exitCode = $null }
    reinstall = [ordered]@{ attempted = $false; exitCode = $null }
    wireHush = [ordered]@{
        executablePresent = $false
        updateExitCode = $null
        managerServicePresent = $false
        managerDisplayName = $null
        managerStatus = $null
        uiWindowPresent = $false
        startMenuShortcutPresent = $false
    }
    error = $null
    completedAtUtc = $null
}

try {
    $result.administrator = Test-Administrator
    if (-not $result.administrator) {
        throw 'Run this script from an Administrator PowerShell window.'
    }
    if (-not (Test-Path -LiteralPath $MsiPath -PathType Leaf)) {
        throw "MSI not found: $MsiPath"
    }

    $before = @(Get-WireHushInstallations)
    $result.productBefore.count = $before.Count
    $result.productBefore.names = @($before | ForEach-Object { $_.DisplayName })
    $result.wireGuardBefore = Get-WireGuardMarker

    if ($Phase -eq 'Upgrade') {
        $result.upgrade.attempted = $true
        $result.upgrade.exitCode = Invoke-Msi -Arguments @('/i', $MsiPath, '/qn', '/norestart')
        if ($result.upgrade.exitCode -ne 0) {
            throw "WireHush upgrade failed with MSI exit code $($result.upgrade.exitCode)."
        }
    }
    else {
        $installed = @((Get-WireHushInstallations) | Where-Object { $_.DisplayName -eq 'WireHush Development' })
        if ($installed.Count -ne 1) {
            throw "Expected one installed WireHush product before uninstall; found $($installed.Count)."
        }
        $result.uninstall.attempted = $true
        $result.uninstall.exitCode = Invoke-Msi -Arguments @('/x', $installed[0].PSChildName, '/qn', '/norestart')
        if ($result.uninstall.exitCode -ne 0) {
            throw "WireHush uninstall failed with MSI exit code $($result.uninstall.exitCode)."
        }
        if (Test-Path -LiteralPath 'C:\Program Files\TunnelMint\wirehush.exe' -PathType Leaf) {
            throw 'WireHush executable remains after uninstall.'
        }
        $result.reinstall.attempted = $true
        $result.reinstall.exitCode = Invoke-Msi -Arguments @('/i', $MsiPath, '/qn', '/norestart')
        if ($result.reinstall.exitCode -ne 0) {
            throw "WireHush reinstall failed with MSI exit code $($result.reinstall.exitCode)."
        }
    }

    $executable = 'C:\Program Files\TunnelMint\wirehush.exe'
    $result.wireHush.executablePresent = Test-Path -LiteralPath $executable -PathType Leaf
    if (-not $result.wireHush.executablePresent) {
        throw "Installed WireHush executable was not found: $executable"
    }
    $update = Start-Process -FilePath $executable -ArgumentList '/update' -Wait -PassThru
    $result.wireHush.updateExitCode = $update.ExitCode
    if ($update.ExitCode -ne 0) {
        throw "Installed WireHush /update failed with exit code $($update.ExitCode)."
    }

    Start-Process -FilePath $executable | Out-Null
    Start-Sleep -Seconds 12
    $manager = Get-Service -Name 'TunnelMintManager' -ErrorAction Stop
    $result.wireHush.managerServicePresent = $true
    $result.wireHush.managerDisplayName = $manager.DisplayName
    $result.wireHush.managerStatus = $manager.Status.ToString()
    if ($manager.DisplayName -ne 'WireHush Manager' -or $manager.Status -ne 'Running') {
        throw 'WireHush manager service is not running with its expected display name.'
    }
    $ui = Get-Process -Name 'wirehush' -ErrorAction SilentlyContinue |
        Where-Object { $_.MainWindowTitle -like 'WireHush*' } |
        Select-Object -First 1
    $result.wireHush.uiWindowPresent = $null -ne $ui
    if (-not $result.wireHush.uiWindowPresent) {
        throw 'WireHush UI window was not found after normal launch.'
    }
    $result.wireHush.startMenuShortcutPresent = Test-Path -LiteralPath (Join-Path $env:ProgramData 'Microsoft\Windows\Start Menu\Programs\WireHush.lnk') -PathType Leaf
    if (-not $result.wireHush.startMenuShortcutPresent) {
        throw 'WireHush Start Menu shortcut was not found.'
    }

    $after = @(Get-WireHushInstallations)
    $result.productAfter.count = $after.Count
    $result.productAfter.names = @($after | ForEach-Object { $_.DisplayName })
    if ($result.productAfter.names -notcontains 'WireHush Development') {
        throw 'Installed-program registry does not identify WireHush Development.'
    }
    $result.wireGuardAfter = Get-WireGuardMarker
    if (-not $result.wireGuardAfter.executablePresent -or
        $result.wireGuardAfter.executableLength -ne $result.wireGuardBefore.executableLength -or
        $result.wireGuardAfter.executableVersion -ne $result.wireGuardBefore.executableVersion) {
        throw 'Official WireGuard executable changed or is unavailable after WireHush operation.'
    }
}
catch {
    $result.error = $_.Exception.Message
}
finally {
    $result.completedAtUtc = [DateTime]::UtcNow.ToString('o')
    New-Item -ItemType Directory -Path (Split-Path -Parent $ResultPath) -Force | Out-Null
    $result | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $ResultPath -Encoding utf8
}

if ($result.error) {
    Write-Error "Task 018 $Phase check failed. Sanitized result: $ResultPath"
    exit 1
}

Write-Host "Task 018 $Phase check completed. Sanitized result: $ResultPath"