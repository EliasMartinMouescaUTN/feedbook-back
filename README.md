# FeedBook Back

Backend minimo para FeedBook. Expone login y contenido hardcodeado por REST
para que la app Android consuma los datos mockeados desde un servicio real.

## Endpoint

- `POST /login`
- `GET /api/books`
- `GET /api/books/{id}`
- `GET /api/books/{id}/progress`
- `GET /api/books/{id}/reviews`
- `GET /api/explore/users`
- `GET /api/authors`
- `GET /api/authors/{id}`
- `POST /api/authors/{id}/follow-toggle`
- `GET /api/home`
- `GET /api/library/me`
- `GET /api/profile/me`
- `PUT /api/profile/me`
- `GET /api/profile/me/preview`
- `GET /api/profile/public`
- `GET /api/stats`
- `GET /api/notifications`

Body JSON:

```json
{
  "username": "demo",
  "password": "demo",
  "secure_login": false
}
```

Tambien acepta `user` en lugar de `username`.

Regla actual:

- Si `username` y `password` son exactamente la misma string, devuelve `200` con un JWT.
- El JWT incluye `secure_login` y `exp`.
- Si no coinciden, devuelve `401`.

## Run

```bash
go test ./...
go run .
```

Levanta en `http://127.0.0.1:8080` por defecto.
Se puede cambiar con `FEEDBOOK_ADDR`.

Para dispositivo fisico con `adb reverse`:

```bash
adb reverse tcp:8080 tcp:8080
```
