$mydoc=[environment]::getfolderpath("mydocuments")
cd -Path $mydoc\MatrixFS\Node\
mkdir log
mkdir storage
Start-Process -FilePath $mydoc\MatrixFS\Node\Node.exe -ArgumentList "-stpath=./storage -log_dir=./log -alsologtostderr=true -master=127.0.0.1:8080 -node=127.0.0.1:9098"
