# github-data-aggregator

A simple Web API built with Go that aggregates data about my GitHub projects and stores the results in a SQLite database. This project is designed for personal use.

## Overview

- **Data Aggregation:** Gathers and compiles information from various GitHub projects.
- **Web API:** Exposes endpoints to access the aggregated data.
- **SQLite Storage:** Uses a lightweight SQLite database to store project data.
- **Go Implementation:** Written in Go for robust performance and ease of deployment.

## Prerequisites

- Go 1.18 or later
- Git
- SQLite

## Installation

1. **Clone the repository:**

```bash
git clone https://github.com/TeoMatosevic/github-data-aggregator.git
```

2. **Navigate to the project directory:**

```bash
cd github-data-aggregator
```

3. **Build the application:**

```bash
go build -o github-aggregator
```

## Configuration

- **Database:** The application uses a SQLite database (default file: `data.db`) to store aggregated data.
- **Environment Variables:** Adjust API settings and database paths via environment variables if needed.

## Running the Application

1. **Start the API server:**

```bash
./github-aggregator
```

2. **Access the API:**
- The API will typically run on `http://localhost:8080`.
- Use an API client or browser to explore available endpoints.

## API Endpoints

| Method | Endpoint         | Description                                |
| ------ | ---------------- | ------------------------------------------ |
| GET    | `/api/v1/data`   | Retrieve aggregated data for all projects. |
| POST   | `/api/v1/repos`  | Sync data about repositores.               |
| POST   | `/api/v1/urls`   | Sync data about repository contents.       |

## License

This project is intended for personal use only. Feel free to modify and adapt the code for your own purposes.
