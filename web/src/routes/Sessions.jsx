import {useEffect, useState} from "react";
import {Col, Container, Row, Table} from "react-bootstrap";
import {apiBaseURL} from "../env.jsx";

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
                setItems(data.map((session) => <SessionItem name={session["name"]} updatedAt={session["updated_at"]}/>));
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
                <Col></Col>
                <Col xs={5}>
                    <Table striped bordered hover>
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Updated At</th>
                            </tr>
                        </thead>
                        <tbody>
                            {items}
                        </tbody>
                    </Table>
                </Col>
                <Col></Col>
            </Row>
        </Container>
    )
}

function SessionItem({name, updatedAt}) {
    return (
        <tr>
            <td>{name}</td>
            <td width="200px">{formatUpdatedAt(updatedAt)}</td>
        </tr>
    )
}

function formatUpdatedAt(updatedAt) {
    return new Date(updatedAt).toLocaleString()
}