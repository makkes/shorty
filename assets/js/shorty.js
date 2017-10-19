function shorten(url, key) {
    var req = new XMLHttpRequest();
    req.addEventListener("load", function() {
        var res = document.querySelector("#result");
        res.value = this.responseText;
        res.classList.add("visible");
        res.focus();
        res.select();
    });

    var s = "/shorten?url=" + encodeURIComponent(url);
    if (key) {
        s += "&key=" + encodedURIComponent(key);
    }
    req.open("GET", s);
    req.send();
}

window.addEventListener("load", function(ev) {
    var form = document.querySelector("form");
    urlElem = document.querySelector("#url");
    keyElem = document.querySelector("#key");
    form.addEventListener("submit", function(ev) {
        ev.preventDefault();
        shorten(urlElem.value, keyElem.value);
    });
    urlElem.focus();
});

