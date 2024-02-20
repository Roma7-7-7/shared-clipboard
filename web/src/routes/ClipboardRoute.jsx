import {Col, Container, Row} from "react-bootstrap";
import {useCallback, useEffect, useRef, useState} from "react";
import {useBeforeUnload, useParams} from "react-router-dom";
import {apiBaseURL} from "../env.jsx";
import axios from "axios";

export default function ClipboardRoute() {
    const params = useParams();
    const [alertMsg, setAlertMsg] = useState("")
    const content = useRef("")
    const lastModified = useRef("");
    const timer = useRef()

    function refresh() {
        const headers = {}
        if (lastModified.current) {
            headers['If-Modified-Since'] = lastModified
        }
        axios.get(apiBaseURL + `/v1/sessions/${params.sessionId}/clipboard`, {
            headers: headers,
            withCredentials: true,
        })
            .then(response => {
                if (response.status===204 || response.status === 304) {
                    return Promise.reject(null);
                }
                if (response.status === 200) {
                    if (lastModified.current === response.headers.get('Last-Modified')) {
                        return Promise.reject(null);
                    }
                    lastModified.current = response.headers.get('Last-Modified')
                    return response.data;
                }

                return Promise.reject(response.data);
            })
            .then(data => {
                if (alertMsg !== "") {
                    setAlertMsg("")
                }
                content.current.value = data;
            })
            .catch(error => {
                if (error === null) {
                    return;
                }
                setAlertMsg("Failed to fetch clipboard content. Please try again later.")
                console.error('Error:', error)
            })
    }

    const handleShare = () => {
        axios.put(apiBaseURL + `/v1/sessions/${params.sessionId}/clipboard`, document.getElementById("clipboardText").value, {
            headers: {"Content-Type": "text/plain"},
            withCredentials: true,
        })
            .then(response => {
                if (response.status !== 204) {
                    return Promise.reject(response.data);
                }
                if (alertMsg !== "") {
                    setAlertMsg("")
                }
            })
            .catch(error => {
                setAlertMsg("Failed to share clipboard content. Please try again later.")
                console.error('Error:', error)
            })
    }

    useEffect(() => {
        if (timer.current) {
            return;
        }

        lastModified.current = "";
        refresh()
        timer.current = setInterval(refresh, 1000)

        return () => {
            clearInterval(timer.current)
            timer.current = null
        }
    })

    return (
        <Container>
            <Row>
                <Col className="col-3"/>
                <Col className="text-center">
                    <Row><h2>Clipboard</h2></Row>
                    {alertMsg && <Row className="mt-3">
                        <Col className="text-center">
                            <div className="alert alert-danger" role="alert">
                                {alertMsg}
                            </div>
                        </Col>
                    </Row>}
                    <Row className="mt-3">
                        <textarea ref={content} className="form-control" id="clipboardText" rows="30"></textarea>
                    </Row>
                    <Row className="mt-3">
                        <Col className="col-3" />
                        <Col className="d-grid gap-2">
                            <button onClick={handleShare} className="btn btn-primary">Share</button>
                        </Col>
                        <Col className="col-3" />
                    </Row>
                </Col>
                <Col className="col-3"/>
            </Row>

        </Container>
    )
}