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
├── network/         # Fase 2: Prototipo de infraestructura de red
│   ├── addressing/  # Tablas de direccionamiento VLSM
│   ├── configs/     # Respaldos de configuración (R-A.txt, SW-A1.txt, etc.)
│   ├── diagrams/    # Diagrama de arquitectura de red actualizado
│   ├── packet-tracer/ # Archivo de simulación prototipo_fase2.pkt
│   └── tests/       # Capturas de evidencia y pruebas de conectividad
├── client/          # cliente en Go (pendiente)
├── server/          # servidor de chat en Go (pendiente)
└── auth-service/    # servicio de autenticación en FastAPI (pendiente)
## Documentación de las Fases

### Fase 1: Planeación de Software
- [Documento de análisis y especificación de requisitos (PDF)](docs/etapa1/Etapa1_Planeacion_Chat_Empresarial.pdf)
- [Diagrama general de casos de uso](docs/etapa1/diagramas/01_casos_uso_general.png)
- [Autenticación y control de sesiones](docs/etapa1/diagramas/02_autenticacion_sesiones.png)
- [Mensajería e historial](docs/etapa1/diagramas/03_mensajeria_historial.png)
- [Administración y estadísticas](docs/etapa1/diagramas/04_administracion_estadisticas.png)

### Fase 2: Prototipo de Red (Infraestructura)
El diseño de la red que alojará la aplicación (VLANs, enrutamiento, DHCP) fue implementado en Cisco Packet Tracer. Todos los entregables técnicos y evidencias de esta fase se encuentran en el directorio `network/`.
- [Informe de avance y matriz de pruebas (PDF)](network/Red_Fase2_Prototipo_PacketTracer_ES.pdf)

## Video de presentación
[Ver video del grupo](https://drive.google.com/drive/folders/1gsS9DFuTsOJIraWtgkEc1jsuhxYSArl_?usp=drive_link)

## Estado del proyecto
- [x] Fase 1: Planeación del software
- [x] Fase 2: Prototipo de red en Packet Tracer
- [ ] Servicio de autenticación (FastAPI)
- [ ] Servidor de chat (Go)
- [ ] Cliente (Go)
- [ ] Estadísticas y pruebas de carga
