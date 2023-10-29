const sessionIdElement = document.getElementById("sessionID");

sessionIdElement.addEventListener("click", function () {
    const range = document.createRange();
    range.selectNode(sessionIdElement);
    window.getSelection().removeAllRanges(); // clear current selection
    window.getSelection().addRange(range); // to select text
    document.execCommand("copy");
    window.getSelection().removeAllRanges();// to deselect
});

fetch(apiHost + '/sessions', {"method": "POST"})
    .then(response => {
        if (!response.ok) {
            throw Error(response.statusText);
        }
        return response.json()
    })
    .then(data => {
        let sessionID = data["session_id"];
        sessionIdElement.innerText = sessionID;
        storeSessionID(sessionID)
    })
    .catch(redirectError)