const sessionIDKey = "sessionID";

if ((window.location.pathname === "/" || window.location.pathname === "/index.html") && hasSessionID()) {
    window.location.href = "/clipboard.html";
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
