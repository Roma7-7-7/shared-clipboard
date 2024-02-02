import {useState} from "react";
import {Alert, Button, Col, Form, InputGroup, Modal, Row} from "react-bootstrap";
import {apiBaseURL} from "../env.jsx";

const signInTitle = "Sign In";
const defaultPasswordFeedback = "Password must be at least 8 chara¡cters long and contain at least one upper case letter, at least one lower case, one number and one special character";

export default function AuthModal({title, onHide, onSignedIn}) {
    const isSignUp = title !== signInTitle;
    const [userName, setUserName] = useState("");
    const [usernameFeedback, setUsernameFeedback] = useState(null);

    const [password, setPassword] = useState("");
    const [passwordFeedback, setPasswordFeedback] = useState(null);

    const [alertMsg, setAlertMsg] = useState(null);

    function cleanup() {
        setUserName("");
        setUsernameFeedback(null);

        setPassword("");
        setPasswordFeedback(null);

        setAlertMsg(null);
    }

    let context = {
        "Sign In": {
            "usernameChange": function (event) {
                const userName = event.target.value;
                if (userName < 3) {
                    setUsernameFeedback("Username is too short");
                    return;
                }

                setUsernameFeedback("");
            },
            "passwordChange": function (event) {
                const password = event.target.value;
                if (password.length < 1) {
                    setPasswordFeedback("Password is required");
                    return;
                }

                setPasswordFeedback("");
            },
            "doSubmit": function (event, onSignedIn) {
                if (usernameFeedback !== "" || passwordFeedback !== "") {
                    event.preventDefault();
                    setAlertMsg("Both username and password are required");
                    return;
                }

                fetch(apiBaseURL + '/signin', {
                    method: "POST",
                    headers: {"Content-Type": "application/json"},
                    body: JSON.stringify({"name": userName, "password": password}),
                    credentials: 'include',
                })
                    .then(response => {
                        return response.json();
                    })
                    .then(data => {
                        if (!data["error"]) {
                            onSignedIn(data);
                            return;
                        }

                        if (data["code"] === "ERR_2201") {
                            setAlertMsg("User with such name does not exist");
                            return
                        }
                        if (data["code"] === "ERR_2103") {
                            setAlertMsg("Password is incorrect");
                            return
                        }
                        setAlertMsg(data["message"]);
                    })
                    .catch(error => {
                        console.error('Error:', error)
                        setAlertMsg("Unexpected error occurred");
                    })
            }
        },
        "Sign Up": {
            "usernameChange": function (event) {
                const userName = event.target.value;
                if (userName.length < 3) {
                    setUsernameFeedback("userName is too short");
                    return;
                }
                let firstLetter = userName[0];
                if (!((firstLetter >= 'a' && firstLetter <= 'z') || (firstLetter >= 'A' && firstLetter <= 'Z'))) {
                    setUsernameFeedback("Username must start with a letter");
                    return;
                }
                for (let i = 1; i < userName.length; i++) {
                    let letter = userName[i];
                    if (letter >= 'a' && letter <= 'z') {
                        continue;
                    }
                    if (letter >= 'A' && letter <= 'Z') {
                        continue;
                    }
                    if (letter >= '0' && letter <= '9') {
                        continue;
                    }
                    if (letter === '_' || letter === '-' || letter === '.' || letter === '@' || letter === '+') {
                        continue;
                    }
                    setUsernameFeedback("Username contains invalid characters");
                    return
                }

                setUsernameFeedback("");
            },
            "passwordChange": function (event) {
                const password = event.target.value;
                setPassword(password)
                if (password.length < 8) {
                    setPasswordFeedback(defaultPasswordFeedback);
                    return;
                }

                let hasLowerCaseLetter = false;
                let hasUpperCaseLetter = false;
                let hasNumber = false;
                let hasSpecial = false;
                for (let i = 0; i < password.length; i++) {
                    let letter = password[i];
                    if (letter >= 'a' && letter <= 'z') {
                        hasLowerCaseLetter = true;
                        continue;
                    }
                    if (letter >= 'A' && letter <= 'Z') {
                        hasUpperCaseLetter = true;
                        continue;
                    }
                    if (letter >= '0' && letter <= '9') {
                        hasNumber = true;
                        continue;
                    }
                    hasSpecial = true;
                }
                if (!(hasUpperCaseLetter && hasLowerCaseLetter && hasNumber && hasSpecial)) {
                    setPasswordFeedback(defaultPasswordFeedback);
                    return;
                }

                setPasswordFeedback("");
            },
            "doSubmit": function (event, onSignedIn) {
                if (usernameFeedback !== "" || passwordFeedback !== "") {
                    event.preventDefault();
                    setAlertMsg("Both username and password are required");
                    return;
                }

                fetch(apiBaseURL + '/signup', {
                    method: "POST",
                    headers: {"Content-Type": "application/json"},
                    body: JSON.stringify({"name": userName, "password": password}),
                    credentials: 'include',
                })
                    .then(response => {
                        return response.json();
                    })
                    .then(data => {
                        if (!data["error"]) {
                            onSignedIn(data);
                            return;
                        }

                        if (data["code"] === "ERR_2101") {
                            setAlertMsg("Password must be at least 8 character long and contain at least one uppercase letter, one lowercase letter, one digit and one special character")
                            return
                        }
                        if (data["code"] === "ERR_2102") {
                            setAlertMsg("User with such name already exists")
                            return
                        }
                        setAlertMsg(data["message"]);
                    })
                    .catch(error => {
                        console.error('Error:', error)
                        setAlertMsg("Unexpected error occurred");
                    })
            }
        }
    }

    let usernameChange = (event) => {
        setAlertMsg("");
        setUserName(event.target.value);
        context[title]["usernameChange"](event);
    }
    let passwordChange = (event) => {
        setAlertMsg("");
        setPassword(event.target.value);
        context[title]["passwordChange"](event);
    }

    const handleSubmit = (event) => {
        setAlertMsg("");

        if (usernameFeedback !== "" || passwordFeedback !== "") {
            event.preventDefault();
            event.stopPropagation();
            setAlertMsg("Both username and password are required")
            return;
        }

        setAlertMsg("");
        context[title]["doSubmit"](event, (data) => {
            cleanup();
            onSignedIn(data);
        });
    };

    return (
        <Modal show={title !== null} onHide={() => {
            cleanup();
            onHide();
        }}>
            <Modal.Header closeButton>
                <Modal.Title>{title}</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <Form onSubmit={handleSubmit}>
                    <Form.Group as={Row} className="mb-3">
                        <InputGroup hasValidation={isSignUp}>
                            <Form.Label column sm="3">User name</Form.Label>
                            <Col sm="8">
                                <Form.Control type="plaintext" onChange={usernameChange}
                                              isValid={usernameFeedback === ""}
                                              isInvalid={usernameFeedback !== null && usernameFeedback !== ""}/>
                                <Form.Control.Feedback type="invalid">
                                    {usernameFeedback}
                                </Form.Control.Feedback>
                            </Col>
                        </InputGroup>
                    </Form.Group>
                    <Form.Group as={Row} className="mb-3">
                        <InputGroup hasValidation>
                            <Form.Label column sm="3">Password</Form.Label>
                            <Col sm="8">
                                <Form.Control type="password" onChange={passwordChange}
                                              isValid={passwordFeedback === ""}
                                              isInvalid={passwordFeedback !== null && passwordFeedback !== ""}/>
                                <Form.Control.Feedback type="invalid">
                                    {passwordFeedback}
                                </Form.Control.Feedback>
                            </Col>
                        </InputGroup>
                    </Form.Group>
                </Form>
                <Alert variant="warning" show={alertMsg !== null && alertMsg !== ""}>{alertMsg}</Alert>
            </Modal.Body>
            <Modal.Footer>
                <Button variant="secondary" onClick={onHide}>Close</Button>
                <Button variant="primary" onClick={handleSubmit}
                        disabled={usernameFeedback !== "" || passwordFeedback !== ""}>{title}</Button>
            </Modal.Footer>
        </Modal>
    )
}
