function shorten(url, key) {

    function feedback(msg) {
        var feedbackElem = document.querySelector("#feedback");
        feedbackElem.querySelector('.text').innerHTML = msg;
        feedbackElem.classList.add('visible');
        window.setTimeout(function () {
            feedbackElem.classList.remove('visible');
        }, 3000);
    }

    if (url === '') {
        feedback('Please enter a URL to shorten');
        return;
    }

    var req = new XMLHttpRequest();
    req.addEventListener("load", function () {
        var res = document.querySelector("#result");
        if (this.status !== 200) {
            let txt = this.responseText;
            if (this.statusText.length > 0) {
                txt = this.statusText + ": " + txt
            }
            feedback(txt);
            return;
        }
        res.value = this.responseText;
        res.classList.add("visible");
        res.focus();
        res.select();
    });
    req.open("GET", "/shorten?url=" + encodeURIComponent(url) + "&key=" + encodeURIComponent(key));
    req.send();
}

window.addEventListener("load", function (ev) {
    var form = document.querySelector("form");
    urlElem = document.querySelector("#url");
    keyElem = document.querySelector("#key");
    form.addEventListener("submit", function (ev) {
        ev.preventDefault();
        shorten(urlElem.value, keyElem.value);
    });
    urlElem.focus();
});
