removeSessionID()

document.addEventListener("DOMContentLoaded", function() {
    const sessionIdElement = document.getElementById("sessionID");
    const proceedButton = document.getElementById('proceedButton');

    proceedButton.addEventListener("click", function () {
        storeSessionID(sessionIdElement.innerText);
        redirect("/clipboard.html")
    })

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
            sessionIdElement.innerText = data["session_id"];
        })
        .catch(redirectError)
});