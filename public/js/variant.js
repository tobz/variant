function loadFragment(targetElement, targetEndpoint) {
    $(targetElement).html($('<img class="loading" src="/public/img/loading.gif" />'));

    $.ajax(targetEndpoint, {
        success: function(data, status, xhr) {
            $(targetElement).html(data)
        },
        error: function(xhr, status, err) {
            $(targetElement).html("Bummer dude.  We couldn't get a response from the server.")
        }
    })
}
