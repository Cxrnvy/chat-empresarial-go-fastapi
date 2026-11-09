# app-go: prototipo del chat empresarial en Go

Prototipo funcional de la aplicación cliente-servidor del proyecto integrador de Programación Orientada a Objetos (UIDE). Es un chat de consola: un **servidor** y uno o varios **clientes** que se ejecutan como procesos independientes y se comunican por TCP.

## Funcionalidades implementadas

- Identificar al usuario con `/login` (contra `data/users.json`) y cerrar sesión con `/logout`.
- Enviar mensajes de texto a una sala (`/say`, o escribir directamente después de `/room`).
- Enviar mensajes directos entre dos personas (`/dm`).
- Recibir los mensajes en tiempo real, sin tener que consultarlos.
- Consultar el historial de una sala (`/history`) o de una conversación directa (`/dmhistory`).
- Ver las salas (`/rooms`), unirse a una sala (`/join`) y ver quién está conectado (`/who`).
- Validaciones: mensaje vacío o de más de 1000 caracteres, usuario o sala inexistente, usuario desactivado, sala a la que no se pertenece, JSON mal formado y comandos desconocidos.
- Registro de eventos del servidor por consola (sin el texto de los mensajes).

## Requisitos de software

- Go 1.21 o superior. El prototipo se desarrolló y probó con **Go 1.22.2** en Linux.
- No usa librerías externas: solo la biblioteca estándar de Go.
- Para la demostración automática (`scripts/demo.sh`): Bash (Linux, macOS o Codespaces).

## Instalación

```bash
git clone https://github.com/Cxrnvy/chat-empresarial-go-fastapi.git
cd chat-empresarial-go-fastapi/app-go
go build ./...
```

## Iniciar el servidor

Desde la carpeta `app-go`:

```bash
go run ./cmd/server
```

Parámetros configurables:

| Parámetro | Valor por defecto | Para qué sirve |
|---|---|---|
| `-addr` | `127.0.0.1:8080` | Dirección IP y puerto en los que escucha el servidor |
| `-data` | `data` | Carpeta con `users.json`, `rooms.json` y `messages.jsonl` |

En el diseño de red del proyecto, el servidor de chat vive en la VLAN 20 de la Sede A (por ejemplo `-addr 10.10.0.194:8080`).

## Ejecutar el cliente

En otra terminal, también desde `app-go`:

```bash
go run ./cmd/client -user ariel
```

| Parámetro | Valor por defecto | Para qué sirve |
|---|---|---|
| `-addr` | `127.0.0.1:8080` | Dirección y puerto del servidor |
| `-user` | (vacío) | Usuario con el que iniciar sesión automáticamente; si se omite, se usa `/login <usuario>` |

Comandos del cliente (también con `/help`):

```
/login <usuario>           inicia sesión
/logout                    cierra la sesión
/rooms                     lista las salas
/join <sala>               te une a una sala
/room <sala>               elige la sala a la que van tus mensajes
/say <sala> <texto>        envía un mensaje a una sala
/dm <usuario> <texto>      envía un mensaje directo
/history [sala] [n]        últimos n mensajes de una sala
/dmhistory <usuario> [n]   últimos n mensajes directos con esa persona
/who                       usuarios conectados
/quit                      sale del programa
```

## Datos de prueba

| Archivo | Contenido |
|---|---|
| `data/users.json` | `cesar` (admin), `ariel` y `jorge` (usuarios) y `lucia` (cuenta desactivada, para probar el rechazo) |
| `data/rooms.json` | `general` (cesar, ariel, jorge) y `proyectos` (cesar, ariel; jorge puede unirse con `/join`) |
| `data/messages.jsonl` | Se crea solo al enviar el primer mensaje (no se sube al repositorio) |

No hay contraseñas: el prototipo solo **identifica** al usuario. La autenticación real la hará el servicio FastAPI de Fundamentos de Seguridad.

## Demostración reproducible

```bash
bash scripts/demo.sh
```

El script compila, levanta el servidor sobre una copia de los datos, conecta dos clientes (`ariel` y `jorge`) como procesos independientes y muestra mensajes en tiempo real, historial y varios casos de error (usuario inexistente, cuenta desactivada, mensaje vacío, sala ajena y servidor caído).

## Pruebas

```bash
go test ./...
go test -race ./...   # además detecta condiciones de carrera
```

Las pruebas incluyen una prueba de integración que levanta un servidor real en un puerto local y conecta clientes TCP.

## Estructura de carpetas

```text
app-go/
├── cmd/
│   ├── client/        # main del cliente (flags -addr y -user)
│   └── server/        # main del servidor (flags -addr y -data)
├── internal/
│   ├── models/        # User, Room, Message, Session y errores de dominio
│   ├── protocol/      # Request y Response: lo que viaja entre cliente y servidor
│   ├── client/        # conexión TCP, interpretación de comandos y salida
│   ├── handlers/      # atiende conexiones y entrega mensajes en tiempo real
│   ├── services/      # ChatService (reglas del chat) y Authenticator
│   └── storage/       # Store y FileStore (archivos JSON)
├── data/              # users.json, rooms.json (y messages.jsonl al ejecutar)
├── scripts/demo.sh    # demostración reproducible
├── go.mod
└── README.md
```

## Limitaciones conocidas

- No hay autenticación real ni cifrado: cualquier persona que conozca un usuario activo puede identificarse como él. Se resolverá con el servicio FastAPI y TLS en las semanas 11 y 12.
- La comunicación usa TCP con mensajes JSON por línea en lugar de WebSocket, para no depender de librerías externas en esta etapa.
- Las membresías nuevas hechas con `/join` solo viven en memoria: se pierden al reiniciar el servidor.
- Los mensajes se guardan en texto plano en `messages.jsonl` (con permisos 0600) y el historial lee el archivo completo en cada consulta.
- No hay vencimiento de sesiones por inactividad, ni reconexión automática del cliente, ni estadísticas, ni funciones de administración.
- Probado en Linux; no se ha probado todavía en Windows.

## Autores y contribuciones

| Integrante | Contribución | Usuario de GitHub |
|---|---|---|
| Cesar Arciniegas | [COMPLETAR] | Cxrnvy |
| Ariel Guerrero | [COMPLETAR] | [COMPLETAR] |
| Jorge Aguilar | [COMPLETAR] | [COMPLETAR] |

Las contribuciones de cada persona se pueden verificar en el historial de commits (`git shortlog -sn`).

## Créditos, licencia y declaración del uso de IA

- **Licencia:** MIT (ver el archivo `LICENSE` en la raíz del repositorio).
- **Código y librerías:** solo biblioteca estándar de Go; no se copió código de terceros.
- **Documentación consultada:** Effective Go y la especificación del lenguaje Go (https://go.dev/doc/).
- **Uso de IA:** se usó Claude (Anthropic, 2026) como apoyo para estructurar el código del prototipo, redactar las pruebas y este README. El equipo revisó, ejecutó y asume el código entregado.
