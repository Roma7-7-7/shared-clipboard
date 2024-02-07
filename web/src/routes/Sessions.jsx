import {useEffect, useState} from "react";
import {Col, Container, Row, Table} from "react-bootstrap";
import {apiBaseURL} from "../env.jsx";
import {Link} from "react-router-dom";
import {Pen, Trash} from "react-bootstrap-icons";

export default function Sessions() {
    const [items, setItems] = useState([])

    useEffect(() => {
        fetch(apiBaseURL + '/v1/sessions', {
            "method": "GET",
            "headers": {"Content-Type": "application/json"},
            credentials: 'include',
        })
            .then(response => {
                if (response.status === 200) {
                    return response.json();
                }
                return Promise.reject(response.status);
            })
            .then(data => {
                setItems(data.map((session) => <SessionItem key={session["session_id"]} session={session} />));
            })
            .catch(error => {
                console.error('Error:', error)
            })

    }, [])

    return (
        <Container>
            <Row className="mb-5">
                <Col className="text-center">
                    <h2>Sessions</h2>
                </Col>
            </Row>
            <Row>
                <Col />
                <Col xs="5">
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
                </Col>
                <Col />
            </Row>
            <Row>
                <Col className="text-center">
                    <Link to="new" className="btn btn-primary">New Session</Link>
                </Col>
            </Row>
        </Container>
    )
}

function SessionItem({session}) {
    return (
        <tr>
            <td>{session['name']}</td>
            <td width="200px">{formatLastUsedAt(session['updated_at'])}</td>
            <td className="text-center">
                <Link to={`${session['session_id']}/edit`} className="btn btn-link">
                    <Pen />
                </Link>
            </td>
            <td className="text-center">
                <a href="#" className="btn btn-link">
                    <Trash />
                </a>
            </td>
        </tr>
    )
}

function formatLastUsedAt(updatedAt) {
    return new Date(updatedAt).toLocaleString()
}