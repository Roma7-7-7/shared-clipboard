import {useEffect, useState} from "react";
import {Alert, Col, Container, Row, Table} from "react-bootstrap";
import {apiBaseURL} from "../env.jsx";
import {Link} from "react-router-dom";
import {Pen, Trash} from "react-bootstrap-icons";
import axios from "axios";

export default function SessionsRoute() {
    const [alertMessage, setAlertMessage] = useState("")
    return (
        <Container>
            <Row>
                <Col/>
                <Col xs="5">
                    <Row className="mb-5 text-center">
                        <h2>Sessions</h2>
                    </Row>
                    {alertMessage && <Row className="mb-3">
                        <Alert variant="danger">
                            {alertMessage}
                        </Alert>
                    </Row>}
                    <Row>
                        <SessionsTable onSuccess={() => {setAlertMessage("")}} onError={(msg) => {setAlertMessage(msg)}} />
                    </Row>
                    <Row className="text-center">
                        <Col />
                        <Col>
                            <Link to="new" className="btn btn-primary">New Session</Link>
                        </Col>
                        <Col />
                    </Row>
                </Col>
                <Col/>
            </Row>
        </Container>
    )
}

function SessionsTable({onSuccess, onError}) {
    const [items, setItems] = useState([])

    function refresh() {
        axios.get(apiBaseURL + '/v1/sessions', {withCredentials: true})
            .then(response => {
                setItems(response.data.map((session) => <SessionItem key={session["session_id"]} session={session} onDelete={() => refresh()} />));
                onSuccess()
            }).catch(error => {
            console.log("Error: ", error)
            onError("Failed to fetch sessions")
        })
    }

    useEffect(() => {
        refresh()
    }, [])

    return (
        <Table striped bordered hover>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Last used at</th>
                    <th className="text-center">Edit</th>
                    <th className="text-center">Delete</th>
                </tr>
            </thead>
            <tbody>
                {items}
            </tbody>
        </Table>
    )

}

function SessionItem({session, onDelete}) {
    function handleDelete(event, sessionID) {
        event.preventDefault()
        axios.delete(apiBaseURL + '/v1/sessions/' + sessionID, {withCredentials: true})
            .then(response => {
                onDelete()
            })
            .catch(error => {
                console.log("Error: ", error)
            })
    }

    return (
        <tr>
            <td><Link to={`${session['session_id']}/clipboard`}>{session['name']}</Link></td>
            <td width="200px">{formatLastUsedAt(session['updated_at_millis'])}</td>
            <td className="text-center">
                <Link to={`${session['session_id']}/edit`} className="btn btn-link">
                    <Pen/>
                </Link>
            </td>
            <td className="text-center">
                <a href="#" className="btn btn-link" onClick={(event) => handleDelete(event, session['session_id'])}>
                    <Trash/>
                </a>
            </td>
        </tr>
    )
}

function formatLastUsedAt(updatedAt) {
    return new Date(updatedAt).toLocaleString()
}