import {useState} from "react";
import Navbar from "./Navbar.jsx";
import AuthModal from "./AuthModal.jsx";
import {apiBaseURL} from "./env.jsx";
import {Alert, Container, Row, Modal, Button} from "react-bootstrap";

function App() {
    const [firstLoad, setFirstLoad] = useState(true)
    const [authModalTitle, setAuthModalTitle] = useState(null)
    const [userInfo, setUserInfo] = useState(null)
    const [alertMsg, setAlertMsg] = useState(null)

    function fetchUserInfo(initial) {
        fetch(apiBaseURL + '/v1/user/info', {
            credentials: 'include',
        })
            .then(response => {
                if (response.status === 401) {
                    return Promise.resolve(null);
                }
                return response.json();
            })
            .then(data => {
                if (data === null) {
                    if (initial) {
                        return;
                    }
                    setAlertMsg("You are not signed in")
                    return;
                }

                if (!data["error"]) {
                    setAuthModalTitle(null);
                    setUserInfo(data);
                    return;
                }


                if (initial) {
                    return;
                }
                setAlertMsg(data["message"]);
            })
            .catch(error => {
                console.error('Error:', error)
                setAlertMsg("Unexpected error occurred");
            })
    }

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
                    setUserInfo(null);
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

    if (firstLoad) {
        setFirstLoad(false);
        fetchUserInfo(true);
    }

    let mainContent = (
        <Row>
            <p className="fs-5">The purpose of this service is to share clipboard content across multiple hosts.</p>
            <p>Please <a href="#" onClick={onSignInClicked}>sign in</a> or <a href="#" onClick={onSignUpClicked}>sign up</a> to start using it.</p>
        </Row>
    )

    if (userInfo !== null) {
        mainContent = (
            <Row>
                <p className="fs-5">You are signed in as <strong>{userInfo["name"]}</strong>.</p>
                <p>Click <a href="#" onClick={onSignOutClicked}>here</a> to sign out.</p>
            </Row>
        )
    }

    return (
        <>
            <Navbar userInfo={userInfo} onSignInClicked={onSignInClicked} onSignUpClicked={onSignUpClicked} onSignOutClicked={onSignOutClicked}/>
            <AuthModal title={authModalTitle} onHide={() => setAuthModalTitle(null)} onSignedIn={() => {
                setAuthModalTitle(null);
                fetchUserInfo(false);
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
            <Container className="mt-5">
                {mainContent}
            </Container>
        </>
    )
}

export default App
