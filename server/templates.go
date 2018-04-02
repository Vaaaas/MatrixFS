package server

import "html/template"

var F0fTpl = template.Must(template.New("404Page").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>404</title>
    <!-- Bootstrap -->
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">
    <style>
        .center {
            text-align: center;
            margin-left: auto;
            margin-right: auto;
            margin-bottom: auto;
            margin-top: auto;
        }
    </style>
</head>
<body>
<nav class="navbar navbar-inverse navbar-fixed-top" role="navigation">
    <div class="container-fluid">
        <div class="navbar-header">
            <a class="navbar-brand" href="./file">MatrixFS</a>
        </div>
        <div id="navbar" class="navbar-collapse collapse">
            <ul class="nav navbar-nav navbar-right">
                <li><a href="./file">文件管理</a></li>
                <li><a href="#">帮助</a></li>
            </ul>
        </div>
    </div>
</nav>
</br>
</br>
</br>
</br>
<div class="container">
    <div class="row">
        <div class="span12">
            <div class="hero-unit center">
                <h1>Page Not Found
                    <small><font face="Tahoma" color="red">Error 404</font></small>
                </h1>
                <br/>
                <p>The page you requested could not be found, either contact your webmaster or try again. Use your
                    browsers <b>Back</b> button to navigate to the page you have prevously come from</p>
            </div>
            <br/>
        </div>
    </div>
</div>
<script src="https://ajax.aspnetcdn.com/ajax/jQuery/jquery-2.2.4.min.js"></script>
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>
</body>
</html>
`))

var IndexTpl = template.Must(template.New("Index").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>MatrixFS</title>
    <!-- Bootstrap -->
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">
</head>
<body>
<div class="page-header">
    <h1 style="text-align: center;"><img style="height:75px; display:inline-block" src="https://ss1.bdstatic.com/70cFuXSh_Q1YnxGkpoWK1HF6hhy/it/u=1105799530,2716486105&fm=27&gp=0.jpg">
        MatrixFS </h1>
</div>
</br>
</br>
</br>
<div class="container main" style="width: 500px;">
    <form name="sysConfig" role="form" action="/node" method="post" onsubmit="return validateForm(this)">
        <div class="form-group">
            <label for="faultNumber">系统最大容错数</label>
            <input type="number" class="form-control" id="faultNumber" name="faultNumber"
                   placeholder="Enter number of faults">
        </div>
        </br>
        <div class="form-group">
            <label for="rowNumber">分块阵列行数</label>
            <input type="number" class="form-control" id="rowNumber" name="rowNumber"
                   placeholder="Enter number of rows">
        </div>
        </br>
        </br>
        <input type="submit" class="btn btn-default center-block" id="submitBtn" value="确认">
    </form>
</div>
<script src="https://ajax.aspnetcdn.com/ajax/jQuery/jquery-2.2.4.min.js"></script>
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>
<script>
    function validateForm(form) {
        var faultNum = form.faultNumber.value;
        var rowNum = form.rowNumber.value;
        if (faultNum >= 2 && rowNum >= 2) {
            if (faultNum > 30 || rowNum > 30) {
                alert("容错数和行数不应大于30");
                return false;
            } else {
                return true;
            }
        }
        else {
            alert('容错数和行数应大于1');
            return false;
        }
    }
</script>
</body>
</html>
`))

var FileTpl = template.Must(template.New("File").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>File Manage</title>
    <!-- Bootstrap -->
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">
</head>
<body>
<div class="modal fade" id="modalDialog" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
    <div class="modal-dialog">
        <div class="modal-content">
            <div class="modal-header">
                <button type="button" class="close" data-dismiss="modal"><span aria-hidden="true">&times;</span><span
                        class="sr-only">Close</span></button>
                <h4 class="modal-title" id="myModalLabel">上传文件</h4>
            </div>
            <form class="form-inline" enctype="multipart/form-data" role="form" name="uploadForm" action="/upload"  onsubmit="return validateForm(this)" method="post">
                <div class="modal-body">
                    <div class="form-group">
                        <div class="input-group">
                            <input type="file" class="file" name="uploadInput" id="inputTag">
                        </div>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-default" data-dismiss="modal">取消</button>
                    <input type="submit" class="btn btn-primary" value="上传">
                </div>
            </form>
        </div>
    </div>
</div>
<nav class="navbar navbar-inverse navbar-fixed-top" role="navigation">
    <div class="container-fluid">
        <div class="navbar-header">
            <a class="navbar-brand" href="./node">MatrixFS</a>
        </div>
        <div id="navbar" class="navbar-collapse collapse">
            <ul class="nav navbar-nav navbar-right">
                <li><a href="./node">系统信息</a></li>
                <li><a href="#">帮助</a></li>
            </ul>
        </div>
    </div>
</nav>
</br>
</br>
</br>
</br>
<div class="container-fluid">
    <div class="row">
        <div class="col-sm-10 col-sm-offset-1 col-md-10 col-md-offset-1 main">
            <div class="container row">
                <h1 class="sub-header col-md-9 col-lg-10">文件列表</h1>
                <button class="btn btn-success col-md-1 col-lg-1" data-toggle="modal" data-target="#modalDialog"
                        style="margin-top:30px">添加文件
                </button>
            </div>
            <div class="table-responsive row">
                <table class="table table-striped">
                    <thead>
                    <tr>
                        <th>文件名</th>
                        <th>文件大小</th>
                        <th>下载</th>
                        <th>删除</th>
                    </tr>
                    </thead>
                    <tbody>

                    {{range .}}
                    <tr>
                        <td>{{.FileFullName}}</td>
                        <td>{{.Size}}</td>
                        <td>
                            <a href="/download?fileName={{.FileFullName}}" target="_blank">
                                <button type="button" class="btn btn-info">下载
                                </button>
                            </a>
                        </td>
                        <td>
                            <a href="/delete?fileName={{.FileFullName}}">
                                <button type="button" class="btn btn-danger">删除
                                </button>
                            </a>
                        </td>
                    </tr>
                    {{end}}
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</div>
<script src="https://ajax.aspnetcdn.com/ajax/jQuery/jquery-2.2.4.min.js"></script>
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>
<script>
    function validateForm(form) {
        var file = form.uploadInput.value;
        if (file == "") {
            alert("请选择文件！");
            return false;
        } else {
            return true;
        }
    }
</script>
</body>
</html>
`))

var NodeTpl = template.Must(template.New("Node").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Node Manage</title>
    <!-- Bootstrap -->
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">
</head>
<body>
<nav class="navbar navbar-inverse navbar-fixed-top" role="navigation">
    <div class="container-fluid">
        <div class="navbar-header">
            <a class="navbar-brand" href="./file">MatrixFS</a>
        </div>
        <div id="navbar" class="navbar-collapse collapse">
            <ul class="nav navbar-nav navbar-right">
                <li><a href="./file">文件管理</a></li>
                <li><a href="#">帮助</a></li>
            </ul>
        </div>
    </div>
</nav>
</br>
</br>
</br>
</br>
<div class="container-fluid">
    <div class="row">
        <div class="col-sm-10 col-sm-offset-1 col-md-10 col-md-offset-1 main">
            <div class="container row">
                <h1 class="sub-header col-md-9 col-lg-10">节点列表</h1>

                {{if .SystemStatus}} 
                    <button class="btn btn-disable col-md-1 col-lg-1" data-toggle="modal" style="margin-top:30px;pointer-events: none;cursor: default;">恢复系统</button>
                {{else}}
                    <button class="btn btn-warning col-md-1 col-lg-1" data-toggle="modal" style="margin-top:30px;"><a href="./restore">恢复系统</a></button>
                {{end}}
            </div>
            <div class="table-responsive row">
                <table class="table table-striped">
                    <thead>
                    <tr>
                        <th>节点ID</th>
                        <th>节点IP</th>
                        <th>节点端口号</th>
                        <th>剩余空间</th>
                        <th>节点状态</th>
                    </tr>
                    </thead>
                    <tbody>
                    {{range .Nodes}}
                    <tr>
                        <td>{{.ID}}</td>
                        <td>{{.Address}}</td>
                        <td>{{.Port}}</td>
                        <td>{{.Volume}} GB</td>
                        <td>{{if .Status}} 运行中 {{else}} 丢失 {{end}}</td>
                    </tr>
                    {{end}}
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</div>
<script src="https://ajax.aspnetcdn.com/ajax/jQuery/jquery-2.2.4.min.js"></script>
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>
</body>
</html>
`))

var InfoTpl = template.Must(template.New("Info").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Message</title>
    <!-- Bootstrap -->
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">
    <style>
        .center {text-align: center; margin-left: auto; margin-right: auto; margin-bottom: auto; margin-top: auto;}
    </style>
</head>
<body>
<nav class="navbar navbar-inverse navbar-fixed-top" role="navigation">
    <div class="container-fluid">
        <div class="navbar-header">
            <a class="navbar-brand" href="./file">MatrixFS</a>
        </div>
        <div id="navbar" class="navbar-collapse collapse">
            <ul class="nav navbar-nav navbar-right">
                <li><a href="./node">系统信息</a></li>
                <li><a href="#">帮助</a></li>
            </ul>
        </div>
    </div>
</nav>
</br>
</br>
</br>
</br>
<div class="container">
    <div class="row">
        <div class="span12">
            <div class="hero-unit center">
                <h1><font face="Tahoma" color="red">{{.info}}</font></h1>
                <br />
            </div>
            <br />
        </div>
    </div>
</div>
<script src="https://ajax.aspnetcdn.com/ajax/jQuery/jquery-2.2.4.min.js"></script>
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>
</body>
</html>
`))