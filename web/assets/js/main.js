redirectOnClick(document.querySelector(".navbar h1"), "/")
redirectOnClick(document.getElementById("initButton"), "/init.html");
redirectOnClick(document.getElementById("homeButton"), "/");

function redirectOnClick(elem, url) {
    if (elem) {
        elem.addEventListener("click", function() {
            window.location.href = url;
        });
    }
}