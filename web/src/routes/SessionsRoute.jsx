import {useEffect, useState} from "react";
import {Alert, Col, Container, Pagination, Row, Table} from "react-bootstrap";
import {apiBaseURL} from "../env.jsx";
import {Link} from "react-router-dom";
import {Pen, Trash} from "react-bootstrap-icons";
import axios from "axios";

export default function SessionsRoute() {
    const [alertMessage, setAlertMessage] = useState("")
    const [pagination, setPagination] = useState({
        page: 1,
        pageSize: 5,
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
                    <PaginationFooter page={pagination.page}
                                      totalPages={Math.ceil(sessions.totalItems / pagination.pageSize)}
                                      onPageChange={handlePageChange}/>
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
            <Row className="text-center">
                <Pagination>
                    {items}
                </Pagination>
            </Row>
        )
    }

    const items = []
    items.push(<Pagination.Item onClick={() => onPageChange(1)} disabled={page === 1}>1</Pagination.Item>)
    switch (page) {
        case 1:
            items.push(<Pagination.Item onClick={() => onPageChange(2)} active>{2}</Pagination.Item>)
            break
        case 2:
            items.push(<Pagination.Item onClick={() => onPageChange(page)} active>{page}</Pagination.Item>)
            items.push(<Pagination.Item onClick={() => onPageChange(page + 1)}>{page + 1}</Pagination.Item>)
            break
    }
    items.push(<Pagination.Ellipsis/>)

    switch (page) {
        case totalPages:
            items.push(<Pagination.Item onClick={() => onPageChange(page - 1)}>{page - 1}</Pagination.Item>)
            items.push(<Pagination.Item onClick={() => onPageChange(page)} active>{page}</Pagination.Item>)
            break
        case totalPages - 1:
            items.push(<Pagination.Item onClick={() => onPageChange(page - 1)}>{page - 1}</Pagination.Item>)
            items.push(<Pagination.Item onClick={() => onPageChange(page)} active>{page}</Pagination.Item>)
            items.push(<Pagination.Item onClick={() => onPageChange(page + 1)}>{page + 1}</Pagination.Item>)
            break
    }

    items.push(<Pagination.Item onClick={() => onPageChange(totalPages - 1)}>{totalPages - 1}</Pagination.Item>)


    return (
        <Pagination>
            {items}
        </Pagination>
    )

}

function formatLastUsedAt(updatedAt) {
    return new Date(updatedAt).toLocaleString()
}