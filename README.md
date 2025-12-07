
# Jupyter API Services

This repository contains the backend API services for the Jupyter platform, built with Go. It provides a robust and scalable multi-tenant architecture to manage companies, users, contacts, and handle secure file uploads to AWS S3, alongside a flexible feature-flag system.

## Features

*   **User Management:** Secure user registration, authentication (JWT), and role-based access control.
*   **Company Management:** Multi-tenant support, allowing the creation and management of multiple company accounts with isolated data.
*   **Contact Management:** Functionality to manage contact information associated with companies or users.
*   **Secure File Uploads:** Integration with AWS S3 for secure and efficient file storage, with company-specific bucket configurations.
*   **Feature Flag System:** Dynamic control over application features on a per-company basis.
*   **Role-Based Access Control (RBAC):** Granular permissions based on user roles (e.g., `superadmin`).

## Technologies Used

*   **Go:** The primary programming language for backend development.
*   **Chi:** A lightweight, idiomatic, and composable router for building HTTP services in Go.
*   **GORM:** An excellent ORM library for Go, used for interacting with the MySQL database.
*   **MySQL:** The relational database used for data persistence.
*   **JWT (JSON Web Tokens):** For secure user authentication and authorization.
*   **AWS SDK for Go:** For seamless integration with AWS services, particularly S3 for file storage.
*   **Swagger:** For API documentation and interactive exploration of endpoints.

## Getting Started

To get a copy of the project up and running on your local machine for development and testing purposes, follow these steps.

### Prerequisites

*   Go (version 1.18 or higher recommended)
*   MySQL database instance
*   AWS S3 bucket and credentials (for file upload functionality)
*   `make` (optional, for convenience with makefile commands)

### Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-username/jupyter_api.git
    cd jupyter_api
    ```

2.  **Install Go modules:**
    ```bash
    go mod tidy
    ```

3.  **Database Setup:**
    *   Create a MySQL database.
    *   Update the database connection string in your configuration (e.g., in `internal/config/config.go` or via environment variables). GORM will handle migrations automatically on application start.

4.  **Environment Variables:**
    Set the necessary environment variables for your database connection, JWT secrets, and AWS S3 credentials.
    Example (adapt as needed for your OS):
    ```bash
    export DB_HOST=127.0.0.1
    export DB_PORT=3306
    export DB_USER=root
    export DB_PASSWORD=password
    export DB_NAME=jupyter_api_db
    export JWT_SECRET=your_jwt_secret_key
    export AWS_REGION=your_aws_region
    export AWS_ACCESS_KEY_ID=your_aws_access_key_id
    export AWS_SECRET_ACCESS_KEY=your_aws_secret_access_key
    ```

### Running the Application

To run the API service:

```bash
go run cmd/api/main.go
```

The API will typically start on port `8080` (check `internal/config/config.go` for the exact port).

## API Documentation

The API documentation is generated using Swagger. Once the application is running, you can access the Swagger UI at:

`http://localhost:8080/swagger/index.html`

This interface allows you to view all available endpoints, their request/response schemas, and interact with the API directly.

## Authentication

This API uses JWT (JSON Web Tokens) for authentication.

1.  **Obtain a Token:** Users will need to authenticate (e.g., by logging in with credentials) to receive a JWT.
2.  **Include Token in Requests:** For protected routes, include the JWT in the `Authorization` header of your HTTP requests:
    ```
    Authorization: Bearer <your_jwt_token>
    ```

The system also implements Role-Based Access Control (RBAC), with specific roles (e.g., `superadmin`) having elevated privileges.

## Project Structure

*   `cmd/`: Contains the main application entry points for different services (e.g., `api`, `seed`).
*   `docs/`: API documentation files (Swagger JSON/YAML).
*   `internal/`: Core application logic, organized by domain or concern (e.g., `auth`, `company`, `user`, `uploader`).
*   `privkeys/`: Stores private keys (e.g., for JWT signing).
*   `go.mod`, `go.sum`: Go module definition and dependency checksums.
