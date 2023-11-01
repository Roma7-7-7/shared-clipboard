removeSessionID()

document.addEventListener("DOMContentLoaded", function() {
    const sessionIdElement = document.getElementById("sessionID");
    const proceedButton = document.getElementById('proceedButton');

    proceedButton.addEventListener("click", function () {
        storeSessionID(sessionIdElement.innerText);
        window.location.href = "/clipboard.html";
    })

    sessionIdElement.addEventListener("click", function () {
        const trimmedText = sessionIdElement.textContent.trim();
        const textArea = document.createElement('textarea');
        textArea.value = trimmedText;
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);
    });

    fetch(apiHost + '/sessions', {"method": "POST"})
        .then(response => {
            if (!response.ok) {
                throw Error(response.statusText);
            }
            return response.json()
        })
        .then(data => {
            proceedButton.removeAttribute("disabled")
            sessionIdElement.innerText = data["session_id"].trim();
        })
        .catch(error => {
            console.error('Error:', error)
            proceedButton.setAttribute("disabled", "disabled")
            window.location.href = "/error.html";
        })
});