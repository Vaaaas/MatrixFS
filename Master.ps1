Set-Location -PATH C:\OneDrive\Software\Go\src\github.com\Vaaaas\MatrixFS\matrix\
$mydoc=[environment]::getfolderpath("mydocuments")
Copy-Item css -Destination $mydoc\MatrixFS\Master\ -Recurse -force
Copy-Item js -Destination $mydoc\MatrixFS\Master\ -Recurse -force
Copy-Item view -Destination $mydoc\MatrixFS\Master\ -Recurse -force
Copy-Item favicon.ico -Destination $mydoc\MatrixFS\Master\ -force
go build -o $mydoc\MatrixFS\Master\master.exe github.com/Vaaaas/MatrixFS/matrix
Set-Location $mydoc\MatrixFS\Master\
mkdir log
$param="-log_dir=./log"
Start-Process -FilePath $mydoc\MatrixFS\Master\master.exe -ArgumentList $param