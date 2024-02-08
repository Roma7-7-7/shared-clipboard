import React from 'react'
import ReactDOM from 'react-dom/client'
import {createBrowserRouter, RouterProvider} from "react-router-dom";
import RootRoute from "./routes/RootRoute.jsx";
import HomeRoute from "./routes/Home.jsx";
import SessionsRoute from "./routes/SessionsRoute.jsx";
import SessionRoute from "./routes/SessionRoute.jsx";

export const router = createBrowserRouter([
    {
        name: 'root', path: '/', element: <RootRoute/>, children: [
            {name: 'home', path: '/', element: <HomeRoute/>},
            {name: 'sessions', path: 'sessions', element: <SessionsRoute/>},
            {name: 'sessions/new', path: 'sessions/new', element: <SessionRoute action="new"/>},
            {name: 'sessions/edit', path: 'sessions/:sessionId/edit', element: <SessionRoute action="edit"/>},
        ]
    },
])

ReactDOM.createRoot(document.getElementById('root')).render(
    <React.StrictMode>
        <RouterProvider router={router}/>
    </React.StrictMode>
)
