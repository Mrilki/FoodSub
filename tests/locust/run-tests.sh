# tests/locust/run-tests.ps1 (Windows PowerShell)

param(
    [string]$Scenario = "light",
    [string]$Host = "http://localhost:8083"
)

Write-Host "Starting Locust Load Test..."
Write-Host "Scenario: $Scenario"
Write-Host "Target: $Host"

switch ($Scenario) {
    "light" {
        $users = 10
        $spawnRate = 2
        $runTime = "5m"
    }
    "medium" {
        $users = 50
        $spawnRate = 5
        $runTime = "10m"
    }
    "heavy" {
        $users = 100
        $spawnRate = 10
        $runTime = "15m"
    }
    "stress" {
        $users = 500
        $spawnRate = 50
        $runTime = "20m"
    }
}

Write-Host "Users: $users"
Write-Host "Spawn Rate: $spawnRate users/sec"
Write-Host "Run Time: $runTime"

locust -f locustfile.py `
    --host $Host `
    --headless `
    --users $users `
    --spawn-rate $spawnRate `
    --run-time $runTime `
    --html "results/locust-report-$Scenario-$(Get-Date -Format 'yyyyMMdd-HHmmss').html" `
    --csv "results/locust-results-$Scenario"

Write-Host "Test completed! Check results/ folder for reports."