[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidatePattern('^https://')]
    [string] $Endpoint,
    [string] $ResultPath = 'C:\TunnelMint-Test\task014-doh-redirect.json'
)

$ErrorActionPreference = 'Stop'

function Invoke-DoHProbe {
    param(
        [string] $Uri,
        [bool] $FollowRedirects
    )

    $probe = [byte[]] (0, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1)
    $handler = [System.Net.Http.HttpClientHandler]::new()
    $handler.AllowAutoRedirect = $FollowRedirects
    $handler.UseProxy = $false
    $client = [System.Net.Http.HttpClient]::new($handler)
    $request = $null
    $response = $null
    try {
        $request = [System.Net.Http.HttpRequestMessage]::new([System.Net.Http.HttpMethod]::Post, $Uri)
        $request.Content = [System.Net.Http.ByteArrayContent]::new($probe)
        $request.Content.Headers.ContentType = [System.Net.Http.Headers.MediaTypeHeaderValue]::new('application/dns-message')
        [void] $request.Headers.Accept.Add([System.Net.Http.Headers.MediaTypeWithQualityHeaderValue]::new('application/dns-message'))
        $response = $client.SendAsync($request, [System.Net.Http.HttpCompletionOption]::ResponseHeadersRead).GetAwaiter().GetResult()
        $location = $null
        if ($null -ne $response.Headers.Location) {
            $location = $response.Headers.Location.AbsoluteUri
        }
        $contentType = $null
        if ($null -ne $response.Content.Headers.ContentType) {
            $contentType = $response.Content.Headers.ContentType.ToString()
        }
        return [ordered]@{
            requestUri = $Uri
            statusCode = [int] $response.StatusCode
            location = $location
            contentType = $contentType
            error = $null
        }
    }
    catch {
        return [ordered]@{
            requestUri = $Uri
            statusCode = $null
            location = $null
            contentType = $null
            error = $_.Exception.Message
        }
    }
    finally {
        if ($null -ne $response) { $response.Dispose() }
        if ($null -ne $request) { $request.Dispose() }
        $client.Dispose()
        $handler.Dispose()
    }
}

$result = [ordered]@{
    startedAtUtc = [DateTime]::UtcNow.ToString('o')
    endpoint = $Endpoint
    priorDiagnosticBehavior = 'The earlier direct probe followed redirects automatically; its HTTP 200 may be the final response after redirects.'
    automatic = $null
    direct = $null
    manualChain = @()
    error = $null
    completedAtUtc = $null
}

try {
    # Preserve the old probe's behavior for comparison, then disable automatic
    # redirects and follow Location values explicitly with the same POST.
    $result.automatic = Invoke-DoHProbe -Uri $Endpoint -FollowRedirects $true
    $result.direct = Invoke-DoHProbe -Uri $Endpoint -FollowRedirects $false

    $current = $Endpoint
    for ($hop = 0; $hop -lt 6; $hop++) {
        $response = Invoke-DoHProbe -Uri $current -FollowRedirects $false
        $result.manualChain += $response
        if ($response.statusCode -lt 300 -or $response.statusCode -ge 400 -or [string]::IsNullOrWhiteSpace($response.location)) {
            break
        }
        $current = [Uri]::new([Uri] $current, $response.location).AbsoluteUri
    }
}
catch {
    $result.error = $_.Exception.Message
}
finally {
    $result.completedAtUtc = [DateTime]::UtcNow.ToString('o')
    $resultDirectory = Split-Path -Parent $ResultPath
    New-Item -ItemType Directory -Path $resultDirectory -Force | Out-Null
    $result | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $ResultPath -Encoding utf8
}

if ($result.error) {
    Write-Error "Task 014 DoH redirect probe failed. Result: $ResultPath"
    exit 1
}

Write-Host "Task 014 DoH redirect probe completed. Result: $ResultPath"
