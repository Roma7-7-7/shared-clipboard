import {useState, useEffect} from "react";
import {useParams} from "react-router-dom";
import {apiBaseURL} from "../env.jsx";
import {Button, Col, Container, Form, InputGroup, Row} from "react-bootstrap";

export default function SessionRoute({action}) {
    const [alertMsg, setAlertMsg] = useState("")
    const [session, setSession] = useState(null)

    const params = useParams()
    let handleSubmit;

    switch (action) {
        case "new":
            handleSubmit = function (event) {
                event.preventDefault()
                setAlertMsg("")
                if (session.name.trim() === "") {
                    setAlertMsg("User name is required")
                    return
                }

                fetch(apiBaseURL + "/v1/sessions", {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json"
                    },
                    body: JSON.stringify({name: session.name}),
                    credentials: "include",
                }).then(response => {
                    if (response.status === 201) {
                        return response.json()
                    } else {
                        setAlertMsg("Failed to create session")
                        return Promise.reject("Failed to create session")
                    }
                }).then(resp => {
                    window.location.href = "/sessions/" + resp["session_id"] + "/clipboard"
                }).catch(error => {
                    console.log("Error: ", error)
                    setAlertMsg("Failed to create session")
                })
            }
            break
        case "edit":
            useEffect(() => {
                fetch(apiBaseURL + "/v1/sessions/" + params.sessionId, {
                    method: "GET",
                    headers: {
                        "Content-Type": "application/json"
                    },
                    credentials: "include",
                }).then(response => {
                    if (response.status === 200) {
                        return response.json()
                    } else {
                        setAlertMsg("Failed to fetch session")
                        return Promise.reject("Failed to fetch session")
                    }
                }).then(resp => {
                    setSession({
                        sessionId: resp["session_id"],
                        name: resp["name"],
                    })
                }).catch(error => {
                    console.log("Error: ", error)
                    setAlertMsg("Failed to fetch session")
                })
            }, [])

            handleSubmit = function (event) {
                event.preventDefault()
                setAlertMsg("")
                if (session.name.trim() === "") {
                    setAlertMsg("User name is required")
                    return
                }

                fetch(apiBaseURL + "/v1/sessions/" + session.sessionId, {
                    method: "PUT",
                    headers: {
                        "Content-Type": "application/json"
                    },
                    body: JSON.stringify({name: session.name}),
                    credentials: "include",
                }).then(response => {
                    if (response.status === 200) {
                        return response.json()
                    } else {
                        setAlertMsg("Failed to update session")
                        return Promise.reject("Failed to update session")
                    }
                }).then(resp => {
                    window.location.href = "/sessions/" + resp["session_id"] + "/clipboard"
                }).catch(error => {
                    console.log("Error: ", error)
                    setAlertMsg("Failed to update session")
                })
            }
            break
        default:
            throw new Error("Invalid path")
    }

    return (
        <Container>
            <Row className="text-center">
                <h2>Create Session</h2>
            </Row>
            <Row>
                <Col/>
                <Col>
                    <Form.Group as={Row} className="mt-3">
                        <InputGroup hasValidation>
                            <Form.Label column>Name:</Form.Label>
                            <Col sm="8">
                                <Form.Control type="plaintext" required
                                              value={session !== null && session.name !== null ? session.name : ""}
                                              onChange={(event) => setSession({
                                                  ...session,
                                                  name: event.target.value
                                              })}
                                              isValid={session !== null && session.name !== null && session.name.trim() !== ""}
                                              isInvalid={session !== null && session.name !== null && session.name.trim() === ""}
                                />
                                <Form.Control.Feedback type="invalid">
                                    User name is required
                                </Form.Control.Feedback>
                            </Col>
                        </InputGroup>
                    </Form.Group>
                </Col>
                <Col/>
            </Row>
            {alertMsg && <Row className="text-center mt-3">
                <Col/>
                <Col>
                    <div className="alert alert-danger" role="alert">
                        {alertMsg}
                    </div>
                </Col>
                <Col/>
            </Row>
            }
            <Row className="mt-3">
                <Col/>
                <Col className="text-center">
                    <Button disabled={session !== null && session.name !== null && session.name.trim() === ""}
                            onClick={handleSubmit}>
                        Submit
                    </Button>
                </Col>
                <Col/>
            </Row>
        </Container>
    )
}