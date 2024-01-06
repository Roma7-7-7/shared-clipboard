document.addEventListener("DOMContentLoaded", function() {
    const createSessionLink = document.getElementById('createSessionLink');

    fetch(apiHost + '/v1/sessions', {credentials: 'include'})
        .then(response => {
            if (!response.ok) {
                throw Error(response.statusText);
            }
            return response.json()
        })
        .then(data => {
            console.log(data)
        })
        .catch(error => {
            console.error('Error:', error)
            window.location.href = "/error.html";
        })
})