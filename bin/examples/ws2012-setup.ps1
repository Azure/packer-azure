Write-Host "Starting file script..." 

Write-Host 'Executing [DateTime]::Now...' 
[DateTime]::Now

Write-Host 'Executing Install-WindowsFeature -Name "XPS-Viewer" -IncludeAllSubFeature' 
Install-WindowsFeature -Name "XPS-Viewer" -IncludeAllSubFeature

Write-Host 'Installing ReportViewer...' 
$filename = "ReportViewer.exe";
$link = "http://download.microsoft.com/download/E/A/1/EA1BF9E8-D164-4354-8959-F96843DD8F46/$filename";
$dstDir = "c:\ShchFileFolder";
New-Item $dstDir -type directory -force | Out-Null
$remotePath = Join-Path $dstDir $filename;
(New-Object System.Net.Webclient).downloadfile($link, $remotePath);
Start-Process $remotePath -NoNewWindow -Wait -Argument '/passive';

Write-Host 'Installing Git Bash...' 
$filename = "Git-1.9.4-preview20140611.exe";
$link = "https://github.com/msysgit/msysgit/releases/download/Git-1.9.4-preview20140611/$filename";
$dstDir = "c:\ShchFileFolder";
New-Item $dstDir -type directory -force | Out-Null
$remotePath = Join-Path $dstDir $filename;
(New-Object System.Net.Webclient).downloadfile($link, $remotePath);
Start-Process $remotePath -NoNewWindow -Wait -Argument '/VERYSILENT /CLOSEAPPLICATIONS';

Write-Host 'Installing Mozilla Firefox...' 
$filename = "Firefox Setup 31.0.exe";
$link = "https://download.mozilla.org/?product=firefox-31.0-SSL&os=win&lang=en-US";
$dstDir = "c:\ShchFileFolder";
New-Item $dstDir -type directory -force | Out-Null
$remotePath = Join-Path $dstDir $filename;
(New-Object System.Net.Webclient).downloadfile($link, $remotePath);
Start-Process $remotePath -NoNewWindow -Wait -Argument '/S';

Write-Host "File script finished!" 
