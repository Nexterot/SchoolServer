$(document).ready(function(){

    // Flexslider 2
    $('.flexslider').flexslider({
        animation: "slide",
        start: function (slider) {
            $('body').removeClass('loading');
        }
    });

    // Open/Close Mobile Nav
    $(".mobile-menu").click(function(){
        if ($(this).hasClass('closed')){
            $(this).addClass("open");
            $("nav ul.main").slideDown("medium");
            $(this).removeClass("closed");
        }
        else if ($(this).hasClass('open')){
            $(this).addClass("closed");
            $("nav ul.main").slideUp("medium");
            $(this).removeClass("open");
        }
    });

    // Calulate Negative Top Margin & Apply To Element
    var anchorHeight = $(".section-nav").height();
    var anchorNegativeMargin = anchorHeight / 2 - anchorHeight;
    $(".section-nav").css("margin-top", anchorNegativeMargin).addClass("dynamic");

    // Internal Links (iDevices)
    if ("standalone" in window.navigator && window.navigator.standalone){
        // Prevent links from opening in new window (Mobile Safari)
        var a = document.getElementsByTagName("a");
        for (var i = 0; i < a.length; i++) {
            a[i].onclick = function () {
                window.location = this.getAttribute("href");
                return false
            }
        }
    }

    // Smooth scroll
    $('a[href^="#"]').on('click',function (e){
        e.preventDefault();
        var target = this.hash,
            $target = $(target);
        $('html, body').stop().animate({
            'scrollTop': $target.offset().top
        }, 700, 'swing', function () {
            window.location.hash = target;
        });
    });

    // Current Year
    var currentYear = (new Date).getFullYear();
    $("footer span.year").text( (new Date).getFullYear() );

    // Placeholders Fix For IE
    $('[placeholder]').focus(function(){
        var input = $(this);
        if (input.val() == input.attr('placeholder')) {
            input.val('');
            input.removeClass('placeholder');
        }
    }).blur(function () {
        var input = $(this);
        if (input.val() == '' || input.val() == input.attr('placeholder')) {
            input.addClass('placeholder');
            input.val(input.attr('placeholder'));
        }
    }).blur();

}); // End of "(document).ready"