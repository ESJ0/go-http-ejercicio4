# 🦸 DC Characters API

API REST en Go que expone información de personajes del universo DC Comics, inspirada en la estructura de [SuperHero API](https://superheroapi.com/).  
Construida únicamente con la librería estándar de Go — sin frameworks externos.

---

## Tema

Los datos representan personajes icónicos de DC Comics: héroes, villanos y antihéroes. Cada personaje incluye estadísticas de poder, apariencia física, biografía y equipo al que pertenece.

---

## Estructura del proyecto

```
.
├── main.go
├── data/
│   └── characters.json
├── Dockerfile
└── docker-compose.yml
```

---

## Ejecutar el servidor

### Local (requiere Go 1.22+)
```bash
go run main.go
```

### Docker
```bash
docker compose up --build
```

El servidor corre en el puerto **24585**.

---

## Endpoints

### `GET /api/ping`
Verificación de salud del servidor.

```bash
curl http://localhost:24585/api/ping
```

**Respuesta:**
```json
{ "mensaje": "pong" }
```

---

### `GET /api/characters`
Retorna todos los personajes registrados.

```bash
curl http://localhost:24585/api/characters
```

---

### `GET /api/characters?id=1`
Busca un personaje por su ID usando query parameter.

```bash
curl "http://localhost:24585/api/characters?id=1"
```

---

### `GET /api/characters/{id}`
Busca un personaje por su ID usando path parameter.

```bash
curl http://localhost:24585/api/characters/1
```

---

### Filtros combinados

Se pueden combinar múltiples parámetros en una sola petición:

| Parámetro         | Tipo   | Descripción                                        |
|-------------------|--------|----------------------------------------------------|
| `name`            | string | Busca por nombre (búsqueda parcial, sin distinción mayúsculas) |
| `alignment`       | string | Filtrar por alineación: `good`, `bad`, `neutral`   |
| `team`            | string | Filtrar por equipo (búsqueda parcial)               |
| `gender`          | string | Filtrar por género: `Male`, `Female`                |
| `min_intelligence`| int    | Inteligencia mínima (0–100)                         |
| `min_combat`      | int    | Combate mínimo (0–100)                              |

**Ejemplos:**
```bash
# Todos los villanos
curl "http://localhost:24585/api/characters?alignment=bad"

# Héroes de la Justice League con combate >= 90
curl "http://localhost:24585/api/characters?alignment=good&team=Justice+League&min_combat=90"

# Personajes femeninos
curl "http://localhost:24585/api/characters?gender=Female"

# Buscar por nombre
curl "http://localhost:24585/api/characters?name=bat"
```

---

### `POST /api/characters`
Crea un nuevo personaje. Persiste los cambios en `data/characters.json`.

```bash
curl -X POST http://localhost:24585/api/characters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Shazam",
    "real_name": "Billy Batson",
    "alignment": "good",
    "publisher": "DC Comics",
    "powerstats": {
      "intelligence": 75,
      "strength": 100,
      "speed": 95,
      "durability": 100,
      "power": 100,
      "combat": 80
    },
    "appearance": {
      "gender": "Male",
      "race": "Human",
      "height_cm": 185,
      "weight_kg": 110,
      "eye_color": "Blue",
      "hair_color": "Black"
    },
    "biography": {
      "full_name": "William Joseph Batson",
      "alter_egos": "Captain Marvel",
      "place_of_birth": "Fawcett City",
      "first_appearance": "Whiz Comics #2"
    },
    "team": "Justice League"
  }'
```

---

### `PUT /api/characters/{id}`
Reemplaza completamente un personaje existente.

```bash
curl -X PUT http://localhost:24585/api/characters/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Batman",
    "real_name": "Bruce Wayne",
    "alignment": "good",
    "publisher": "DC Comics",
    "powerstats": {
      "intelligence": 100,
      "strength": 30,
      "speed": 30,
      "durability": 55,
      "power": 50,
      "combat": 100
    },
    "appearance": {
      "gender": "Male",
      "race": "Human",
      "height_cm": 188,
      "weight_kg": 95,
      "eye_color": "Blue",
      "hair_color": "Black"
    },
    "biography": {
      "full_name": "Bruce Wayne",
      "alter_egos": "No alter egos found.",
      "place_of_birth": "Gotham City, New Jersey",
      "first_appearance": "Detective Comics #27"
    },
    "team": "Justice League"
  }'
```

---

### `PATCH /api/characters/{id}`
Actualiza parcialmente un personaje (solo los campos enviados).

```bash
# Actualizar solo el equipo y el alineamiento
curl -X PATCH http://localhost:24585/api/characters/12 \
  -H "Content-Type: application/json" \
  -d '{ "team": "Gotham City Sirens", "alignment": "neutral" }'

# Actualizar solo powerstats
curl -X PATCH http://localhost:24585/api/characters/1 \
  -H "Content-Type: application/json" \
  -d '{ "powerstats": { "combat": 100, "intelligence": 100 } }'
```

---

### `DELETE /api/characters/{id}`
Elimina un personaje. Persiste los cambios en el archivo JSON.

```bash
curl -X DELETE http://localhost:24585/api/characters/3
```

---

## Esquema del objeto `Character`

```json
{
  "id": 1,
  "name": "Batman",
  "real_name": "Bruce Wayne",
  "alignment": "good",
  "publisher": "DC Comics",
  "powerstats": {
    "intelligence": 100,
    "strength": 26,
    "speed": 27,
    "durability": 50,
    "power": 47,
    "combat": 100
  },
  "appearance": {
    "gender": "Male",
    "race": "Human",
    "height_cm": 188,
    "weight_kg": 95,
    "eye_color": "Blue",
    "hair_color": "Black"
  },
  "biography": {
    "full_name": "Bruce Wayne",
    "alter_egos": "No alter egos found.",
    "place_of_birth": "Gotham City, New Jersey",
    "first_appearance": "Detective Comics #27"
  },
  "team": "Justice League"
}
```

| Campo           | Tipo   | Descripción                                        |
|-----------------|--------|----------------------------------------------------|
| `id`            | int    | ID único autogenerado                              |
| `name`          | string | Nombre del superhéroe / villano (**requerido**)    |
| `real_name`     | string | Identidad secreta (**requerido**)                  |
| `alignment`     | string | `good`, `bad` o `neutral`                          |
| `publisher`     | string | Editorial (por defecto `DC Comics`)                |
| `powerstats`    | object | Estadísticas de poder, cada una entre 0 y 100      |
| `appearance`    | object | Datos físicos del personaje                        |
| `biography`     | object | Historia y primera aparición                       |
| `team`          | string | Equipo u organización al que pertenece             |

---

## Validaciones

### Campos obligatorios — todos deben estar presentes y no vacíos en POST y PUT

| Campo | Regla |
|---|---|
| `name` | Obligatorio, no vacío |
| `real_name` | Obligatorio, no vacío |
| `alignment` | Obligatorio, solo acepta: `good`, `bad`, `neutral` |
| `publisher` | Obligatorio, no vacío |
| `team` | Obligatorio, no vacío |
| `powerstats.intelligence` … `combat` | Cada uno entre **0 y 100** |
| `appearance.gender` | Obligatorio, no vacío |
| `appearance.race` | Obligatorio, no vacío |
| `appearance.eye_color` | Obligatorio, no vacío |
| `appearance.hair_color` | Obligatorio, no vacío |
| `appearance.height_cm` | Obligatorio, mayor a 0 |
| `appearance.weight_kg` | Obligatorio, mayor a 0 |
| `biography.full_name` | Obligatorio, no vacío |
| `biography.place_of_birth` | Obligatorio, no vacío |
| `biography.first_appearance` | Obligatorio, no vacío |

### Campos desconocidos (typos)

Si se envía un campo con nombre incorrecto (por ejemplo `"gende"` en vez de `"gender"`, o `"powerstast"` en vez de `"powerstats"`), la API retorna un error **400** indicando exactamente qué campo no existe:

```json
{
  "status": 400,
  "error": "Bad Request",
  "mensaje": "El campo 'gende' no existe. Verifique el nombre del campo (posible typo)."
}
```

Esto aplica para POST, PUT y PATCH (en PATCH también se validan los campos dentro de `powerstats`, `appearance` y `biography`).

### Parámetros de query inválidos

Un parámetro de tipo incorrecto (ej. `?id=abc`) retorna `400 Bad Request`.

---

## Respuestas de error

Todos los errores se devuelven como JSON estructurado:

```json
{
  "status": 404,
  "error": "Not Found",
  "mensaje": "No se encontró ningún personaje con id=999"
}
```

| Código | Significado                                  |
|--------|----------------------------------------------|
| 400    | Parámetro inválido o JSON malformado         |
| 404    | Personaje no encontrado                      |
| 405    | Método HTTP no permitido en el endpoint      |
| 422    | Fallo de validación de datos                 |
| 500    | Error interno al persistir en el archivo     |

---

## Persistencia

`POST`, `PUT`, `PATCH` y `DELETE` escriben los cambios de vuelta en `data/characters.json`, por lo que los datos sobreviven reinicios del servidor.