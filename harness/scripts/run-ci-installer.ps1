# Windows twin of run-ci-installer.sh: observes the real newsjack.exe doing
# a bare-binary bootstrap install, then runs the same JSON assertion battery
# as the Linux container script. No install script exists on Windows; the
# binary owns the whole flow, so this harness only asserts.
#
# Usage (from repo root, pwsh):
#   harness/scripts/run-ci-installer.ps1 -DistDir .tmp/newsjack-release
#
# Requires: the release dist (with the windows artifacts), python, go.

[CmdletBinding()]
param(
  [Parameter(Mandatory = $true)][string]$DistDir,
  [int]$ServePort = 8765,
  [int]$MockMCPPort = 8970
)

$ErrorActionPreference = 'Stop'
$repoRoot = Resolve-Path (Join-Path $PSScriptRoot '..' '..')
$DistDir = Resolve-Path $DistDir

function Log([string]$message) {
  Write-Host "ci-installer(windows): $message"
}

function Assert([bool]$condition, [string]$message) {
  if (-not $condition) {
    throw "ASSERTION FAILED: $message"
  }
}

function Invoke-CLI {
  # stderr streams through to the job log; only stdout is captured so JSON
  # output stays parseable even when the CLI logs progress to stderr.
  param([string]$Exe, [string[]]$CliArgs)
  $output = & $Exe @CliArgs
  if ($LASTEXITCODE -ne 0) {
    throw "command failed ($LASTEXITCODE): $Exe $($CliArgs -join ' ')"
  }
  return $output
}

function New-ScratchHome([string]$name) {
  $base = if ($env:RUNNER_TEMP) { $env:RUNNER_TEMP } else { [IO.Path]::GetTempPath() }
  $path = Join-Path $base $name
  if (Test-Path $path) { Remove-Item -Recurse -Force $path }
  New-Item -ItemType Directory -Path $path | Out-Null
  return $path
}

function Remove-PathEntries([string]$pattern) {
  return (($env:PATH -split ';') | Where-Object { $_ -and ($_ -notmatch $pattern) }) -join ';'
}

# The battery asserts CLI JSON contracts; these mirror run-ci-installer.sh.
function Test-SetupAndDoctor([string]$cli) {
  $setup = Invoke-CLI $cli @('setup', '--json') | Out-String | ConvertFrom-Json
  foreach ($field in 'monitors_dir', 'agent_prompt', 'recommended_runtime', 'recommended_scheduler', 'agent_command') {
    Assert ($null -ne $setup.$field) "setup --json field $field"
  }
  $doctor = Invoke-CLI $cli @('doctor', '--json') | Out-String | ConvertFrom-Json
  Assert ($doctor.root_ok -eq $true) 'doctor root_ok'
  Assert ($doctor.mcp_bridge.transport -eq 'native') 'doctor mcp_bridge.transport native'
  Assert ($doctor.install.skills_mode -eq 'managed') 'doctor skills_mode managed'
}

function Test-SkillsInstalled([string]$userProfileDir) {
  foreach ($skill in 'newsjack-detector', 'newsjack-monitor-setup') {
    $skillPath = Join-Path $userProfileDir ".claude\skills\$skill\SKILL.md"
    Assert (Test-Path $skillPath) "skill installed: $skillPath"
  }
  $scripts = Join-Path $userProfileDir '.claude\skills\newsjack-detector\scripts'
  Assert (-not (Test-Path $scripts)) 'skill scripts dir must not be installed'
}

function Test-MockDetector([string]$cli) {
  $detector = Invoke-CLI $cli @('detector', 'run', 'AI search visibility', '--mock', '--limit', '1') | Out-String | ConvertFrom-Json
  Assert ($detector.monitor.mock -eq $true) 'detector mock flag'
  Assert ($detector.monitor.queries -contains 'AI search visibility') 'detector query echo'
  Assert (@($detector.signals).Count -ge 1) 'detector returned signals'
}

function Test-MonitorLifecycle([string]$cli, [string]$scratch) {
  $profilePath = Join-Path $scratch 'profile.json'
  @'
{
  "company": "Harness Coffee",
  "description": "Specialty coffee company used for installer verification.",
  "topics": ["coffee supply chain"],
  "search_terms": ["coffee supply chain"],
  "feed_urls": ["https://example.com/feed.xml"],
  "x_news": {"enabled": true},
  "x_trends": {"mode": "none", "woeids": [], "locations": []},
  "standing": ["coffee sourcing"],
  "proof_assets": ["sourcing data"]
}
'@ | Set-Content -Path $profilePath -Encoding utf8

  $init = Invoke-CLI $cli @('monitor', 'init', 'harness-coffee', '--profile', $profilePath) | Out-String | ConvertFrom-Json
  Assert ($init.slug -eq 'harness-coffee') 'monitor init slug'
  Assert ($null -ne $init.profile_path) 'monitor init profile_path'

  $test = Invoke-CLI $cli @('monitor', 'test', 'harness-coffee', '--mock', '--limit', '2') | Out-String | ConvertFrom-Json
  foreach ($field in 'candidates', 'summary', 'report_target') {
    Assert ($null -ne $test.$field) "monitor test field $field"
  }
  Assert (Test-Path $test.candidates) 'monitor test candidates artifact'
  Assert (Test-Path $test.summary) 'monitor test summary artifact'
  Assert (-not (Test-Path $test.report_target)) 'monitor test must not write the report'

  $schedule = Invoke-CLI $cli @('monitor', 'schedule', 'harness-coffee', '--runtime', 'claude', '--every', '1h') | Out-String | ConvertFrom-Json
  Assert ($schedule.system_cron -eq $false) 'monitor schedule avoids system cron'
  Assert ($schedule.runtime -eq 'claude') 'monitor schedule runtime'
  Assert ($schedule.suggested_minute -ge 1 -and $schedule.suggested_minute -le 59) 'monitor schedule minute range'
  $scheduleBody = Get-Content -Raw $schedule.schedule_path
  Assert ($scheduleBody -notmatch 'crontab|launchd|systemd') 'schedule markdown free of system schedulers'
  Assert ($scheduleBody -match 'never minute 0') 'schedule markdown keeps the never-minute-0 rule'

  $status = Invoke-CLI $cli @('monitor', 'status', 'harness-coffee') | Out-String | ConvertFrom-Json
  Assert ($status.exists -eq $true) 'monitor status exists'
  Assert ($status.run_count -eq 1) 'monitor status run_count'
}

function Test-BridgeSmokeWithoutNode([string]$cli) {
  $mock = Start-Process -FilePath 'go' -ArgumentList @('run', '.', '--addr', "127.0.0.1:$MockMCPPort", '--key', 'mock-key') `
    -WorkingDirectory (Join-Path $repoRoot 'harness\mock-mcp') -PassThru -NoNewWindow
  try {
    $ready = $false
    for ($i = 0; $i -lt 100; $i++) {
      try {
        $response = Invoke-WebRequest -Uri "http://127.0.0.1:$MockMCPPort/mcp" -Method Get -SkipHttpErrorCheck -TimeoutSec 2
        if ($response.StatusCode -gt 0) { $ready = $true; break }
      } catch {
        Start-Sleep -Milliseconds 300
      }
    }
    Assert $ready 'mock MCP server became ready'

    $savedPath = $env:PATH
    $savedKey = $env:MEDIALYST_API_KEY
    $savedURL = $env:NEWSJACK_MEDIALYST_MCP_URL
    try {
      # Prove the bridge needs no Node: strip node/npm dirs from PATH.
      $env:PATH = Remove-PathEntries 'node|npm'
      $env:MEDIALYST_API_KEY = 'mock-key'
      $env:NEWSJACK_MEDIALYST_MCP_URL = "http://127.0.0.1:$MockMCPPort/mcp"
      $messages = @(
        '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"harness-smoke","version":"0"}}}'
        '{"jsonrpc":"2.0","method":"notifications/initialized"}'
        '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'
      )
      $bridgeOut = $messages | & $cli mcp-bridge
      if ($LASTEXITCODE -ne 0) { throw "mcp-bridge exited $LASTEXITCODE" }
      $joined = $bridgeOut -join "`n"
      Assert ($joined -match 'protocolVersion') 'bridge relayed the initialize result'
      Assert ($joined -match 'mock_search') 'bridge relayed the SSE tool list'
    } finally {
      $env:PATH = $savedPath
      $env:MEDIALYST_API_KEY = $savedKey
      $env:NEWSJACK_MEDIALYST_MCP_URL = $savedURL
    }
  } finally {
    if (-not $mock.HasExited) { Stop-Process -Id $mock.Id -Force }
  }
}

function Test-AutoUpdateSwapsRunningExe([string]$cli, [string]$newsjackHome) {
  # Stale the recorded version, run a user-facing command, and assert the
  # binary updated itself in place: the one scenario only a real Windows
  # machine can prove (a running exe swapping itself via the rename dance).
  $stateFile = Join-Path $newsjackHome 'install.json'
  $state = Get-Content -Raw $stateFile | ConvertFrom-Json
  $liveVersion = $state.version
  $state.version = 'v0.0.0-stale'
  $state | ConvertTo-Json -Depth 8 | Set-Content -Path $stateFile -Encoding utf8
  Set-Content -Path (Join-Path $newsjackHome 'newsjack\VERSION') -Value 'v0.0.0-stale' -Encoding utf8

  $savedNoUpdate = $env:NEWSJACK_NO_AUTO_UPDATE
  try {
    $env:NEWSJACK_NO_AUTO_UPDATE = ''
    $doctor = Invoke-CLI $cli @('doctor', '--json') | Out-String | ConvertFrom-Json
    Assert ($doctor.root_ok -eq $true) 'doctor stdout stayed valid JSON through auto-update'
  } finally {
    $env:NEWSJACK_NO_AUTO_UPDATE = $savedNoUpdate
  }

  $restored = (Get-Content -Raw (Join-Path $newsjackHome 'newsjack\VERSION')).Trim()
  Assert ($restored -eq $liveVersion) "auto-update restored VERSION ($restored vs $liveVersion)"
  $parked = "$cli.old"
  Assert (Test-Path $parked) 'auto-update parked the previous exe at .old'

  # The parked exe unlocks once the pre-update process exits; the next run
  # cleans it up. Allow a few attempts for the file handle to release.
  $cleaned = $false
  for ($i = 0; $i -lt 10; $i++) {
    Invoke-CLI $cli @('version') | Out-Null
    if (-not (Test-Path $parked)) { $cleaned = $true; break }
    Start-Sleep -Milliseconds 500
  }
  Assert $cleaned 'next run removed the parked .old binary'
}

function Invoke-BootstrapLeg {
  param(
    [string]$LegName,
    [string]$NewsjackHome,
    [switch]$StripGit,
    [switch]$FullBattery
  )
  Log "leg: $LegName (NEWSJACK_HOME=$NewsjackHome)"
  $scratch = New-ScratchHome "newsjack-$LegName-scratch"
  $bareExe = Join-Path $scratch 'newsjack.exe'
  Copy-Item (Join-Path $DistDir 'newsjack_windows_amd64.exe') $bareExe

  $saved = @{
    PATH                    = $env:PATH
    NEWSJACK_HOME           = $env:NEWSJACK_HOME
    NEWSJACK_RELEASE_BASE   = $env:NEWSJACK_RELEASE_BASE
    NEWSJACK_RUNTIMES       = $env:NEWSJACK_RUNTIMES
    NEWSJACK_INSTALL_MCP    = $env:NEWSJACK_INSTALL_MCP
    NEWSJACK_NO_AUTO_UPDATE = $env:NEWSJACK_NO_AUTO_UPDATE
    NEWSJACK_NO_PATH_UPDATE = $env:NEWSJACK_NO_PATH_UPDATE
    USERPROFILE             = $env:USERPROFILE
    HOME                    = $env:HOME
  }
  # Run from the scratch dir, not the repo checkout: inside the repo,
  # newsjackRoot() resolves to the source tree and setup never bootstraps
  # (the same reason the Linux battery does `cd /tmp` before its checks).
  Push-Location $scratch
  try {
    if ($StripGit) { $env:PATH = Remove-PathEntries 'git' }
    $env:NEWSJACK_HOME = $NewsjackHome
    $env:NEWSJACK_RELEASE_BASE = "http://127.0.0.1:$ServePort"
    $env:NEWSJACK_RUNTIMES = 'claude'
    $env:NEWSJACK_INSTALL_MCP = '0'
    $env:NEWSJACK_NO_AUTO_UPDATE = '1'
    # Keep the runner's HKCU PATH untouched.
    $env:NEWSJACK_NO_PATH_UPDATE = '1'
    # The Go CLI's homeDir() prefers HOME over USERPROFILE; set both so the
    # leg stays isolated even on runners that define HOME.
    $env:USERPROFILE = $scratch
    $env:HOME = $scratch

    Invoke-CLI $bareExe @('setup', '--json') | Out-Null
    $installedCli = Join-Path $NewsjackHome 'bin\newsjack.exe'
    Assert (Test-Path $installedCli) 'bootstrap installed the managed CLI binary'
    Assert (Test-Path (Join-Path $NewsjackHome 'newsjack\VERSION')) 'bootstrap installed the bundle'
    Assert (Test-Path (Join-Path $NewsjackHome 'install.json')) 'bootstrap wrote install state'

    $version = Invoke-CLI $installedCli @('version') | Out-String
    Log "installed version: $($version.Trim())"

    Test-SetupAndDoctor $installedCli
    Test-SkillsInstalled $scratch
    if ($FullBattery) {
      Test-MockDetector $installedCli
      Test-MonitorLifecycle $installedCli $scratch
      Test-BridgeSmokeWithoutNode $installedCli
      Test-AutoUpdateSwapsRunningExe $installedCli $NewsjackHome
    }
    Log "leg passed: $LegName"
  } finally {
    Pop-Location
    foreach ($key in $saved.Keys) {
      Set-Item -Path "env:$key" -Value $saved[$key]
    }
  }
}

Log "serving release dist from $DistDir on port $ServePort"
$server = Start-Process -FilePath 'python' -ArgumentList @('-m', 'http.server', "$ServePort", '--directory', "$DistDir") -PassThru -NoNewWindow
try {
  $ready = $false
  for ($i = 0; $i -lt 50; $i++) {
    try {
      Invoke-WebRequest -Uri "http://127.0.0.1:$ServePort/checksums.txt" -TimeoutSec 2 | Out-Null
      $ready = $true
      break
    } catch {
      Start-Sleep -Milliseconds 300
    }
  }
  Assert $ready 'release dist server became ready'

  # Leg 1: full battery in a plain temp home.
  Invoke-BootstrapLeg -LegName 'managed' -NewsjackHome (Join-Path (New-ScratchHome 'newsjack-managed-home') '.newsjack') -FullBattery

  # Leg 2: no git on PATH — install/setup must not silently shell out.
  Invoke-BootstrapLeg -LegName 'no-git' -NewsjackHome (Join-Path (New-ScratchHome 'newsjack-nogit-home') '.newsjack') -StripGit

  # Leg 3: spaces in the profile path — the classic Windows installer bug.
  $spacedBase = New-ScratchHome 'newsjack spaced home'
  Invoke-BootstrapLeg -LegName 'spaced-profile' -NewsjackHome (Join-Path $spacedBase 'jane smith\.newsjack')

  Log 'windows installer battery complete'
} finally {
  if (-not $server.HasExited) { Stop-Process -Id $server.Id -Force }
}
