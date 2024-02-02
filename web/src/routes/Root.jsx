import {useState} from "react";
import {Alert, Modal} from "react-bootstrap";
import {Navigate, Outlet} from "react-router-dom";
import {apiBaseURL} from "../env.jsx";
import {getUserInfo, setUserInfo} from "../storage.js";
import Navbar from "../components/Navbar.jsx";
import AuthModal from "../components/AuthModal.jsx";

export default function Root() {
    const [authModalTitle, setAuthModalTitle] = useState(null)
    const [alertMsg, setAlertMsg] = useState(null)

    function onSignOutClicked() {
        fetch(apiBaseURL + '/signout', {
            "method": "POST",
            "headers": {"Content-Type": "application/json"},
            "body": JSON.stringify({}),
            credentials: 'include',
        })
            .then(response => {
                if (response.status === 204) {
                    return Promise.resolve({"error": false, "message": "Signed out successfully"});
                }
                return response.json();
            })
            .then(data => {
                if (!data["error"]) {
                    userInfo = setUserInfo(null);
                    window.location.pathname = "/";
                    return;
                }

                setAlertMsg(data["message"]);
            })
            .catch(error => {
                console.error('Error:', error)
                setAlertMsg("Unexpected error occurred");
            })
    }

    function onSignInClicked() {
        setAuthModalTitle("Sign In");
    }

    function onSignUpClicked() {
        setAuthModalTitle("Sign Up");
    }

    let userInfo = getUserInfo();

    if (!userInfo && window.location.pathname !== "/") {
        window.location.pathname = "/";
        return (<></>)
    }
    if (userInfo && window.location.pathname === "/") {
        window.location.pathname = "/sessions";
        return (<></>)
    }

    return (
        <>
            <Navbar userInfo={userInfo} onSignInClicked={onSignInClicked} onSignUpClicked={onSignUpClicked} onSignOutClicked={onSignOutClicked}/>
            <AuthModal title={authModalTitle} onHide={() => setAuthModalTitle(null)} onSignedIn={(userInfo) => {
                setAuthModalTitle(null);
                setUserInfo(userInfo);
            }}/>
            <Modal show={alertMsg !== null} onHide={() => setAlertMsg(null)}>
                <Modal.Header closeButton>
                    <Modal.Title>Error</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <Alert variant="danger">
                        {alertMsg}
                    </Alert>
                </Modal.Body>
            </Modal>
            <Outlet context={{
                userInfo: userInfo,
                onSignInClicked: onSignInClicked,
                onSignUpClicked: onSignUpClicked,
                onSignOutClicked: onSignOutClicked,
            }} />
        </>
    )
}
