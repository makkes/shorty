function shorten(url) {

    function feedback(msg) {
        var feedbackElem = document.querySelector("#feedback");
        feedbackElem.querySelector('.text').innerHTML = msg;
        feedbackElem.classList.add('visible');
        window.setTimeout(function() {
            feedbackElem.classList.remove('visible');
        }, 3000);
    }

    if (url === '') {
        feedback('Please enter a URL to shorten');
        return;
    }

    var req = new XMLHttpRequest();
    req.addEventListener("load", function() {
        var res = document.querySelector("#result");
        if (this.status !== 200) {
            feedback('An unrecoverable error occurred.');
            return;
        }
        res.value = this.responseText;
        res.classList.add("visible");
        res.focus();
        res.select();
    });

    var s = "/shorten?url=" + encodeURIComponent(url);
    if (key) {
        s += "&key=" + encodeURIComponent(key);
    }
    req.open("GET", s);
    req.send();
}
function shorten_new(url, key) {
    var req = new XMLHttpRequest();
    req.addEventListener("load", function() {
        var res = document.querySelector("#result");
        if (this.status !== 200) {
            feedback('An unrecoverable error occurred.');
            return;
        }
        res.value = this.responseText;
        res.classList.add("visible");
        res.focus();
        res.select();
    });

    var s = "/shorten?url=" + encodeURIComponent(url);
    if (key) {
        s += "&key=" + encodeURIComponent(key);
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
