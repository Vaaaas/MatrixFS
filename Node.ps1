Set-Location -PATH $env:GOPATH\src\github.com\Vaaaas\MatrixFS
cd node
$mydoc=[environment]::getfolderpath("mydocuments")
go build -o $mydoc\MatrixFS\Node\Node.exe Node.go
cd -Path $mydoc\MatrixFS\Node\
for($i=0;$i -lt 6;$i++){
    $port=9090+$i
    $storage="storage"+$i
    $log="log"+$i
    $param="-stpath=./"+$storage+" -alsologtostderr=true -log_dir=./"+$log+" -node=127.0.0.1:"+$port
    mkdir $log
    mkdir $storage
    Start-Process -FilePath $mydoc\MatrixFS\Node\Node.exe -ArgumentList $param
}