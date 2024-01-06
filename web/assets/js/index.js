document.addEventListener("DOMContentLoaded", function() {
    (() => {
        'use strict'

        const forms = document.querySelectorAll('.needs-validation')

        Array.from(forms).forEach(form => {
            form.addEventListener('submit', event => {
                hideAlert(modalLiveAlertPlaceholder)
                event.preventDefault()

                if (!form.checkValidity()) {
                    event.stopPropagation()
                    form.classList.add('was-validated')
                    return
                }

                modalFuncs[form.id]()
            }, false)
        })
    })()

});

var modalFuncs = {
    signInForm: function () {
        const signInLiveAlertPlaceholder = document.getElementById('signInLiveAlertPlaceholder');
        const signInUserName = document.getElementById('signInUserName');
        const signInPassword = document.getElementById('signInPassword');

        hideAlert(signInLiveAlertPlaceholder)

        fetch(apiHost + '/signin', {
            "method": "POST",
            "headers": {"Content-Type": "application/json"},
            "body": JSON.stringify({"name": signInUserName.value, "password": signInPassword.value})
        })
            .then(response => {
                if (!response.ok) {
                    if (response.status === 400) {
                        return response.json()
                    }
                    throw Error(response.statusText);
                }
                return response.json()
            })
            .then(data => {
                if (data["error"]) {
                    if (data["code"] === "ERR_2103") {
                        showAlert(signInLiveAlertPlaceholder, "User with such name does not exist")
                        return
                    }
                    showAlert(signInLiveAlertPlaceholder, data["message"])
                    return
                }

                window.location.href = "/sessions.html";
            })
            .catch(error => {
                console.error('Error:', error)
                window.location.href = "/error.html";
            })
    },

    signUpForm: function () {
        const signUpLiveAlertPlaceholder = document.getElementById('signUpLiveAlertPlaceholder');
        const signUpUserName = document.getElementById('signUpUserName');
        const signUpPassword = document.getElementById('signUpPassword');

        hideAlert(signUpLiveAlertPlaceholder)

        fetch(apiHost + '/signup', {
            "method": "POST",
            "headers": {"Content-Type": "application/json"},
            "body": JSON.stringify({"name": signUpUserName.value, "password": signUpPassword.value})
        })
            .then(response => {
                if (!response.ok) {
                    if (response.status === 400) {
                        return response.json()
                    }
                    throw Error(response.statusText);
                }
                return response.json()
            })
            .then(data => {
                if (data["error"]) {
                    if (data["code"] === "ERR_2101") {
                        showAlert(signUpLiveAlertPlaceholder, "Password must be at least 8 character long and contain at least one uppercase letter, one lowercase letter, one digit and one special character")
                        return
                    }
                    if (data["code"] === "ERR_2102") {
                        showAlert(signUpLiveAlertPlaceholder, "User with such name already exists")
                        return
                    }
                    showAlert(signUpLiveAlertPlaceholder, data["message"])
                    return
                }

                window.location.href = "/sessions.html";
            })
            .catch(error => {
                console.error('Error:', error)
                window.location.href = "/error.html";
            })
    }
}

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
