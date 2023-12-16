# scouts-server

Backend Go implementation of [Aaron's Scouts game](https://github.com/AaronLieb/Scouts).

## Package Documentation

Refer to [`package scouts`](https://godoc.org/libdb.so/scouts-server/scouts) for documentation.

## API Documentation

### User API

The server publicly exposes these endpoints:

- `GET /api/v1/users/{id}`: get a user by ID
- `POST /api/v1/users`: create a new user
- `POST /api/v1/login`: log in as a user  

### Games API

The server publicly exposes these endpoints:

- `GET /api/v1/games`: list all games
- `GET /api/v1/games/{id}`: get a game by ID
- `GET /api/v1/games/{id}/events`: subscribe to events for a game using Server-Sent Events
- `POST /api/v1/games`: create a new game
- `POST /api/v1/games/{id}/join`: join a game
- `POST /api/v1/games/{id}/move`: make a move in a game

All the above endpoints require the following headers:

- `Content-Type: application/json`
- `Authorization`: the session token, either with type `Bearer` or `Bot`.
