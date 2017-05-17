function shorten(url) {
    var req = new XMLHttpRequest();
    req.addEventListener("load", function() {
        var res = document.querySelector("#result");
        res.value = this.responseText;
        res.classList.add("visible");
        res.focus();
        res.select();
    });
    req.open("GET", "/shorten?url=" + encodeURIComponent(url));
    req.send();
}

window.addEventListener("load", function(ev) {
    var form = document.querySelector("form");
    urlElem = document.querySelector("#url");
    form.addEventListener("submit", function(ev) {
        ev.preventDefault();
        shorten(urlElem.value);
    });
    urlElem.focus();
});

