# FeedBook Back

Backend minimo para la POC de login seguro de FeedBook.

## Endpoint

- `POST /login`

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
go run .
```

Levanta en `http://127.0.0.1:8080`.

Para dispositivo fisico con `adb reverse`:

```bash
adb reverse tcp:8080 tcp:8080
```
