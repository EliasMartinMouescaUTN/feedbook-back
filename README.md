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
- `POST /api/push/register`
- `POST /api/push/send`

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

## Push notifications via Firebase

La app registra su token FCM con:

```http
POST /api/push/register
Content-Type: application/json

{"token":"<fcm-token>","platform":"android"}
```

Para enviar una notificacion push desde el backend:

```http
POST /api/push/send
Content-Type: application/json

{"title":"FeedBook","body":"Nueva actividad en tu biblioteca"}
```

Si se incluye `token`, se envia solo a ese token; si no, se envia a todos los
tokens registrados desde que el backend esta corriendo.

Configuracion requerida:

- Android: descargar `google-services.json` desde Firebase Console para el
  package `com.example.feedbook` y ubicarlo en `app/google-services.json`.
- Backend: crear una service account en Firebase/Google Cloud y levantar el
  backend con el archivo en `back/firebase-service-account.json`.
  Tambien se puede usar `GOOGLE_APPLICATION_CREDENTIALS=/ruta/service-account.json`
  o `FIREBASE_CREDENTIALS_FILE=/ruta/service-account.json`.
  El project ID por defecto es `feedbook-9132b`; se puede pisar con
  `FIREBASE_PROJECT_ID=<project-id>`.

Script local:

```bash
python3 scripts/send_push.py --title "FeedBook" --body "Nueva actividad"
python3 scripts/send_push.py --token "<fcm-token>" --title "FeedBook" --body "Prueba"
```

Para dispositivo fisico con `adb reverse`:

```bash
adb reverse tcp:8080 tcp:8080
```
