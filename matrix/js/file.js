/**
 * Created by vaaaas on 2016/11/2.
 */

function validateForm(form) {
    var file = form.uploadInput.value;
    if (file == "") {
        alert("请选择文件！");
        return false;
    } else {
        return true;
    }
}