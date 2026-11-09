# Chat empresarial cliente-servidor (Go + FastAPI)

Proyecto Integrador de tercer nivel, Universidad Internacional del Ecuador (UIDE), Ingeniería en Sistemas de Información.

## Descripción del sistema

Chat interno para una empresa de soluciones tecnológicas con dos sedes. Los empleados inician sesión, escriben en salas o en conversaciones directas y pueden volver a leer lo que se dijo. Se compone de tres partes:

| Componente | Carpeta | Tecnología | Estado |
|---|---|---|---|
| Cliente y servidor de chat | `app-go/` | Go (biblioteca estándar), TCP + JSON | Prototipo funcional (semana 10) |
| Servicio de autenticación | `auth-service/` | Python, FastAPI | En desarrollo (Fundamentos de Seguridad) |
| Red de dos sedes | `network/` | OpenWrt, VLAN, OSPF; simulación en Packet Tracer | Prototipo en Packet Tracer (semana 9) |

## Problema que resuelve

Cuando una empresa no tiene un canal propio, las conversaciones de trabajo terminan en aplicaciones personales: nadie controla quién tiene acceso, el historial se pierde cuando alguien cambia de teléfono y la dirección no puede medir cómo se usa el canal. Un chat interno con cuentas y datos bajo control de la organización resuelve esos tres puntos.

## Requisitos de software y hardware

- Go 1.21 o superior (probado con 1.22.2).
- Cualquier computador con Linux, macOS o Windows (probado en Linux).
- Para la parte de red: Cisco Packet Tracer (simulación) y, desde la semana 10, Raspberry Pi con OpenWrt y switches Cisco.

## Cómo ejecutar el prototipo

Las instrucciones completas están en [`app-go/README.md`](app-go/README.md). En resumen:

```bash
cd app-go
go run ./cmd/server                 # terminal 1
go run ./cmd/client -user ariel     # terminal 2
bash scripts/demo.sh                # demostración automática
```

## Estructura del repositorio

```text
.
├── README.md
├── LICENSE
├── .gitignore
├── app-go/                  # cliente y servidor de chat en Go (ver su README)
│   ├── cmd/                 # client/ y server/
│   ├── internal/            # models, protocol, client, handlers, services, storage
│   ├── data/                # users.json y rooms.json de prueba
│   ├── scripts/demo.sh
│   └── go.mod
├── auth-service/            # servicio de autenticación en FastAPI (en desarrollo)
├── network/                 # Fase 2 de Redes: prototipo de red entre las dos sedes
│   ├── addressing/          # tablas de direccionamiento VLSM
│   ├── configs/             # respaldos de configuración (R-A.txt, SW-A1.txt, etc.)
│   ├── diagrams/            # diagrama de topología actualizado
│   ├── packet-tracer/       # simulación en prototipo_fase2.pkt
│   └── tests/               # capturas y pruebas de conectividad
└── docs/
    ├── etapa1/              # Etapa 1: planeación (PDF y diagramas de casos de uso)
    ├── seguridad/           # Fundamentos de Seguridad, Fase 2: diseño de seguridad
    └── poo/                 # Programación Orientada a Objetos, entrega de la semana 10
```

## Estado del proyecto

- [x] Etapa 1: planeación del software
- [ ] Prototipo de red en Packet Tracer (Fase 2 de Redes, en curso)
- [ ] Diseño de seguridad (Fase 2 de Fundamentos de Seguridad, en revisión)
- [x] Prototipo funcional del chat en Go (semana 10)
- [ ] Servicio de autenticación FastAPI
- [ ] Integración Go + FastAPI + red (semanas 11 y 12)
- [ ] Estadísticas, pruebas completas y presentación final

## Autores y contribuciones

| Integrante | Usuario de GitHub |
|---|---|
| Cesar Arciniegas | Cxrnvy |
| Ariel Guerrero | [COMPLETAR] |
| Jorge Aguilar | [COMPLETAR] |

Cada integrante hace sus commits desde su propia cuenta. Los mensajes usan un prefijo (`feat:`, `fix:`, `docs:`, `test:`, `chore:`) y antes de subir cambios se hace `git pull --rebase`.

## Licencia, créditos y uso de IA

- **Licencia:** MIT (archivo `LICENSE`).
- **Créditos:** el código de `app-go/` usa solo la biblioteca estándar de Go. Las referencias bibliográficas están en los documentos de `docs/`.
- **Uso de IA:** se usó Claude (Anthropic, 2026) como apoyo para estructurar código, pruebas y documentos. El equipo revisó el contenido y asume el resultado.
