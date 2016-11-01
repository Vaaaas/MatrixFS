/**
 * Created by vaaaas on 2016/11/1.
 */

function validateForm() {
    var $submitBtn = $('#submitBtn');
    var $faultNumBox = $('#faultNumber');
    var $rowNumBox = $('#rowNumber');
    $submitBtn.click(function () {
        var faultNum = $faultNumBox.text();
        var rowNum = $rowNumBox.text();
        if (faultNum >= 2 && rowNum >= 2) {
            if (faultNum > 30 || rowNum > 30) {
                window.alert("容错数和行数不应大于30");
                return false;
            }
        } else {
            window.alert("容错数和行数应大于1");
            return false;
        }
    })

}