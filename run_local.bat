@echo off
setlocal enabledelayedexpansion

:: ──────────────────────────────────────────────
:: CONFIGURATION - ADAPT ON YOUR ENVIRONNMENT
:: ──────────────────────────────────────────────
set PROJECT_ID=reliable-plasma-476112-q3
set REGION=europe-west1
set CLUSTER_NAME=cluster-test
set SA_KEY_PATH=%cd%\ci-runner-access.json
set IMAGE_NAME=gke-allow-runner-action
set ACTION=cleanup

:: ──────────────────────────────────────────────
echo.
echo GKE Allow GitHub Runner - Mode: %ACTION%
echo.

docker run --rm ^
  -v "%SA_KEY_PATH%:/tmp/sa.json" ^
  -e GOOGLE_APPLICATION_CREDENTIALS=/tmp/sa.json ^
  -e INPUT_CLUSTER_NAME=%CLUSTER_NAME% ^
  -e INPUT_REGION=%REGION% ^
  -e INPUT_PROJECT_ID=%PROJECT_ID% ^
  -e INPUT_ACTION=%ACTION% ^
  %IMAGE_NAME%

echo.
echo Script finished.
pause
