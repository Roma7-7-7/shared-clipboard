import React from 'react'
import ReactDOM from 'react-dom/client'
import {createBrowserRouter, RouterProvider} from "react-router-dom";
import Root from "./routes/Root.jsx";
import Home from "./routes/Home.jsx";
import Sessions from "./routes/Sessions.jsx";
import Session from "./routes/Session.jsx";

export const router = createBrowserRouter([
    {name: 'root', path: '/', element: <Root />, children: [
        {name: 'home', path: '/', element: <Home />},
        {name: 'sessions', path: 'sessions', element: <Sessions />},
        {name: 'sessions/new', path: 'sessions/new', element: <Session action="new" />},
        {name: 'sessions/edit', path: 'sessions/:sessionId/edit', element: <Session action="edit" />},
    ]},
])

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>
)
