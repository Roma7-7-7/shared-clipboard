import  {Link} from "react-router-dom";
import {Container, Navbar as NavbarB, Nav, Button} from "react-bootstrap";

export default function Navbar({userInfo, onSignInClicked, onSignUpClicked, onSignOutClicked}) {
    return (
        <>
            <NavbarB expand="lg" bg="dark" variant="dark" className="mb-5">
                <Container>
                    <Link to="/" className="navbar-brand">Clipboard share</Link>
                    <NavbarB.Collapse>
                        <Nav className="me-auto"></Nav>
                        {userInfo === null && (
                            <Nav>
                                <Button variant="outline-primary me-2" onClick={onSignInClicked}>Sign In</Button>
                                <Button variant="outline-secondary me-2" onClick={onSignUpClicked}>Sign Up</Button>
                            </Nav>
                        )}
                        {(userInfo !== null) && (
                            <Nav>
                                <Button variant="outline-primary me-2" onClick={onSignOutClicked}>Sign Out</Button>
                            </Nav>
                        )}
                    </NavbarB.Collapse>
                </Container>
            </NavbarB>
        </>
    )
}
