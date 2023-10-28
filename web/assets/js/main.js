redirectOnClick(document.querySelector(".navbar h1"), "/")
redirectOnClick(document.getElementById("initButton"), "/init.html");
redirectOnClick(document.getElementById("homeButton"), "/");
redirectOnClick(document.getElementById("joinButton"), "/join.html");

function redirectOnClick(elem, url) {
    if (elem) {
        elem.addEventListener("click", function() {
            redirect(url);
        });
    }
}

function storeSessionID(sessionID) {
    localStorage.setItem("sessionID", sessionID);
}

function redirect(path) {
    window.location.href = path;
}

function redirectError(err) {
    console.log(err);
    redirect("/error.html");
}