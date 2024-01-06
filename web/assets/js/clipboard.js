let lastModified = "";

document.addEventListener("DOMContentLoaded", function () {
    const liveAlertPlaceholder = document.getElementById('liveAlertPlaceholder');
    const clipboardText = document.getElementById('clipboardText');
    const shareTextButton = document.getElementById('shareTextButton');

    shareTextButton.addEventListener('click', function () {
        hideAlert(liveAlertPlaceholder)

        fetch(apiHost + '/sessions/' + getSessionID() + "/clipboard", {
            method: 'PUT',
            headers: {
                'Content-Type': 'text/plain'
            },
            body: clipboardText.value
        })
            .then(response => {
                if (response.status === 204) {
                    return Promise.resolve(null);
                }
                return response.json()
            })
            .then(data => {
                if (data === null) {
                    showSuccessAlert(liveAlertPlaceholder, 'Content shared');
                    return;
                }
                if (data['error']) {
                    showDangerAlert(liveAlertPlaceholder, data['message']);
                    return;
                }
                showSuccessAlert(liveAlertPlaceholder, 'Content shared');
            })
            .catch(error => {
                console.error('Error:', error)
                showDangerAlert(liveAlertPlaceholder, 'Failed to share content');
            })
    });

    setInterval(function () {
        fetch(apiHost + '/sessions/' + getSessionID() + '/clipboard', {
            method: 'GET',
            headers: {
                'If-Modified-Since': lastModified
            }
        })
            .then(response => {
                if (response.status === 200) {
                    lastModified = response.headers.get('Last-Modified');
                    return response.body.getReader().read()
                }
                if (response.status === 204 || response.status === 304) {
                    return Promise.resolve(null);
                }

                throw new Error('Unexpected response');
            })
            .then(data => {
                if (data === null) {
                    return;
                }
                if (data['error']) {
                    showDangerAlert(liveAlertPlaceholder, data['message']);
                    return;
                }
                clipboardText.value = new TextDecoder().decode(data.value);
            })
            .catch(error => {
                console.error('Error:', error)
                showDangerAlert(liveAlertPlaceholder, 'Failed to get shared clipboard content')
            })
    }, 1000);
});

function showSuccessAlert(placeholder, message) {
    placeholder.innerHTML = ''

    const wrapper = document.createElement('div')
    wrapper.innerHTML = [
        `<div class="alert alert-success alert-dismissible" role="alert">`,
        `   <div>${message}</div>`,
        '   <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>',
        '</div>'
    ].join('')

    placeholder.append(wrapper)
}

function showDangerAlert(placeholder, message) {
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
