import {Container, Navbar as NavbarB, Nav, Button} from "react-bootstrap";

export default function Navbar({userInfo, onSignInClicked, onSignUpClicked, onSignOutClicked}) {
    let rightNav = (<>
        <Nav>
            <Button variant="outline-primary me-2" onClick={onSignInClicked}>Sign In</Button>
            <Button variant="outline-secondary me-2" onClick={onSignUpClicked}>Sign Up</Button>
        </Nav>
    </>)

    if (userInfo !== null) {
        rightNav = (<>
            <Nav>
                <Button variant="outline-primary me-2" onClick={onSignOutClicked}>Sign Out</Button>
            </Nav>
        </>)
    }

    return (
        <>
            <NavbarB expand="lg" bg="dark" variant="dark">
                <Container>
                    <NavbarB.Brand href="#home">Clipboard share</NavbarB.Brand>
                    <NavbarB.Collapse>
                        <Nav className="me-auto"></Nav>
                        {rightNav}
                    </NavbarB.Collapse>
                </Container>
            </NavbarB>
        </>
    )
}
