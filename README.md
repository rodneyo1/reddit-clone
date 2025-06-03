# Real-Time Forum

A real-time forum built with Go, SQLite, and vanilla JavaScript. It features user registration, login, post creation, commenting, and private messaging — all in a single-page application (SPA) experience.


## Features

### User Authentication
- Registration with:
  - Nickname, Age, Gender, First Name, Last Name, Email, Password
- Login using Email or Nickname + Password
- Secure password hashing with `bcrypt`
- Session handling via cookies (1 session per user)

### Posts and Comments
- Create and view posts
- Posts have categories
- Comment on posts
- Feed-based display with filters

### Private Messaging (Real-Time Chat)
- WebSocket-powered private chat
- User list showing online/offline status
- Chat sorted by last message or alphabetically
- Scroll to load more messages (pagination with throttling)
- Real-time updates: no page reload needed
- Message format includes sender, timestamp, and content


## Tech Stack

| Layer        | Tech             |
|--------------|------------------|
| **Frontend** | HTML, CSS, JavaScript (No frameworks) |
| **Backend**  | Golang + Gorilla WebSocket |
| **Database** | SQLite           |
| **Auth**     | bcrypt, UUID     |
| **SPA**      | Hash-based routing in JS |
| **Container**| Docker           |


## Setup Instructions

### Prerequisites
- [Go](https://golang.org/dl/)
- SQLite installed locally (or bundled via Go)
- Basic knowledge of HTTP, SQL, and WebSockets


### Local Run (Without Docker)

```bash
# Install dependencies
go get ./...

# Run the app
go run .


##  Testing & Debugging

Run tests with:

```bash
go test ./...
```

### Common Issues

| Problem                  | Solution                                            |
| ------------------------ | --------------------------------------------------- |
| Port already in use      | Kill process using port 8080 or change it in config |
| Login not persisting     | Ensure cookies are being sent correctly             |
| WebSocket not connecting | Confirm backend is listening at correct endpoint    |


## License

This project is licensed under the MIT License.


## Authors

* **Rodney Otieno** – [@rodneyo1](https://github.com/rodneyo1)
* **hanapiko** – [@hanapiko](https://github.com/hanapiko)


## Contributions

We welcome contributions! Follow these steps:

1. Fork the repo
2. Create a branch: `git checkout -b feature/chat`
3. Commit: `git commit -am 'Add chat feature'`
4. Push: `git push origin feature/chat`
5. Submit a pull request

