[CmdletBinding()]
param(
    [string] $ResultPath = 'C:\TunnelMint-Test\task014-doh-transport.json'
)

$ErrorActionPreference = 'Stop'

$result = [ordered]@{
    startedAtUtc = [DateTime]::UtcNow.ToString('o')
    administrator = $false
    eventCount = 0
    events = @()
    error = $null
    completedAtUtc = $null
}

try {
    $principal = [Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
    $result.administrator = $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    if (-not $result.administrator) {
        throw 'Run this script from an Administrator PowerShell window.'
    }

    $executable = 'C:\Program Files\TunnelMint\tunnelmint.exe'
    if (-not (Test-Path -LiteralPath $executable -PathType Leaf)) {
        throw 'The installed TunnelMint executable was not found.'
    }

    # /dumplog is held in memory only. Keep exclusively the fixed diagnostic
    # messages, which contain candidate indices, address families, and coarse
    # failure categories; no endpoint, configuration, or key material is kept.
    $pattern = '^.+: \[TUN\] DoH transport (?:dialing bootstrapped candidate \d+ of \d+ \(IPv[46]\)|connected to bootstrapped candidate \d+ of \d+ \(IPv[46]\)|candidate \d+ of \d+ \(IPv[46]\) failed: (?:timeout|canceled|network-error))$'
    $result.events = @(& $executable /dumplog /tail 2>&1 |
        ForEach-Object { [string] $_ } |
        Where-Object { $_ -match $pattern })
    $result.eventCount = $result.events.Count
    if ($result.eventCount -eq 0) {
        throw 'No sanitized DoH transport events were found. Retry DoH activation with the current build, then run this script again.'
    }
}
catch {
    $result.error = $_.Exception.Message
}
finally {
    $result.completedAtUtc = [DateTime]::UtcNow.ToString('o')
    $resultDirectory = Split-Path -Parent $ResultPath
    New-Item -ItemType Directory -Path $resultDirectory -Force | Out-Null
    $result | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $ResultPath -Encoding utf8
}

if ($result.error) {
    Write-Error "Task 014 DoH log extraction failed. Sanitized result: $ResultPath"
    exit 1
}

Write-Host "Task 014 DoH log extraction completed. Sanitized result: $ResultPath"
