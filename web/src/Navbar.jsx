import {useState} from 'react';
import {Container, Row, Col, Navbar as NavbarB, Nav, Button, Modal, Form, InputGroup, Alert} from "react-bootstrap";

export default function Navbar() {
    const [showModal, setShowModal] = useState(false);
    const [modalTitle, setModalTitle] = useState("Sign In");

    const handleModalClose = () => setShowModal(false);

    function handleModalShow(title) {
        setModalTitle(title);
        setShowModal(true);
    }

    return (
        <>
            <NavbarB expand="lg" bg="dark" variant="dark">
                <Container>
                    <NavbarB.Brand href="#home">Clipboard share</NavbarB.Brand>
                    <NavbarB.Collapse>
                        <Nav className="me-auto"></Nav>
                        <Nav>
                            <Button variant="outline-primary me-2" onClick={() => {handleModalShow('Sign In')}}>Sign In</Button>
                            <Button variant="outline-secondary me-2" onClick={() => {
                        </Nav>
                    </NavbarB.Collapse>
                </Container>
            </NavbarB>

            <AuthModal show={showModal} onHide={handleModalClose} title={modalTitle} />
        </>
    )
}

const signInTitle = "Sign In";
const defaultPasswordFeedback = "Password must be at least 6 characters long and contain at least one letter, one number and one special character";
const defaultUsernameFeedback = "Username must be at least 3 characters long and start with a letter";

function AuthModal({show, onHide, title}) {
    const hasValidation = title !== signInTitle;
    const [usernameFeedback, setUsernameFeedback] = useState(defaultUsernameFeedback);
    const [usernameValid, setUsernameValid] = useState(false);
    const [passwordFeedback, setPasswordFeedback] = useState(defaultPasswordFeedback);
    const [passwordValid, setPasswordValid] = useState(false);

    let submitActions = {
        "Sign In": () => {console.log("Sign In")},
        "Sign Up": () => {console.log("Sign Up")}
    }

    let onHideDelegate = onHide;
    onHide = () => {
        setUsernameValid(false);
        setUsernameFeedback(defaultUsernameFeedback);
        setPasswordValid(false);
        setPasswordFeedback(defaultPasswordFeedback);
        onHideDelegate();
    }
    const handleSubmit = (event) => {
        // const form = event.currentTarget;
        if (!usernameValid || !passwordValid) {
            event.preventDefault();
            event.stopPropagation();
            return;
        }

        submitActions[title]();
    };

    function usernameChange(event) {
        if (!hasValidation) {
            return
        }

        const username = event.target.value;
        if (username.length < 3) {
            setUsernameFeedback("Username is too short");
            setUsernameValid(false);
            return;
        }
        let firstLetter = username[0];
        if (!((firstLetter >= 'a' && firstLetter <= 'z') || (firstLetter >= 'A' && firstLetter <= 'Z'))) {
            setUsernameFeedback("Username must start with a letter");
            setUsernameValid(false);
            return;
        }
        for (let i = 1; i < username.length; i++) {
            let letter = username[i];
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
            setUsernameValid(false);
            return
        }

        setUsernameFeedback("");
        setUsernameValid(true);
    }

    function passwordChange(event) {
        if (!hasValidation) {
            return
        }

        const password = event.target.value;
        if (password.length < 6) {
            setPasswordFeedback(defaultPasswordFeedback);
            setPasswordValid(false);
            return;
        }

        let hasLetter = false;
        let hasNumber = false;
        let hasSpecial = false;
        for (let i = 0; i < password.length; i++) {
            let letter = password[i];
            if (letter >= 'a' && letter <= 'z') {
                hasLetter = true;
                continue;
            }
            if (letter >= 'A' && letter <= 'Z') {
                hasLetter = true;
                continue;
            }
            if (letter >= '0' && letter <= '9') {
                hasNumber = true;
                continue;
            }
            hasSpecial = true;
        }
        if (!(hasLetter && hasNumber && hasSpecial)) {
            setPasswordFeedback(defaultPasswordFeedback);
            setPasswordValid(false);
            return;
        }

        setPasswordFeedback("");
        setPasswordValid(true);
    }

    return (
        <li>
            <Modal show={show} onHide={onHide}>
                <Modal.Header closeButton>
                    <Modal.Title>{title}</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <Form onSubmit={handleSubmit}>
                        <Form.Group as={Row} className="mb-3">
                            <InputGroup hasValidation={hasValidation}>
                                <Form.Label column sm="3">User name</Form.Label>
                                <Col sm="8">
                                    <Form.Control type="plaintext" onChange={usernameChange} isValid={hasValidation && usernameValid} isInvalid={hasValidation && !usernameValid} />
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
                                    <Form.Control type="password" onChange={passwordChange} isValid={hasValidation && passwordValid} isInvalid={hasValidation && !passwordValid} />
                                    <Form.Control.Feedback type="invalid">
                                        {passwordFeedback}
                                    </Form.Control.Feedback>
                                </Col>
                            </InputGroup>
                        </Form.Group>
                    </Form>
                    <Alert variant="danger" show={true}>Both user</Alert>
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={onHide}>Close</Button>
                    <Button variant="primary" onClick={handleSubmit}>{title}</Button>
                </Modal.Footer>
            </Modal>
        </li>
    )
}
