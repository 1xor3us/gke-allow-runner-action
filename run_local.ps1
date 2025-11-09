# GKE-Allow.ps1

# ──────────────────────────────────────────────
# ADAPT FOR YOUR ENVIRONNMENT
# ──────────────────────────────────────────────
$env:PROJECT_ID = "reliable-plasma-476112-q3"
$env:SA_KEY_PATH = "$(Get-Location)\ci-runner-access.json"
$env:IMAGE_NAME = "gke-allow-runner-action"
$env:ACTION = "cleanup"

# CLUSTERS
$env:INPUT_CLUSTERS = @"
europe-west1/cluster-test
europe-west1/cluster-test-eu
us-central1/cluster-test-us
"@
# ──────────────────────────────────────────────

# ──────────────────────────────────────────────
# MAIN (DONT TOUCH)
# ──────────────────────────────────────────────
Write-Host ""
Write-Host "GKE Allow GitHub Runner - Mode: $env:ACTION"
Write-Host ""

Write-Host "📋 Clusters configured:"
$env:INPUT_CLUSTERS -split "`n" | ForEach-Object { Write-Host "  - $_" }

docker run --rm `
  -v "$($env:SA_KEY_PATH):/tmp/sa.json" `
  -e GOOGLE_APPLICATION_CREDENTIALS=/tmp/sa.json `
  -e INPUT_CLUSTERS="$($env:INPUT_CLUSTERS)" `
  -e INPUT_PROJECT_ID="$env:PROJECT_ID" `
  -e INPUT_ACTION="$env:ACTION" `
  $env:IMAGE_NAME

Write-Host ""
Write-Host "Script finished."
Pause
