$mydoc=[environment]::getfolderpath(“mydocuments”)
cd -Path $mydoc\MatrixFS\Node\
mkdir log
mkdir storage
Start-Process -FilePath $mydoc\MatrixFS\Node\Node.exe -ArgumentList "-stpath=./storage -log_dir=./log -alsologtostderr=true -master=192.168.199.201:8080 -node=192.168.199.220:9090"
