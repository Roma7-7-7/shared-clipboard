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
        fetch(apiHost + '/sessions?joinKey=' + sessionIDInput.value)
            .then(response => {
                return response.json()
            })
            .then(data => {
                if (data['error']) {
                    showAlert(liveAlertPlaceholder, data['message'])
                    return;
                }

                storeSessionID(data['session_id'])
                window.location.href = '/clipboard.html'
            })
            .catch(error => {
                showAlert(liveAlertPlaceholder, 'Failed to join session')
            })
    });
});

function showAlert(placeholder, message) {
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
