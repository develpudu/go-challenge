# Uala challenge Golang

Esta es una implementación simplificada de una plataforma de microblogging similar a Twitter, desarrollada en Golang siguiendo los principios de Clean Architecture para el challenge técnico para Desarrollador Backend en Ualá.

## Características

- **Base de Datos**: Se utilizará una base de datos en memoria para simplificar el desarrollo inicial. En producción, se recomienda utilizar una base de datos relacional como PostgreSQL por su capacidad de manejo de grandes volúmenes de datos y su soporte para transacciones.

- **Publicación de Tweets**: Los usuarios podrán publicar mensajes cortos de hasta 280 caracteres.

- **Seguimiento de Usuarios**: Los usuarios podrán seguir a otros usuarios para ver sus tweets en el timeline.

- **Timeline**: Se mostrará una línea de tiempo con los tweets de los usuarios seguidos.


## Arquitectura

Se utilizará una arquitectura de capas, siguiendo principios de Clean Architecture para asegurar un diseño mantenible y escalable:

- **Capa de Aplicación**: Maneja la lógica de negocio y las reglas de la aplicación.

- **Capa de Dominio**: Define los modelos y entidades del sistema.

- **Capa de Infraestructura**: Maneja la interacción con la base de datos y otros servicios externos.

## Escalabilidad

La aplicación está diseñada para ser escalable:

- Arquitectura limpia que permite cambiar fácilmente la implementación de los repositorios
- Optimización para lecturas con caché de timelines
- Posibilidad de implementar bases de datos distribuidas y sistemas de colas

## Optimización para Lecturas

La aplicación está optimizada para lecturas:

- Caché de timelines en memoria
- Estructuras de datos eficientes para consultas rápidas
- Índices para búsquedas optimizadas

## Estructura del Proyecto

```
.
├── cmd/                   # Punto de entrada de la aplicación
│   └── main.go            # Inicia la aplicación
├── application/           # Capa de aplicación
│   └── usecase/           # Casos de uso
├── domain/                # Capa de dominio
│   ├── entity/            # Entidades
│   └── repository/        # Interfaces de repositorio
├── docker/                # Archivos de configuración de Docker
│   ├── Dockerfile         # Configuración para construir la imagen
│   └── docker-compose.yml # Configuración para desarrollo local
├── infrastructure/        # Capa de infraestructura
│   ├── api/               # API REST
│   │   ├── handler/       # Controladores
│   └── repository/        # Implementaciones de repositorio
│       └── memory/        # Repositorios en memoria
├── integration/           # Pruebas de integración
│   └── api_test.go        # Tests de la API
├── scripts/               # Scripts de utilidad
│   └── tests.sh           # Script para ejecutar pruebas
├── docs/                  # Documentación de la API
├── business.txt           # Assumptions
├── TODO.md                # Lista de tareas pendientes y mejoras futuras
└── README.md              # Documentación
```

## Requisitos

- Go 1.16 o superior
- Docker y Docker Compose (opcional, para ejecución con contenedores)

## Instalación

### Método 1: Ejecución local

1. Clonar el repositorio:

```bash
git clone https://github.com/develpudu/go-challenge.git
cd go-challenge
```

2. Ejecutar la aplicación:

```bash
go run cmd/main.go
```

### Método 2: Ejecución con Docker

1. Clonar el repositorio:

```bash
git clone https://github.com/develpudu/go-challenge.git
cd go-challenge
```

2. Ejecutar el script de automatización:

```bash
./scripts/automation.sh start
```

Este script se encargará de construir la imagen Docker y levantar todos los servicios necesarios automáticamente.

La aplicación estará disponible en http://localhost:8080

Para verificar el estado de los servicios, puede ejecutar:

```bash
./scripts/automation.sh status
```

## Scripts de Automatización

El proyecto incluye varios scripts para facilitar el desarrollo, despliegue y monitorización de la aplicación:

### Script de Automatización General

```bash
./scripts/automation.sh [comando]
```

Comandos disponibles:
- `build` - Construye la imagen Docker de la aplicación
- `start` - Inicia los servicios con Docker Compose
- `stop` - Detiene los servicios
- `restart` - Reinicia los servicios
- `logs` - Muestra los logs de los servicios
- `test` - Ejecuta todas las pruebas
- `test:unit` - Ejecuta solo pruebas unitarias
- `test:int` - Ejecuta solo pruebas de integración
- `clean` - Limpia recursos no utilizados (imágenes, volúmenes)
- `status` - Muestra el estado de los servicios
- `help` - Muestra la ayuda del script

### Script de Despliegue

```bash
./scripts/deploy.sh [entorno] [opciones]
```

Entornos disponibles:
- `dev` - Entorno de desarrollo
- `staging` - Entorno de pruebas
- `prod` - Entorno de producción

Opciones:
- `--build` - Reconstruir la imagen antes del despliegue
- `--force` - Forzar el despliegue incluso si hay pruebas fallidas

### Script de Monitorización

```bash
./scripts/monitor.sh [comando]
```

Comandos disponibles:
- `status` - Muestra el estado actual de la aplicación
- `resources` - Muestra el uso de recursos de los contenedores
- `logs` - Muestra los logs recientes de la aplicación
- `health` - Realiza una verificación de salud de la aplicación
- `report` - Genera un informe completo del estado del sistema


## Pruebas

Para ejecutar todas las pruebas del proyecto, utilice el script proporcionado:

```bash
./scripts/tests.sh
```

Las pruebas de integración verifican el funcionamiento correcto de la API REST, incluyendo la creación de usuarios, publicación de tweets, seguimiento de usuarios y obtención de timelines.

## API REST

### Usuarios

- `POST /users` - Crear un nuevo usuario
- `GET /users` - Obtener todos los usuarios
- `GET /users/{id}` - Obtener un usuario específico
- `POST /users/follow` - Seguir a un usuario (requiere User-ID en header y followed_id en body)
- `POST /users/unfollow` - Dejar de seguir a un usuario (requiere User-ID en header y followed_id en body)

### Tweets

- `POST /tweets` - Crear un nuevo tweet (requiere User-ID en header)
- `GET /tweets` - Obtener todos los tweets
- `GET /tweets/{id}` - Obtener un tweet específico
- `GET /users/tweets` - Obtener tweets de un usuario específico (requiere User-ID en header)
- `GET /timeline` - Obtener timeline de un usuario (requiere User-ID en header)
- `GET /timeline` - Obtener el timeline del usuario autenticado

## Autenticación

Para simplificar, la aplicación utiliza un encabezado `User-ID` para identificar al usuario que realiza la petición.

### Documentación de la API

La documentación esta en español latinoamericano, mientras que el código esta documentado en inglés.

- El archivo `docs/swagger.json` contiene la especificación OpenAPI/Swagger de la API.