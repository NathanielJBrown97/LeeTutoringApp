function onSignUp(googleUser) {
    var profile = googleUser.getBasicProfile();
    var idToken = googleUser.getAuthResponse().id_token;

    // Send the ID token to the backend to create the parent account
    fetch('/create-parent', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + idToken
        },
        body: JSON.stringify({
            email: profile.getEmail(),
            name: profile.getName()
        })
    }).then(response => response.json())
    .then(data => {
        if (data.success) {
            window.location.replace('/signed-in.html');
        } else {
            alert('Failed to sign up. Please try again.');
        }
    });
}
