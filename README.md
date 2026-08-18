# Interseguro Coding Challenge — QR (Go) + Estadísticas (Node.js)

Dos APIs REST que se comunican por HTTP:

- **go-api** (Go + Fiber): recibe una matriz rectangular y devuelve su **factorización QR**, calculada mediante **rotaciones de Givens**.
- **node-api** (Node.js + Express): recibe las matrices **Q** y **R**, y calcula estadísticas (máximo, mínimo, promedio, suma total, y si alguna matriz es diagonal).

El cliente llama **una sola vez** a `go-api`; internamente `go-api` llama a `node-api` y devuelve una respuesta combinada.

```
Cliente ──POST /api/v1/matrix/qr──▶ go-api ──POST /api/v1/stats──▶ node-api
                                       │ (QR vía rotaciones                │ (max/min/avg/sum,
                                       │  de Givens)                       │  diagonal check)
                                       ◀───────────────────────────────────┘
Cliente ◀── { matrices: {Q,R}, statistics, meta } ──┘
```

---

## Decisiones de diseño (para sustentar en la entrevista)

### 1. "Rotación de la matriz" = rotaciones de Givens para calcular QR

El enunciado original menciona la "rotación de la matriz" **tres veces**, siempre junto al requisito de factorización QR — nunca como una operación separada. El bullet de "Funcionalidad requerida" dice textualmente:

> "Implementar la lógica para realizar la rotación de la matriz **y** la operación adicional de manera eficiente y correcta"

dentro del mismo punto cuyo único entregable explícito es "devuelva la factorización QR". No hay ninguna otra mención de qué debería hacer una "rotación" de forma aislada.

La lectura que reconcilia las tres menciones sin contradicción es que **"rotación"** nombra el **algoritmo** para calcular QR: las **rotaciones de Givens**, el método clásico de álgebra lineal que anula las entradas sub-diagonales de una matriz una por una mediante rotaciones 2x2. Por eso `go-api` implementa QR específicamente con rotaciones de Givens (no Householder, no Gram-Schmidt) — ver [`go-api/internal/matrix/qr.go`](go-api/internal/matrix/qr.go).

### 2. QR completo (Q siempre cuadrada), no reducido

`Q` es siempre `m x m` ortogonal y `R` es `m x n` triangular superior (trapezoidal si `m ≠ n`), incluso si la entrada es rectangular. Se eligió la forma **completa** sobre la reducida/económica para que `Q` sea siempre cuadrada y ortogonal sin importar la forma de la entrada — esto simplifica el chequeo de "matriz diagonal" en `node-api`, que solo tiene sentido para matrices cuadradas.

### 3. Contrato de datos: `{ "Q": [...], "R": [...] }`, no un array anónimo

`go-api` envía a `node-api` las matrices **nombradas**, no como un array posicional. Esto le permite a `node-api` reportar diagonalidad por nombre (`perMatrix.Q.isDiagonal`) y hace el endpoint de `node-api` reusable con cualquier conjunto de matrices nombradas, no solo Q/R.

### 4. Semántica de "matriz diagonal"

- Solo tiene sentido para matrices **cuadradas**: una matriz no cuadrada nunca es diagonal (por definición, no es un error).
- Se usa una tolerancia `epsilon = 1e-9` en vez de igualdad exacta, porque QR se calcula en punto flotante y una entrada "matemáticamente cero" puede llegar como `1e-16` por ruido numérico.
- Una matriz `1x1` es trivialmente diagonal (no tiene entradas fuera de la diagonal).
- Una matriz `0x0` se considera **no diagonal** (simplificación deliberada por previsibilidad; en la práctica es inalcanzable, ya que `go-api` siempre valida que la entrada sea no vacía antes de calcular QR).
- El enunciado pide literalmente "verificar si **alguna** matriz es diagonal": `node-api` responde eso directamente en `statistics.overall.anyDiagonal` (booleano) y `statistics.overall.diagonalMatrices` (lista de nombres), además del detalle por matriz en `perMatrix.<nombre>.isDiagonal`.

### 5. JWT como sustituto de un IdP real

No hay una base de usuarios real en el alcance de este challenge. `go-api` expone `POST /api/v1/auth/token`, que intercambia una **pre-shared key** (`X-API-Key`, configurada por variable de entorno) por un JWT de corta duración. Esto demuestra el flujo de autenticación JWT de punta a punta sin construir un sistema de identidad completo — en producción, este endpoint se reemplazaría por un IdP real (Auth0, Cognito, un servicio de usuarios propio, etc.).

Para la llamada interna `go-api → node-api`, `go-api` firma un **token de servicio** de corta vida (60s) con el mismo secreto compartido (`JWT_SECRET`), así `node-api` tampoco queda abierto sin autenticación dentro de la red interna.

### 6. `[][]float64` en vez de un slice plano + dimensiones

El formato de la API ya es "array de arrays", así que representar la matriz como `[][]float64` evita una capa de conversión en cada límite HTTP. Para matrices muy grandes, un `[]float64` plano con dimensiones explícitas sería más eficiente en localidad de caché — pero a la escala de este challenge, la legibilidad directa gana.

### 7. Sin reintentos en la llamada Go → Node

Decisión de alcance explícita, no un descuido: si `node-api` no responde a tiempo o falla, `go-api` devuelve un error mapeado (504/502) en vez de reintentar. Se documenta como mejora futura razonable (con backoff exponencial y límite de reintentos).

---

## Estructura del repositorio

```
interseguro/
├── docker-compose.yml
├── .env.example
├── scripts/integration-check.sh
├── go-api/                  # Fiber: POST /api/v1/matrix/qr, POST /api/v1/auth/token
│   ├── cmd/api/main.go
│   └── internal/{config,dto,matrix,middleware,client,handlers}/
├── node-api/                 # Express: POST /api/v1/stats
│   └── src/{config,routes,controllers,services,middleware,utils}/
└── frontend/                 # HTML/JS estático, sin build step
```

---

## Contratos de API

### go-api (`http://localhost:8080`)

#### `GET /health`
Sin autenticación.
```bash
curl http://localhost:8080/health
```
```json
{ "status": "ok", "service": "go-api", "time": "2026-08-18T10:00:00Z" }
```

#### `POST /api/v1/auth/token`
```bash
curl -X POST http://localhost:8080/api/v1/auth/token \
  -H "X-API-Key: change-me-preshared-key"
```
200:
```json
{ "token": "<jwt>", "tokenType": "Bearer", "expiresIn": 3600 }
```
401 (`X-API-Key` inválida o ausente):
```json
{ "error": { "code": "INVALID_API_KEY", "message": "invalid or missing X-API-Key header" } }
```

#### `POST /api/v1/matrix/qr`
```bash
TOKEN="<jwt del paso anterior>"
curl -X POST http://localhost:8080/api/v1/matrix/qr \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"matrix": [[12, -51, 4], [6, 167, -68], [-4, 24, -41]]}'
```
200:
```json
{
  "matrices": { "Q": [[...]], "R": [[...]] },
  "statistics": {
    "perMatrix": {
      "Q": { "rows": 3, "cols": 3, "max": 1, "min": -0.86, "average": 0.11, "sum": 1.0, "isDiagonal": false },
      "R": { "rows": 3, "cols": 3, "max": 176.25, "min": 0, "average": 15.3, "sum": 137.7, "isDiagonal": false }
    },
    "overall": { "count": 18, "max": 176.25, "min": -0.86, "average": 7.7, "sum": 138.7, "anyDiagonal": false, "diagonalMatrices": [] }
  },
  "meta": { "inputRows": 3, "inputCols": 3, "computedAtMs": 1 }
}
```
Errores: `400 INVALID_JSON`, `400 INVALID_INPUT` (matriz vacía, no rectangular, o con `NaN`/`Inf`), `401 UNAUTHORIZED` (token ausente/expirado/inválido), `502 UPSTREAM_UNAVAILABLE`/`UPSTREAM_ERROR`/`UPSTREAM_INVALID_RESPONSE`, `504 UPSTREAM_TIMEOUT` (si `node-api` falla o no responde a tiempo).

### node-api (`http://localhost:4000`) — también invocable de forma independiente

#### `GET /health` — sin autenticación.

#### `POST /api/v1/stats`
```bash
curl -X POST http://localhost:4000/api/v1/stats \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"matrices": {"Q": [[1,0],[0,1]], "R": [[2,3],[0,4]]}}'
```
200:
```json
{
  "perMatrix": {
    "Q": { "rows": 2, "cols": 2, "max": 1, "min": 0, "average": 0.5, "sum": 2, "isDiagonal": true },
    "R": { "rows": 2, "cols": 2, "max": 4, "min": 0, "average": 2.25, "sum": 9, "isDiagonal": false }
  },
  "overall": { "count": 8, "max": 4, "min": 0, "average": 1.375, "sum": 11, "anyDiagonal": true, "diagonalMatrices": ["Q"] }
}
```
Errores: `400 INVALID_INPUT` (falta `matrices`, matriz no rectangular, valores no numéricos), `400 INVALID_JSON`, `401 UNAUTHORIZED`.

---

## Cómo correrlo localmente

```bash
cp .env.example .env
# editar .env: JWT_SECRET y GO_API_PRESHARED_KEY con valores propios

docker compose up --build
```

- go-api: http://localhost:8080
- node-api: http://localhost:4000
- frontend: http://localhost:8081

> **Nota sobre `go-api/go.sum`:** este repo no incluye `go.sum` versionado porque fue creado sin acceso a un toolchain de Go local para generarlo (ver más abajo). El `Dockerfile` corre `go mod tidy` automáticamente durante el build (requiere acceso a internet en ese paso). Para desarrollo local fuera de Docker, correr `go mod tidy` una vez dentro de `go-api/` antes de `go run`/`go test`.

## Tests

```bash
# go-api
cd go-api && go mod tidy && go test ./... -v -cover

# node-api (sin dependencias externas de test runner: usa node:test)
cd node-api && npm install && npm test

# integración end-to-end contra el stack levantado con docker compose
GO_API_PRESHARED_KEY=<tu-key> ./scripts/integration-check.sh
```

`node-api` fue efectivamente instalado y probado en este entorno (24/24 tests verdes). `go-api` fue escrito y revisado cuidadosamente pero **no pudo compilarse/testearse en este entorno** por no contar con el toolchain de Go instalado — correr `go test ./...` es el primer paso recomendado antes de confiar en el binario.

---

## Despliegue en la nube

No se aprovisionó infraestructura real (sin cuenta/credenciales de un proveedor disponibles en este entorno). El diseño ya es "cloud-ready":

- Contenedores sin estado, configurables por variables de entorno.
- `/health` en ambas APIs, listo para probes de Cloud Run / ECS Fargate / Kubernetes.
- Sugerencia concreta: `go-api` y `node-api` como servicios en **Cloud Run** o **ECS Fargate** (autoscaling, sin gestión de servidores); `frontend` como sitio estático (Cloud Storage + CDN, S3 + CloudFront, o el mismo patrón de contenedor nginx); `JWT_SECRET`/`GO_API_PRESHARED_KEY` en un secret manager gestionado (Secret Manager / AWS Secrets Manager) en vez de `.env`.

---

## Supuestos y trade-offs (resumen)

- "Rotación de la matriz" interpretada como rotaciones de Givens para QR (no una operación de rotación separada).
- QR completo (Q cuadrada `m x m`) en vez de reducido.
- Diagonalidad: solo matrices cuadradas, tolerancia `1e-9`, `0x0` → no diagonal.
- JWT emitido a partir de una pre-shared key, no de un IdP real.
- Sin reintentos en la llamada `go-api → node-api`.
- `[][]float64` en vez de slice plano + dimensiones.
- Sin despliegue real en la nube por falta de credenciales; diseño preparado para ello.
- `go-api` no fue compilado/testeado localmente por falta de Go instalado en el entorno de desarrollo; se dejó `go mod tidy` corriendo automáticamente en el `Dockerfile` como red de seguridad.
