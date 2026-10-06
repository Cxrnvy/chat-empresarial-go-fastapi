# Chat empresarial cliente-servidor (Go + FastAPI)

Proyecto Integrador de Programación Orientada a Objetos, UIDE, paralelo 3ro A

## Descripción del sistema

Es una aplicación de chat interno para una empresa. Los empleados inician sesión con una cuenta creada por la organización, escriben en salas o en conversaciones directas, y pueden volver a leer lo que se dijo. Un administrador gestiona cuentas, salas y sesiones, y consulta estadísticas de uso.

La aplicación se divide en tres partes que se comunican entre sí:

| Componente | Carpeta | Tecnología | Qué hace |
|---|---|---|---|
| Cliente | `client/` | Go | Pide credenciales, guarda el token, muestra mensajes y estadísticas. |
| Servidor de chat | `server/` | Go, WebSocket | Recibe, guarda y reenvía mensajes; mantiene salas y presencia; calcula indicadores. |
| Servicio de autenticación | `auth-service/` | Python, FastAPI, JWT | Guarda cuentas y roles, verifica contraseñas, emite y valida tokens. |

## Problema que resuelve

Cuando una empresa no tiene un canal propio, las conversaciones de trabajo terminan en aplicaciones personales. Nadie controla quién tiene acceso, el historial se pierde cuando alguien cambia de teléfono o se va, y la dirección no puede medir cómo se usa el canal. Un chat interno con cuentas y datos bajo control de la organización resuelve esos tres puntos.

## Alcance preliminar

**Incluye:** autenticación con roles, mensajes en salas y directos en tiempo real, historial, lista de usuarios conectados, control de sesiones y estadísticas básicas de uso.

**Fuera del alcance, por ahora:** archivos adjuntos, llamadas, cifrado de extremo a extremo, notificaciones push y aplicaciones móviles.

## Integrantes

| Integrante | Usuario de GitHub |
|---|---|
| Cesar Bolivar Arciniegas Mejia | Cxrnvy |
| Ariel Alejandro Guerrero Velasco | Arialejo01 |
| Jorge Emilio Aguilar Gaibor | jeaguilarg007-lang |

## Estructura del repositorio

```text
.
├── README.md
├── .gitignore
├── docs/
│   └── etapa1/
│       ├── Etapa1_Planeacion_Chat_Empresarial.pdf
│       └── diagramas/
│           ├── 01_casos_uso_general.png
│           ├── 02_autenticacion_sesiones.png
│           ├── 03_mensajeria_historial.png
│           └── 04_administracion_estadisticas.png
├── client/          # cliente en Go (pendiente)
├── server/          # servidor de chat en Go (pendiente)
└── auth-service/    # servicio de autenticación en FastAPI (pendiente)
```

## Documentación de la Etapa 1

- [Documento de análisis y especificación de requisitos (PDF)](docs/etapa1/Etapa1_Planeacion_Chat_Empresarial.pdf)
- [Diagrama general de casos de uso](docs/etapa1/diagramas/01_casos_uso_general.png)
- [Autenticación y control de sesiones](docs/etapa1/diagramas/02_autenticacion_sesiones.png)
- [Mensajería e historial](docs/etapa1/diagramas/03_mensajeria_historial.png)
- [Administración y estadísticas](docs/etapa1/diagramas/04_administracion_estadisticas.png)

La matriz de trazabilidad y la descripción de los casos de uso están dentro del PDF.

## Video de presentación

[Ver video del grupo](https://drive.google.com/file/d/1HZ8he3dyg9F9lnj_2f-57o_MOJoM8AuJ/view?usp=drive_link)

## Estado del proyecto

- [x] Etapa 1: planeación del software
- [ ] Servicio de autenticación (FastAPI)
- [ ] Servidor de chat (Go)
- [ ] Cliente (Go)
- [ ] Estadísticas y pruebas de carga

## Cómo trabajamos

Cada integrante hace sus commits desde su propia cuenta. Escribimos los mensajes con un prefijo que indica el tipo de cambio: `docs:`, `feat:`, `fix:`, `chore:`, `test:`. Antes de subir cambios hacemos `git pull --rebase` para no pisar el trabajo de los demás.
