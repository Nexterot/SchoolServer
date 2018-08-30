$('window').load(function(){
    if (/webkit/.test(navigator.userAgent.toLowerCase())) {
        $('[autocomplete="on"]').each(function() {
            $(this).attr('autocomplete', 'off');
        });
    }
});