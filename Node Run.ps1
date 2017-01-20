Set-Location -PATH C:\Node00\GO\src\github.com\Vaaaas\MatrixFS
cd node
$mydoc=[environment]::getfolderpath(“mydocuments”)
go build -o $mydoc\MatrixFS\Node\Node.exe node.go
cd -Path $mydoc\MatrixFS\Node\
Start-Process -FilePath $mydoc\MatrixFS\Node\Node.exe -ArgumentList "-stpath=./storage -log_dir=./log -master=192.168.199.201:8080 -node=192.168.199.220:9090"
