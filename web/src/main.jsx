import React from 'react'
import ReactDOM from 'react-dom/client'
import {createBrowserRouter, RouterProvider} from "react-router-dom";
import Root from "./routes/Root.jsx";
import Home from "./routes/Home.jsx";
import Sessions from "./routes/Sessions.jsx";

export const router = createBrowserRouter([
    {path: '/', element: <Root />, children: [
        {path: '/', element: <Home />},
        {path: '/sessions', element: <Sessions />},
    ]},
])

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>
)
