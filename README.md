# Harmony

**Harmony** is a chat room application inspired by platforms like Discord. It allows users to create servers that contain both vocal and text chat channels, which can be shared and utilized by others in an intuitive and collaborative environment.

## Scope

The primary goal of this project is to showcase the technical skills and technologies utilized in its development. It highlights backend, frontend, and architectural expertise, serving as a portfolio piece for future opportunities.

## Features

- **Server Management**: Users can create servers with customizable names and icons.
- **Text and Vocal Channels**: Servers can host text and vocal chat channels to facilitate communication.
- **User Roles and Permissions**: Role-based access control ensures secure and organized interaction.
- **Real-time Messaging**: Live chat powered by WebSockets for seamless communication.
- **Authentication**: Secure user registration and login via JWT.

## Technologies

- **Backend**: Built with Go for scalable and performant server-side logic.
- **Frontend**: React with TypeScript for a modern, type-safe, and responsive user interface.
- **Database**: MongoDB for persistent data storage.
- **Real-time Communication**: WebSocket or WebRTC for instant interaction.

## Project Management

This project is actively managed on GitHub. You can view the progress, tasks, and milestones in the dedicated [project board](https://github.com/users/ImCispa/projects/1).

## Installation and Setup

1. Clone the repository
   ```bash
   git clone https://github.com/ImCispa/harmony.git
   cd harmony
   ```
2. Install dependencies for the backend and frontend
    - Backend
        ```bash
        go mod download
        ```
    - Frontend
        ```bash
        cd frontend
        npm install
        ```
3.  Configure environment variables
    - Add `.env` files for both backend and frontend with the required settings (e.g., database connection, API keys).
4.  Run the development servers
    - Backend
        ```bash
        go run main.go
        ```
    - Frontend
        ```bash
        npm start
        ```
## Contributing
Contributions are welcome! Please follow these steps:

1.  Fork the repository
2.  Create a feature branch
    ```bash
    git checkout -b feature/your-feature-name
    ```
3.  Commit your changes and open a pull request.

## License
This project is released under the MIT License.

## Contact
If you have any questions or suggestions, feel free to reach out via the repository's Issues.
