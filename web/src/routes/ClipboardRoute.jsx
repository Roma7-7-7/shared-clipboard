import {Col, Container, Row} from "react-bootstrap";
import {apiBaseURL} from "../env.jsx";
import {useEffect, useRef, useState} from "react";
import {useParams} from "react-router-dom";

export default function ClipboardRoute() {
    const [alertMsg, setAlertMsg] = useState("")
    const [content, setContent] = useState("")
    const lastModified = useRef("");

    const params = useParams();
    function refresh() {
        const headers = {}
        if (lastModified.current) {
            headers['If-Modified-Since'] = lastModified
        }
        fetch(apiBaseURL + `/v1/sessions/${params.sessionId}/clipboard`, {
            method: "GET",
            headers: headers,
            credentials: 'include',
            cache: "no-store",
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
                    return response.text();
                }

                return Promise.reject(response.text());
            })
            .then(data => {
                if (alertMsg !== "") {
                    setAlertMsg("")
                }
                setContent(data);
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
        fetch(apiBaseURL + `/v1/sessions/${params.sessionId}/clipboard`, {
            "method": "PUT",
            "headers": {"Content-Type": "text/plain"},
            body: document.getElementById("clipboardText").value,
            credentials: 'include',
        })
            .then(response => {
                if (response.status !== 204) {
                    return Promise.reject(response.text());
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

    const handleChange = (event) => {
        setContent(event.target.value);
    }

    useEffect(() => {
        lastModified.current = "";
        refresh()
        setInterval(refresh, 1000)
    }, [])

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
                        <textarea value={content} onChange={handleChange} className="form-control" id="clipboardText" rows="30"></textarea>
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