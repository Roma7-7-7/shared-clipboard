import {useEffect, useState} from "react";
import {Alert, Col, Container, Form, Pagination, Row, Table} from "react-bootstrap";
import {apiBaseURL} from "../env.jsx";
import {Link} from "react-router-dom";
import {Pen, Trash} from "react-bootstrap-icons";
import axios from "axios";

export default function SessionsRoute() {
    const [alertMessage, setAlertMessage] = useState("")
    const [pagination, setPagination] = useState({
        page: 1,
        pageSize: 10,
        totalPages: 1,
        sortBy: "updated_at",
        sortByDesc: true,
    })
    const [sessions, setSessions] = useState({
        items: [],
        totalItems: 0,
    })

    function refresh() {
        const offset = (pagination.page - 1) * pagination.pageSize;
        axios.get(apiBaseURL + `/v1/sessions?sortBy=${pagination.sortBy}&desc=${pagination.sortByDesc}&limit=${pagination.pageSize}&offset=${offset}`, {withCredentials: true})
            .then(response => {
                setSessions({
                    items: response.data.items,
                    totalItems: response.data.totalItems,
                })
                if (alertMessage !== "") {
                    setAlertMessage("")
                }
            }).catch(error => {
            console.log("Error: ", error)
            setAlertMessage("Failed to fetch sessions")
        })
    }

    useEffect(() => {
        refresh()
    }, [pagination])

    const handlePageChange = (page) => {
        setPagination({
            ...pagination,
            page: page,
        })
    }

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
                        <SessionsTable sessions={sessions.items} onDelete={() => refresh()}/>
                    </Row>
                    <Row>
                        <Col/>
                        <Col>
                            <PaginationFooter page={pagination.page}
                                              totalPages={Math.ceil(sessions.totalItems / pagination.pageSize)}
                                              onPageChange={handlePageChange}/>
                        </Col>
                        <Col style={{"display": "block ruby", "margin-left": "auto", "margin-right": 0}}>
                            Page size: <Form.Select style={{width: "75px"}} value={pagination.pageSize}
                                               onChange={(event) => {
                                                   setPagination({
                                                       ...pagination,
                                                       page: 1,
                                                       pageSize: event.target.value,
                                                   })
                                               }}>
                            <option value="5">5</option>
                            <option value="10">10</option>
                            <option value="20">20</option>
                        </Form.Select>
                        </Col>
                    </Row>

                    <Row className="text-center">
                        <Col/>
                        <Col>
                            <Link to="new" className="btn btn-primary">New Session</Link>
                        </Col>
                        <Col/>
                    </Row>
                </Col>
                <Col/>
            </Row>
        </Container>
    )
}

function SessionsTable({sessions, onDelete}) {
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
            {sessions.map((session) => <SessionItem key={session["session_id"]} session={session}
                                                    onDelete={onDelete}/>)}
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

function PaginationFooter({page, totalPages, onPageChange}) {
    if (totalPages <= 1) {
        return (<></>)
    }

    if (totalPages <= 4) {
        const items = []
        for (let i = 1; i <= totalPages; i++) {
            items.push(<Pagination.Item onClick={() => onPageChange(i)} active={page === i}>{i}</Pagination.Item>)
        }
        return (
            <Pagination>
                {items}
            </Pagination>
        )
    }

    return (
        <Pagination>
            <Pagination.First onClick={() => onPageChange(1)} disabled={page === 1}/>
            <Pagination.Prev onClick={() => onPageChange(page - 1)} disabled={page === 1}/>

            <Pagination.Item active>{page}</Pagination.Item>

            <Pagination.Next onClick={() => onPageChange(page + 1)} disabled={page === totalPages}/>
            <Pagination.Last onClick={() => onPageChange(totalPages)} disabled={page === totalPages}/>
        </Pagination>
    );
}

function formatLastUsedAt(updatedAt) {
    return new Date(updatedAt).toLocaleString()
}