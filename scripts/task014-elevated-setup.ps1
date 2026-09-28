[CmdletBinding()]
param(
    [string] $MsiPath = 'C:\Dev\tunnelmint\installer\dist\tunnelmint-amd64-0.1.0.msi',
    [string] $ResultPath = 'C:\TunnelMint-Test\task014-elevated-results.json'
)

$ErrorActionPreference = 'Stop'

function Get-TunnelMintInstallations {
    $uninstallPaths = @(
        'HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*',
        'HKLM:\Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*'
    )

    Get-ItemProperty -Path $uninstallPaths -ErrorAction SilentlyContinue |
        Where-Object { $_.DisplayName -eq 'TunnelMint Development' } |
        ForEach-Object { $_.PSChildName }
}

function Invoke-Msi {
    param(
        [Parameter(Mandatory)] [string[]] $Arguments
    )

    $process = Start-Process -FilePath "$env:SystemRoot\System32\msiexec.exe" `
        -ArgumentList $Arguments -Wait -PassThru
    return $process.ExitCode
}

$result = [ordered]@{
    startedAtUtc = [DateTime]::UtcNow.ToString('o')
    administrator = $false
    msiPath = $MsiPath
    uninstall = [ordered]@{ attempted = $false; productCodes = @(); exitCodes = @() }
    install = [ordered]@{ attempted = $false; exitCode = $null }
    installedUpdate = [ordered]@{ path = 'C:\Program Files\TunnelMint\tunnelmint.exe'; exitCode = $null }
    manager = [ordered]@{ exists = $false; status = $null; startType = $null }
    ui = [ordered]@{ started = $false; running = $false; windowTitle = $null; windowHandle = $null; exitCode = $null }
    error = $null
    completedAtUtc = $null
}

try {
    $principal = [Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
    $result.administrator = $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    if (-not $result.administrator) {
        throw 'Run this script from an Administrator PowerShell window.'
    }

    if (-not (Test-Path -LiteralPath $MsiPath -PathType Leaf)) {
        throw "MSI not found: $MsiPath"
    }

    $productCodes = @(Get-TunnelMintInstallations)
    $result.uninstall.productCodes = $productCodes
    foreach ($productCode in $productCodes) {
        $result.uninstall.attempted = $true
        $exitCode = Invoke-Msi -Arguments @('/x', $productCode, '/qn', '/norestart')
        $result.uninstall.exitCodes += $exitCode
        if ($exitCode -ne 0) {
            throw "TunnelMint uninstall failed with MSI exit code $exitCode."
        }
    }

    $result.install.attempted = $true
    $result.install.exitCode = Invoke-Msi -Arguments @('/i', $MsiPath, '/qn', '/norestart')
    if ($result.install.exitCode -ne 0) {
        throw "TunnelMint install failed with MSI exit code $($result.install.exitCode)."
    }

    $installedClient = $result.installedUpdate.path
    if (-not (Test-Path -LiteralPath $installedClient -PathType Leaf)) {
        throw "Installed TunnelMint client not found: $installedClient"
    }

    $updateProcess = Start-Process -FilePath $installedClient -ArgumentList '/update' -Wait -PassThru
    $result.installedUpdate.exitCode = $updateProcess.ExitCode
    if ($result.installedUpdate.exitCode -ne 0) {
        throw "Installed TunnelMint /update failed with exit code $($result.installedUpdate.exitCode)."
    }

    # The normal launch invokes the elevated manager installer and exits. The
    # manager service then creates the actual UI process, so inspect that
    # process rather than treating the short-lived launcher as the UI.
    Start-Process -FilePath $installedClient | Out-Null
    $result.ui.started = $true
    Start-Sleep -Seconds 12

    $service = Get-Service -Name 'TunnelMintManager' -ErrorAction Stop
    $result.manager.exists = $true
    $result.manager.status = $service.Status.ToString()
    $result.manager.startType = $service.StartType.ToString()

    $uiProcess = Get-Process -Name 'tunnelmint' -ErrorAction SilentlyContinue |
        Where-Object { $_.MainWindowTitle -like 'TunnelMint*' } |
        Select-Object -First 1
    if ($uiProcess) {
        $result.ui.running = $true
        $result.ui.windowTitle = $uiProcess.MainWindowTitle
        $result.ui.windowHandle = [Int64] $uiProcess.MainWindowHandle
    }

    if ($service.Status -ne 'Running') {
        throw "TunnelMintManager service is $($service.Status), not Running."
    }
    if (-not $result.ui.running) {
        throw 'TunnelMint normal startup did not produce a TunnelMint UI window.'
    }
}
catch {
    $result.error = $_.Exception.Message
}
finally {
    $result.completedAtUtc = [DateTime]::UtcNow.ToString('o')
    $resultDirectory = Split-Path -Parent $ResultPath
    New-Item -ItemType Directory -Path $resultDirectory -Force | Out-Null
    $result | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath $ResultPath -Encoding utf8
}

if ($result.error) {
    Write-Error "Task 014 elevated setup failed. Sanitized result: $ResultPath"
    exit 1
}

Write-Host "Task 014 elevated setup passed. Sanitized result: $ResultPath"
