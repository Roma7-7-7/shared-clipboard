const sessionIDKey = "sessionID";

if ((window.location.pathname === "/" || window.location.pathname === "/index.html") && hasSessionID()) {
    redirect("/clipboard.html");
}

function storeSessionID(sessionID) {
    localStorage.setItem(sessionIDKey, sessionID);
}

function getSessionID() {
    return localStorage.getItem(sessionIDKey);
}

function removeSessionID() {
    localStorage.removeItem(sessionIDKey);
}

function hasSessionID() {
    return getSessionID() !== null;
}

function redirect(path) {
    window.location.href = path;
}

function redirectError(err) {
    console.log(err);
    redirect("/error.html");
}

function redirectOnClick(elem, url) {
    if (elem) {
        elem.addEventListener("click", function() {
            redirect(url);
        });
    }
}

document.addEventListener("DOMContentLoaded", function() {
    redirectOnClick(document.querySelector(".navbar h1"), "/")
    redirectOnClick(document.getElementById("startButton"), "/start.html");
    redirectOnClick(document.getElementById("homeButton"), "/");
    redirectOnClick(document.getElementById("joinButton"), "/join.html");
});