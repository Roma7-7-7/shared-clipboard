removeSessionID()

document.addEventListener("DOMContentLoaded", function() {
    const proceedButton = document.getElementById('proceedButton');
    const sessionIDInput = document.getElementById('sessionIDInput');

    sessionIDInput.addEventListener('input', function () {
        proceedButton.disabled = sessionIDInput.value.length <= 0;
    });

    sessionIDInput.addEventListener('keyup', function (event) {
        if (event.key === 'Enter') {
            proceedButton.click()
        }
    });

    proceedButton.addEventListener('click', function () {
        fetch(apiHost + '/sessions/' + sessionIDInput.value)
            .then(response => {
                if (response.ok) {
                    storeSessionID(sessionIDInput.value)
                    redirect('/clipboard.html')
                    return
                }

                if (response.status === 404) {
                    alert('Session not found')
                }
            })
            .catch(redirectError)
    });
});