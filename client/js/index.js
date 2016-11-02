/**
 * Created by vaaaas on 2016/11/1.
 */

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