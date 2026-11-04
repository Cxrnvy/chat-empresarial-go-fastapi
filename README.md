# Chat empresarial cliente-servidor (Go + FastAPI)

Proyecto Integrador — 3.er nivel (Ingeniería en Sistemas de Información, UIDE, Paralelo A)

## De qué trata el proyecto

Construimos un sistema de chat interno pensado para comunicar dos sedes de una empresa. La idea es que los empleados inicien sesión con una cuenta corporativa, conversen en tiempo real por salas o mensajes directos y puedan consultar el historial cuando lo necesiten. Del lado administrativo, el sistema permite gestionar cuentas, salas y sesiones activas, además de revisar estadísticas generales de uso sin invadir la privacidad de los mensajes.

Separamos la solución en tres componentes principales para no mezclar responsabilidades:

| Componente | Carpeta | Tecnología | Qué hace |
|---|---|---|---|
| Cliente | `client/` | Go | Solicita las credenciales, mantiene el token en memoria y muestra el chat y las estadísticas. |
| Servidor de chat | `server/` | Go, WebSocket | Recibe, cifra, guarda y reenvía los mensajes; controla quién está en cada sala y calcula indicadores. |
| Servicio de autenticación | `auth-service/` | Python, FastAPI, JWT | Administra cuentas y roles, verifica contraseñas y emite o revoca tokens de sesión. |

## Qué problema buscamos resolver

Cuando una empresa no cuenta con un canal propio, el trabajo diario termina en aplicaciones de mensajería personal. Eso trae tres problemas claros: nadie controla realmente quién sigue teniendo acceso a los grupos, la información se pierde si alguien cambia de teléfono o deja la empresa, y no hay forma de medir el uso del canal. Tener una herramienta interna con cuentas, red y almacenamiento bajo control de la organización resuelve esos tres puntos.

## Alcance

* **Qué incluye:** autenticación con roles (usuario y administrador), chat en tiempo real en salas y mensajes directos, historial persistente, lista de usuarios conectados, control de sesiones y reportes agregados de uso.
* **Qué queda fuera por ahora:** envío de archivos adjuntos, llamadas de voz o video, cifrado de extremo a extremo, notificaciones push y versión móvil.

## Equipo

| Integrante | Usuario de GitHub |
|---|---|
| Cesar Bolivar Arciniegas Mejia | [Cxrnvy](https://github.com/Cxrnvy) |
| Ariel Alejandro Guerrero Velasco | [Arialejo01](https://github.com/Arialejo01) |
| Jorge Emilio Aguilar Gaibor | [jeaguilarg007-lang](https://github.com/jeaguilarg007-lang) |

## Estructura del repositorio

```text
.
├── README.md
├── .gitignore
├── docs/
│   ├── etapa1/
│   │   ├── Etapa1_Planeacion_Chat_Empresarial.pdf
│   │   └── diagramas/
│   │       ├── 01_casos_uso_general.png
│   │       ├── 02_autenticacion_sesiones.png
│   │       ├── 03_mensajeria_historial.png
│   │       └── 04_administracion_estadisticas.png
│   └── seguridad/
│       └── Seguridad_Fase2_Diseno_de_seguridad.pdf
├── network/           # Fase 2: Prototipo de red entre las dos sedes
│   ├── addressing/    # Tablas de direccionamiento VLSM
│   ├── configs/       # Respaldos de configuración (R-A.txt, SW-A1.txt, etc.)
│   ├── diagrams/      # Diagrama de topología actualizado
│   ├── packet-tracer/ # Simulación en prototipo_fase2.pkt
│   └── tests/         # Capturas y pruebas de conectividad
├── client/            # Cliente en Go (en desarrollo)
├── server/            # Servidor de chat en Go (en desarrollo)
└── auth-service/      # Servicio de autenticación en FastAPI (en desarrollo)
