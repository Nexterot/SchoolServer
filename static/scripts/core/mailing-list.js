/*///////////////////////////////////////////////////////////////////////
Ported to jquery from prototype by Joel Lisenby (joel.lisenby@gmail.com)
http://joellisenby.com

original prototype code by Aarron Walter (aarron@buildingfindablewebsites.com)
http://buildingfindablewebsites.com

Distrbuted under Creative Commons license
http://creativecommons.org/licenses/by-sa/3.0/us/
///////////////////////////////////////////////////////////////////////*/

$(document).ready(function() {

    $('#signup').submit(function() {
        // update user interface
        $('#response').html('Adding email address...');

        // Prepare query string and send AJAX request
        $.ajax({
            url: 'inc/store-address.php',
            data: 'ajax=true&email=' + escape($('#email').val()),
            success: function(msg) {
                $('#response').html(msg);
                $('input#email').val("");
            }
        });

        return false;
    });

    $('#signup2').submit(function() {
        // update user interface
        $('#response2').html('Adding email address...');

        // Prepare query string and send AJAX request
        $.ajax({
            url: 'inc/store-address.php',
            data: 'ajax=true&email=' + escape($('#email2').val()),
            success: function(msg) {
                $('#response2').html(msg);
                $('input#email2').val("");
            }
        });

        return false;
    });
});