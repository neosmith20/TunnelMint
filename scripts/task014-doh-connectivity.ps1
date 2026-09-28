[CmdletBinding()]
param(
    [string] $ResultPath = 'C:\TunnelMint-Test\task014-doh-connectivity.json'
)

$ErrorActionPreference = 'Stop'

function Get-AdapterCounters {
    param([string[]] $Names)

    $counters = @{}
    foreach ($name in $Names) {
        $stats = Get-NetAdapterStatistics -Name $name -ErrorAction Stop
        $counters[$name] = [ordered]@{
            sent = [Int64] $stats.SentBytes
            received = [Int64] $stats.ReceivedBytes
        }
    }
    return $counters
}

function Add-CounterDelta {
    param(
        [hashtable] $Before,
        [hashtable] $After,
        [string[]] $Names,
        [System.Collections.IDictionary] $Destination
    )

    foreach ($name in $Names) {
        $Destination.sent += $After[$name].sent - $Before[$name].sent
        $Destination.received += $After[$name].received - $Before[$name].received
    }
}

$result = [ordered]@{
    startedAtUtc = [DateTime]::UtcNow.ToString('o')
    administrator = $false
    managerRunning = $false
    tunnelServices = [ordered]@{ count = 0; runningCount = 0 }
    tunnelAdapters = [ordered]@{ matchedCount = 0; upCount = 0; routeCount = 0 }
    tcp443 = [ordered]@{ attempted = $false; connected = $false; tunnelAdapterSentBytesDelta = 0; tunnelAdapterReceivedBytesDelta = 0 }
    dohHttps = [ordered]@{ attempted = $false; succeeded = $false; statusCode = $null; failureCategory = $null; tunnelAdapterSentBytesDelta = 0; tunnelAdapterReceivedBytesDelta = 0 }
    error = $null
    completedAtUtc = $null
}

try {
    $principal = [Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
    $result.administrator = $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    if (-not $result.administrator) {
        throw 'Run this script from an Administrator PowerShell window.'
    }

    $result.managerRunning = (Get-Service -Name 'TunnelMintManager' -ErrorAction Stop).Status -eq 'Running'
    $services = @(Get-Service -Name 'TunnelMintTunnel$*' -ErrorAction SilentlyContinue)
    $result.tunnelServices.count = $services.Count
    $result.tunnelServices.runningCount = @($services | Where-Object { $_.Status -eq 'Running' }).Count
    if ($result.tunnelServices.runningCount -eq 0) {
        throw 'No running TunnelMint tunnel service was found. Activate the known-good plain-DNS tunnel first.'
    }

    # Tunnel adapter aliases match tunnel service suffixes. They remain only in
    # memory, so tunnel names and configuration data are never written out.
    $aliases = @($services | ForEach-Object { $_.Name.Substring('TunnelMintTunnel$'.Length) })
    $adapters = @(Get-NetAdapter -Name $aliases -IncludeHidden -ErrorAction Stop)
    $result.tunnelAdapters.matchedCount = $adapters.Count
    $result.tunnelAdapters.upCount = @($adapters | Where-Object { $_.Status -eq 'Up' }).Count
    $result.tunnelAdapters.routeCount = @($adapters | ForEach-Object { @(Get-NetRoute -InterfaceIndex $_.ifIndex -ErrorAction SilentlyContinue) }).Count
    if ($result.tunnelAdapters.upCount -eq 0) {
        throw 'No active TunnelMint tunnel adapter was found.'
    }

    $before = Get-AdapterCounters -Names $aliases
    $result.tcp443.attempted = $true
    try {
        $result.tcp443.connected = [bool] (Test-NetConnection -ComputerName 'cloudflare-dns.com' -Port 443 -InformationLevel Quiet -WarningAction SilentlyContinue)
    }
    catch {
        $result.tcp443.connected = $false
    }
    Start-Sleep -Seconds 2
    $after = Get-AdapterCounters -Names $aliases
    Add-CounterDelta -Before $before -After $after -Names $aliases -Destination $result.tcp443

    $before = Get-AdapterCounters -Names $aliases
    $result.dohHttps.attempted = $true
    try {
        $probe = [byte[]] (0, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1)
        $response = Invoke-WebRequest -Uri 'https://cloudflare-dns.com/dns-query' -Method Post -ContentType 'application/dns-message' -Headers @{ Accept = 'application/dns-message' } -Body $probe -TimeoutSec 10 -UseBasicParsing
        $result.dohHttps.statusCode = [int] $response.StatusCode
        $result.dohHttps.succeeded = $response.StatusCode -ge 200 -and $response.StatusCode -lt 300
        if (-not $result.dohHttps.succeeded) {
            $result.dohHttps.failureCategory = 'http-status'
        }
    }
    catch {
        $message = $_.Exception.Message
        if ($message -match '(?i)timed out|timeout') {
            $result.dohHttps.failureCategory = 'timeout'
        } elseif ($message -match '(?i)certificate|SSL|TLS') {
            $result.dohHttps.failureCategory = 'tls'
        } elseif ($message -match '(?i)name|resolve|DNS') {
            $result.dohHttps.failureCategory = 'name-resolution'
        } else {
            $result.dohHttps.failureCategory = 'other'
        }
    }
    Start-Sleep -Seconds 2
    $after = Get-AdapterCounters -Names $aliases
    Add-CounterDelta -Before $before -After $after -Names $aliases -Destination $result.dohHttps
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
    Write-Error "Task 014 DoH connectivity probe failed. Sanitized result: $ResultPath"
    exit 1
}

Write-Host "Task 014 DoH connectivity probe completed. Sanitized result: $ResultPath"
