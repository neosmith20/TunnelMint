[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidatePattern('^https://')]
    [string] $Endpoint,
    [string] $ResultPath = 'C:\TunnelMint-Test\task014-doh-leak-acceptance.json'
    ,
    [switch] $ManualServerBreak
)

$ErrorActionPreference = 'Stop'
$captureActive = $false
$firewallRuleName = "TunnelMint Task014 DoH break $([Guid]::NewGuid().ToString('N'))"
$captureDirectory = 'C:\TunnelMint-Test'

function Get-EndpointFingerprint {
    param([string] $Value)

    $sha = [Security.Cryptography.SHA256]::Create()
    try {
        $bytes = [Text.Encoding]::UTF8.GetBytes($Value)
        return ([BitConverter]::ToString($sha.ComputeHash($bytes))).Replace('-', '').ToLowerInvariant()
    }
    finally {
        $sha.Dispose()
    }
}

function Get-TunnelContext {
    $services = @(Get-Service -Name 'TunnelMintTunnel$*' -ErrorAction SilentlyContinue | Where-Object { $_.Status -eq 'Running' })
    if ($services.Count -ne 1) {
        throw "Expected exactly one running TunnelMint tunnel service; found $($services.Count)."
    }
    $service = $services[0]
    $cim = Get-CimInstance Win32_Service -Filter "Name = '$($service.Name)'" -ErrorAction Stop
    if (-not $cim.ProcessId) {
        throw 'The running TunnelMint tunnel service has no process ID.'
    }
    $alias = $service.Name.Substring('TunnelMintTunnel$'.Length)
    $adapter = Get-NetAdapter -Name $alias -IncludeHidden -ErrorAction Stop
    if ($adapter.Status -ne 'Up') {
        throw 'The TunnelMint tunnel adapter is not Up.'
    }
    return [ordered]@{
        serviceName = $service.Name
        processId = [int] $cim.ProcessId
        adapterAlias = $alias
        interfaceIndex = [int] $adapter.ifIndex
    }
}

function Get-DnsLoopbackSummary {
    param([int] $InterfaceIndex)

    $addresses = @(
        Get-DnsClientServerAddress -InterfaceIndex $InterfaceIndex -ErrorAction Stop |
            ForEach-Object { $_.ServerAddresses } |
            Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
    )
    $nonLoopback = @($addresses | Where-Object { $_ -notin @('127.0.0.1', '::1') })
    return [ordered]@{
        configuredAddressCount = $addresses.Count
        loopbackAddressCount = @($addresses | Where-Object { $_ -in @('127.0.0.1', '::1') }).Count
        nonLoopbackAddressCount = $nonLoopback.Count
        loopbackOnly = $addresses.Count -gt 0 -and $nonLoopback.Count -eq 0
    }
}

function Get-SystemDnsSummary {
    $adapters = @(Get-NetAdapter -IncludeHidden -ErrorAction Stop | Where-Object { $_.Status -eq 'Up' })
    $addresses = @(
        foreach ($adapter in $adapters) {
            Get-DnsClientServerAddress -InterfaceIndex $adapter.ifIndex -ErrorAction SilentlyContinue |
                ForEach-Object { $_.ServerAddresses } |
                Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
        }
    )
    $nonLoopback = @($addresses | Where-Object { $_ -notin @('127.0.0.1', '::1') })
    return [ordered]@{
        upAdapterCount = $adapters.Count
        configuredAddressCount = $addresses.Count
        loopbackAddressCount = @($addresses | Where-Object { $_ -in @('127.0.0.1', '::1') }).Count
        nonLoopbackAddressCount = $nonLoopback.Count
        loopbackOnly = $addresses.Count -gt 0 -and $nonLoopback.Count -eq 0
    }
}

function Get-EndpointAddresses {
    param([Uri] $Uri)

    $records = @()
    foreach ($type in @('A', 'AAAA')) {
        try {
            $records += @(Resolve-DnsName -Name $Uri.DnsSafeHost -Type $type -DnsOnly -ErrorAction Stop)
        }
        catch {
            # A hostname may legitimately have only one address family.
        }
    }
    return @($records | Where-Object { $_.IPAddress } | ForEach-Object { $_.IPAddress } | Sort-Object -Unique)
}

function Get-ConnectionSummary {
    param(
        [int] $ProcessId,
        [string[]] $EndpointAddresses
    )

    $connections = @(Get-NetTCPConnection -State Established -OwningProcess $ProcessId -RemotePort 443 -ErrorAction SilentlyContinue)
    $endpointConnections = @($connections | Where-Object { $EndpointAddresses -contains $_.RemoteAddress })
    return [ordered]@{
        established443Count = $connections.Count
        endpoint443Count = $endpointConnections.Count
    }
}

function Start-Port53Capture {
    param([string] $EtlPath)

    & pktmon filter remove *> $null
    & pktmon filter add 'TunnelMintTask014Port53' -p 53 *> $null
    & pktmon start --capture --file-name $EtlPath *> $null
    $script:captureActive = $true
}

function Stop-Port53Capture {
    param([string] $EtlPath)

    if ($script:captureActive) {
        & pktmon stop *> $null
        $script:captureActive = $false
    }
    & pktmon filter remove *> $null
    $textPath = [IO.Path]::ChangeExtension($EtlPath, '.txt')
    & pktmon etl2txt $EtlPath --out $textPath --brief *> $null
    return $textPath
}

function Get-Port53Summary {
    param([string] $TextPath)

    $lines = @(Get-Content -LiteralPath $TextPath -ErrorAction Stop)
    $events = @()
    for ($index = 0; $index -lt $lines.Count; $index++) {
        $line = [string] $lines[$index]
        if ($line -notmatch '(?i)\b(?:UDP|TCP)\b' -or
            $line -notmatch '(?i)\bPort\s+(?:Dest|Src)\s+53\b|\b(?:Dest|Src)\s+Port\s+53\b') {
            continue
        }

        # Current pktmon text renders each packet on one line. Keep a small
        # context fallback for older builds that split the packet fields over
        # adjacent lines, but never mix adjacent complete packet records.
        $context = $line
        $addresses = @(
            [regex]::Matches($context, '(?i)(?:IP\s+(?:Src|Source|Dst|Dest|Destination)|(?:Src|Source|Dst|Dest|Destination)\s+IP)\s*[:= ]\s*([0-9a-f:.]+)') |
                ForEach-Object { $_.Groups[1].Value }
        )
        if ($addresses.Count -eq 0) {
            $start = [Math]::Max(0, $index - 8)
            $end = [Math]::Min($lines.Count - 1, $index + 8)
            $context = [string]::Join("`n", @($lines[$start..$end]))
            $addresses = @(
                [regex]::Matches($context, '(?i)(?:IP\s+(?:Src|Source|Dst|Dest|Destination)|(?:Src|Source|Dst|Dest|Destination)\s+IP)\s*[:= ]\s*([0-9a-f:.]+)') |
                    ForEach-Object { $_.Groups[1].Value }
            )
        }
        $nonLoopbackAddresses = @($addresses | Where-Object { $_ -notin @('127.0.0.1', '::1') })
        $isNonLoopback = if ($addresses.Count -gt 0) {
            $nonLoopbackAddresses.Count -gt 0
        }
        else {
            # If this Windows build omits labelled addresses, retain the
            # conservative legacy classification and report the event as
            # unclassified for the wire-level acceptance decision.
            $context -notmatch '(?i)127\.0\.0\.1.*127\.0\.0\.1|::1.*::1'
        }
        $isTx = $context -match '(?i)(?:Direction|Dir)\s*[:= ]\s*(?:Tx|Transmit|Outbound)\b'
        $isWire = $context -match '(?i)(?:Type|Component)\s*[:= ]\s*(?:Ethernet|NDIS|NetBufferList|NIC)\b'
        $isDropped = $context -match '(?i)\b(?:drop|dropped|blocked)\b'
        $events += [pscustomobject]@{
            nonLoopback = [bool] $isNonLoopback
            classifiedAddresses = $addresses.Count -gt 0
            tx = [bool] $isTx
            wire = [bool] $isWire
            dropped = [bool] $isDropped
        }
    }
    $nonLoopback = @($events | Where-Object { $_.nonLoopback })
    $wireTx = @($events | Where-Object { $_.nonLoopback -and $_.tx -and $_.wire -and -not $_.dropped })
    $unclassified = @($events | Where-Object { -not $_.classifiedAddresses -or -not $_.tx -or -not $_.wire })
    return [ordered]@{
        port53EventCount = $events.Count
        loopbackOnlyEventCount = $events.Count - $nonLoopback.Count
        nonLoopbackEventCount = $nonLoopback.Count
        nonLoopbackTrafficObserved = $nonLoopback.Count -gt 0
        outboundWireEventCount = $wireTx.Count
        unclassifiedEventCount = $unclassified.Count
        # Zero matching packet records is valid evidence when pktmon produced
        # a readable capture; there is then nothing to classify.
        wireEvidenceComplete = $unclassified.Count -eq 0
    }
}

function Invoke-FreshDnsQuery {
    param(
        [int] $ProcessId,
        [string[]] $EndpointAddresses,
        [int] $TimeoutSeconds = 20
    )

    $queryName = "tm014-$([Guid]::NewGuid().ToString('N')).example.com"
    $job = Start-Job -ScriptBlock {
        param([string] $Name)
        try {
            Resolve-DnsName -Name $Name -Type A -DnsOnly -ErrorAction Stop | Out-Null
            'response'
        }
        catch {
            'error'
        }
    } -ArgumentList $queryName

    $samples = 0
    $maxEstablished443 = 0
    $maxEndpoint443 = 0
    $deadline = [DateTime]::UtcNow.AddSeconds($TimeoutSeconds)
    try {
        while ($job.State -eq 'Running' -and [DateTime]::UtcNow -lt $deadline) {
            $summary = Get-ConnectionSummary -ProcessId $ProcessId -EndpointAddresses $EndpointAddresses
            $samples++
            $maxEstablished443 = [Math]::Max($maxEstablished443, $summary.established443Count)
            $maxEndpoint443 = [Math]::Max($maxEndpoint443, $summary.endpoint443Count)
            Start-Sleep -Milliseconds 250
        }
        if ($job.State -eq 'Running') {
            Stop-Job -Job $job -ErrorAction SilentlyContinue
            return [ordered]@{
                queryCompleted = $false
                queryResponseObserved = $false
                timedOut = $true
                connectionSamples = $samples
                maxEstablished443Count = $maxEstablished443
                maxEndpoint443Count = $maxEndpoint443
            }
        }
        $outcome = [string] (Receive-Job -Job $job -ErrorAction SilentlyContinue | Select-Object -First 1)
        $summary = Get-ConnectionSummary -ProcessId $ProcessId -EndpointAddresses $EndpointAddresses
        $maxEstablished443 = [Math]::Max($maxEstablished443, $summary.established443Count)
        $maxEndpoint443 = [Math]::Max($maxEndpoint443, $summary.endpoint443Count)
        return [ordered]@{
            queryCompleted = $true
            queryResponseObserved = $outcome -eq 'response'
            timedOut = $false
            connectionSamples = $samples
            maxEstablished443Count = $maxEstablished443
            maxEndpoint443Count = $maxEndpoint443
        }
    }
    finally {
        Remove-Job -Job $job -Force -ErrorAction SilentlyContinue
    }
}

$result = [ordered]@{
    startedAtUtc = [DateTime]::UtcNow.ToString('o')
    administrator = $false
    endpointFingerprintSha256 = Get-EndpointFingerprint $Endpoint
    endpointAddressCount = 0
    tunnel = $null
    dnsLoopback = $null
    baseline = $null
    broken = $null
    systemDns = $null
    breakRule = [ordered]@{ mode = $null; attempted = $false; created = $false; removed = $false }
    serverLogRequiredForExactPathProof = $true
    error = $null
    completedAtUtc = $null
}

try {
    $principal = [Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
    $result.administrator = $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    if (-not $result.administrator) {
        throw 'Run this script from an Administrator PowerShell window.'
    }
    $uri = [Uri]::new($Endpoint)
    $context = Get-TunnelContext
    $result.tunnel = [ordered]@{ running = $true; adapterUp = $true }
    $result.dnsLoopback = Get-DnsLoopbackSummary -InterfaceIndex $context.interfaceIndex
    $result.systemDns = Get-SystemDnsSummary
    if (-not $result.dnsLoopback.loopbackOnly) {
        throw 'The TunnelMint adapter DNS configuration is not loopback-only.'
    }

    $endpointAddresses = @(Get-EndpointAddresses -Uri $uri)
    $result.endpointAddressCount = $endpointAddresses.Count
    if ($endpointAddresses.Count -eq 0) {
        throw 'The configured DoH hostname returned no addresses.'
    }

    $baselineEtl = Join-Path $captureDirectory 'task014-doh-baseline.etl'
    Start-Port53Capture -EtlPath $baselineEtl
    try {
        $query = Invoke-FreshDnsQuery -ProcessId $context.processId -EndpointAddresses $endpointAddresses
    }
    finally {
        $baselineText = Stop-Port53Capture -EtlPath $baselineEtl
    }
    $result.baseline = [ordered]@{
        query = $query
        port53 = Get-Port53Summary -TextPath $baselineText
    }

    if ($ManualServerBreak) {
        $result.breakRule.mode = 'manual-server-break'
        Write-Host 'Disable or revoke the configured DoH endpoint on its server, then press Enter to run the broken-endpoint query.'
        [void] (Read-Host)
    }
    else {
        $result.breakRule.mode = 'local-firewall-block'
        $result.breakRule.attempted = $true
        New-NetFirewallRule -DisplayName $firewallRuleName -Direction Outbound -Action Block -Protocol TCP -RemotePort 443 -RemoteAddress $endpointAddresses -Profile Any -ErrorAction Stop | Out-Null
        $result.breakRule.created = $true
    }

    $brokenEtl = Join-Path $captureDirectory 'task014-doh-broken.etl'
    Start-Port53Capture -EtlPath $brokenEtl
    try {
        $query = Invoke-FreshDnsQuery -ProcessId $context.processId -EndpointAddresses $endpointAddresses
    }
    finally {
        $brokenText = Stop-Port53Capture -EtlPath $brokenEtl
    }
    $result.broken = [ordered]@{
        query = $query
        port53 = Get-Port53Summary -TextPath $brokenText
    }

    $failedConditions = @()
    if (-not $result.baseline.query.queryResponseObserved) {
        $failedConditions += 'baseline DNS query did not receive a response'
    }
    if (-not $result.baseline.port53.wireEvidenceComplete -or $result.baseline.port53.outboundWireEventCount -gt 0) {
        $failedConditions += 'baseline capture did not prove zero outbound plaintext DNS'
    }
    if ($result.broken.query.queryResponseObserved) {
        $failedConditions += 'DNS still received a response after the endpoint break'
    }
    if (-not $result.broken.port53.wireEvidenceComplete -or $result.broken.port53.outboundWireEventCount -gt 0) {
        $failedConditions += 'broken-endpoint capture did not prove zero outbound plaintext DNS'
    }
    if ($failedConditions.Count -gt 0) {
        throw ('Acceptance conditions not met: ' + ($failedConditions -join '; '))
    }
}
catch {
    $result.error = $_.Exception.Message
}
finally {
    if ($captureActive) {
        & pktmon stop *> $null
        $captureActive = $false
    }
    & pktmon filter remove *> $null
    if ($result.breakRule.created) {
        Remove-NetFirewallRule -DisplayName $firewallRuleName -ErrorAction SilentlyContinue
        $result.breakRule.removed = $true
    }
    foreach ($path in @(
        (Join-Path $captureDirectory 'task014-doh-baseline.etl'),
        (Join-Path $captureDirectory 'task014-doh-baseline.txt'),
        (Join-Path $captureDirectory 'task014-doh-broken.etl'),
        (Join-Path $captureDirectory 'task014-doh-broken.txt')
    )) {
        if (Test-Path -LiteralPath $path) { Remove-Item -LiteralPath $path -Force -ErrorAction SilentlyContinue }
    }
    $result.completedAtUtc = [DateTime]::UtcNow.ToString('o')
    New-Item -ItemType Directory -Path (Split-Path -Parent $ResultPath) -Force | Out-Null
    $result | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $ResultPath -Encoding utf8
}

if ($result.error) {
    Write-Error "Task 014 DoH leak acceptance failed. Sanitized result: $ResultPath"
    exit 1
}

Write-Host "Task 014 DoH leak acceptance completed. Sanitized result: $ResultPath"
