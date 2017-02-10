Set-Location -PATH C:\OneDrive\Software\Go\src\github.com\Vaaaas\MatrixFS\matrix\
$mydoc=[environment]::getfolderpath("mydocuments")
Copy-Item js -Destination $mydoc\MatrixFS\Master\ -Recurse -force
Copy-Item view -Destination $mydoc\MatrixFS\Master\ -Recurse -force
Copy-Item favicon.ico -Destination $mydoc\MatrixFS\Master\ -force
go build -o $mydoc\MatrixFS\Master\Master.exe github.com/Vaaaas/MatrixFS/matrix
Set-Location $mydoc\MatrixFS\Master\
mkdir log
mkdir temp
$param="-log_dir=./log -alsologtostderr=true"
Start-Process -FilePath $mydoc\MatrixFS\Master\Master.exe -ArgumentList $param