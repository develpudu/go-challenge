# Uala challenge Golang

Esta es una implementación simplificada de una plataforma de microblogging similar a Twitter, desarrollada en Golang siguiendo los principios de Clean Architecture para el challenge técnico para Desarrollador Backend en Ualá.

## Características

- Publicación de tweets (máximo 280 caracteres)
- Seguimiento de usuarios
- Timeline personalizado con tweets de usuarios seguidos
- API REST para interactuar con la plataforma

## Arquitectura

La aplicación sigue los principios de Clean Architecture, dividida en las siguientes capas:

### Capa de Dominio

Contiene las entidades del negocio (User, Tweet) y las interfaces de repositorio que definen cómo se accede a los datos.

### Capa de Aplicación

Contiene los casos de uso que implementan la lógica de negocio utilizando las entidades del dominio y las interfaces de repositorio.

### Capa de Infraestructura

Contiene las implementaciones concretas de los repositorios (en memoria) y los controladores de la API REST.

## Estructura del Proyecto

```
.
├── application/           # Capa de aplicación
│   └── usecase/           # Casos de uso
├── domain/                # Capa de dominio
│   ├── entity/            # Entidades
│   └── repository/        # Interfaces de repositorio
├── infrastructure/        # Capa de infraestructura
│   ├── api/               # API REST
│   │   └── handler/       # Controladores
│   └── repository/        # Implementaciones de repositorio
│       └── memory/        # Repositorios en memoria
├── main.go                # Punto de entrada de la aplicación
├── business.txt           # Assumptions
└── README.md              # Documentación
```

## Requisitos

- Go 1.16 o superior

## Instalación

1. Clonar el repositorio:

```bash
git clone https://github.com/develpudu/go-challenge.git
cd go-challenge
```

2. Ejecutar la aplicación:

```bash
go run main.go
```

La aplicación estará disponible en http://localhost:8080

## API REST

### Usuarios

- `POST /users` - Crear un nuevo usuario
- `GET /users` - Obtener todos los usuarios
- `GET /users/{id}` - Obtener un usuario específico
- `POST /users/follow` - Seguir a un usuario
- `POST /users/unfollow` - Dejar de seguir a un usuario

### Tweets

- `POST /tweets` - Crear un nuevo tweet
- `GET /tweets` - Obtener todos los tweets
- `GET /tweets/{id}` - Obtener un tweet específico
- `GET /users/tweets?user_id={id}` - Obtener tweets de un usuario específico
- `GET /timeline` - Obtener el timeline del usuario autenticado

## Autenticación

Para simplificar, la aplicación utiliza un encabezado `User-ID` para identificar al usuario que realiza la petición.

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