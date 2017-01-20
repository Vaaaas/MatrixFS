Set-Location -PATH C:\OneDrive\Software\Go\src\github.com\Vaaaas\MatrixFS
cd node
$mydoc=[environment]::getfolderpath(“mydocuments”)
go build -o $mydoc\MatrixFS\Node\Node.exe node.go
cd -Path $mydoc\MatrixFS\Node\
for($i=0;$i -lt 26;$i++){
    $port=9090+$i
    $storage="storage"+$i
    $log="log"+$i
    $param="-stpath=./"+$storage+" -log_dir=./"+$log+" -node=127.0.0.1:"+$port
    $param
    Start-Process -FilePath $mydoc\MatrixFS\Node\Node.exe -ArgumentList $param
}