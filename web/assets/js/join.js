removeSessionID()

document.addEventListener("DOMContentLoaded", function() {
    const liveAlertPlaceholder = document.getElementById('liveAlertPlaceholder');
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
                    window.location.href = '/clipboard.html'
                    return
                }

                if (response.status === 404) {
                    setAlert(liveAlertPlaceholder, 'Session not found')
                }
            })
            .catch(error => {
                console.error('Error:', error)
                window.location.href = "/error.html";
            })
    });
});

function setAlert(placeholder, message) {
    placeholder.innerHTML = ''

    const wrapper = document.createElement('div')
    wrapper.innerHTML = [
        `<div class="alert alert-danger alert-dismissible" role="alert">`,
        `   <div>${message}</div>`,
        '   <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>',
        '</div>'
    ].join('')

    placeholder.append(wrapper)
}
