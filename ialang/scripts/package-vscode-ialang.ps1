param(
    [string]$ExtensionDir = "tools/vscode-ialang",
    [string]$OutFile = "",
    [switch]$InstallDeps
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Invoke-Step {
    param(
        [Parameter(Mandatory = $true)][string]$Message,
        [Parameter(Mandatory = $true)][scriptblock]$Action
    )
    Write-Host "==> $Message"
    & $Action
}

$repoRoot = Split-Path -Parent $PSScriptRoot
$extPath = Join-Path $repoRoot $ExtensionDir
$packageJsonPath = Join-Path $extPath "package.json"

if (!(Test-Path -LiteralPath $extPath)) {
    throw "Extension directory not found: $extPath"
}
if (!(Test-Path -LiteralPath $packageJsonPath)) {
    throw "package.json not found: $packageJsonPath"
}

Push-Location $extPath
try {
    $nodeModulesPath = Join-Path $extPath "node_modules"
    if ($InstallDeps -or !(Test-Path -LiteralPath $nodeModulesPath)) {
        Invoke-Step "Installing npm dependencies" { npm install }
    }

    Invoke-Step "Compiling extension" { npm run compile }

    if ([string]::IsNullOrWhiteSpace($OutFile)) {
        Invoke-Step "Packaging extension (.vsix)" { npx --yes @vscode/vsce package }
    } else {
        $resolvedOut = $OutFile
        if (![System.IO.Path]::IsPathRooted($resolvedOut)) {
            $resolvedOut = Join-Path $repoRoot $resolvedOut
        }
        $outDir = Split-Path -Parent $resolvedOut
        if (![string]::IsNullOrWhiteSpace($outDir) -and !(Test-Path -LiteralPath $outDir)) {
            New-Item -ItemType Directory -Path $outDir | Out-Null
        }
        Invoke-Step "Packaging extension to $resolvedOut" { npx --yes @vscode/vsce package --out $resolvedOut }
    }

    Write-Host "==> Done"
} finally {
    Pop-Location
}
