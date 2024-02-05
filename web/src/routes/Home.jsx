import {Container, Row} from "react-bootstrap";
import {Link, useOutletContext} from "react-router-dom";

export default function Home() {
    const context = useOutletContext()
    const userInfo = context.userInfo
    const onSignInClicked = context.onSignInClicked
    const onSignUpClicked = context.onSignUpClicked

    return (
        <Container>
            {userInfo === null && (
                <Row>
                    <p className="fs-5">The purpose of this service is to share clipboard content across multiple hosts.</p>
                    <p>Please <a href="#" onClick={onSignInClicked}>sign in</a> or <a href="#" onClick={onSignUpClicked}>sign up</a> to start using it.</p>
                </Row>
            )}
            {userInfo !== null && (<>
                <Row>
                    <p className="fs-5">You are signed in as <strong>{userInfo["name"]}</strong>.</p>
                </Row>
                <Row>
                    <p className="fs-5">Please navigate to the <Link to="/sessions">sessions</Link> page to start using the service.</p>
                </Row>
            </>)}
        </Container>
    )
}