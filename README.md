# Uala challenge Golang

Esta es una implementación simplificada de una plataforma de microblogging similar a Twitter, desarrollada en Golang siguiendo los principios de Clean Architecture para el challenge técnico para Desarrollador Backend en Ualá.

## Características

- **Publicación de Tweets**: Los usuarios podrán publicar mensajes cortos de hasta 280 caracteres.
- **Seguimiento de Usuarios**: Los usuarios podrán seguir a otros usuarios para ver sus tweets en el timeline.
- **Timeline**: Se mostrará una línea de tiempo con los tweets de los usuarios seguidos, optimizada con caché.
- **Despliegue Flexible**: Soporte para ejecución local, con Docker, y serverless en AWS.

## Arquitectura

Se utiliza una arquitectura de capas, siguiendo principios de Clean Architecture:

- **Capa de Dominio**: Define las entidades (`User`, `Tweet`), interfaces de repositorio y reglas de negocio.
- **Capa de Aplicación**: Contiene los casos de uso (crear usuario, publicar tweet, seguir, obtener timeline, etc.).
- **Capa de Infraestructura**: Implementa los detalles técnicos:
    - Repositorios: Implementaciones en memoria (`memory`) y en AWS DynamoDB (`dynamodb`).
    - Caché: Implementación de caché para timelines usando AWS ElastiCache (Redis) (`cache`).
    - API: Exposición de la API REST (`api/handler`).
    - Configuración y Entrypoint: (`cmd/main.go`) que maneja diferentes modos de ejecución (local, aws).
    - IaC: Definición de infraestructura AWS con SAM (`infrastructure/aws/template.yaml`) y Terraform (`infrastructure/aws/elasticache.tf`).

## Optimización para Lecturas

La aplicación está optimizada para lecturas del timeline mediante:

- **AWS ElastiCache (Redis)**: Se utiliza para cachear las timelines generadas, reduciendo la carga sobre DynamoDB.
- **AWS DynamoDB GSI**: Un Global Secondary Index en la tabla de Tweets permite consultas eficientes por `UserID`.
- **Consultas Concurrentes**: Al generar una timeline (en caso de cache miss), las consultas a DynamoDB para obtener los tweets de los usuarios seguidos se realizan de forma concurrente.

## Estructura del Proyecto

```
.
├── cmd/                   # Punto de entrada de la aplicación (main.go)
├── application/           # Capa de aplicación (usecase)
├── domain/                # Capa de dominio (entity, repository interfaces)
├── infrastructure/        # Capa de infraestructura
│   ├── api/               # Implementación API REST (handler)
│   ├── cache/             # Implementación de caché (Redis)
│   ├── repository/        # Implementaciones de repositorio
│   │   ├── dynamodb/      # Repositorio DynamoDB
│   │   └── memory/        # Repositorio en memoria
│   └── aws/               # Infraestructura como Código AWS
│       ├── template.yaml  # Plantilla SAM (Lambda, API GW, DynamoDB)
│       └── elasticache.tf # Configuración Terraform (ElastiCache)
├── integration/           # Pruebas de integración (api_test.go)
├── scripts/               # Scripts de utilidad
│   ├── deploy-serverless.sh # Script de despliegue Serverless (SAM)
│   ├── automation.sh      # Script de automatización Docker (legacy)
│   └── tests.sh           # Script para ejecutar pruebas
├── docs/                  # Documentación adicional (API, Arquitectura)
├── business.txt           # Assumptions
├── TODO.md                # Lista de tareas pendientes
├── go.mod, go.sum         # Dependencias Go
└── README.md              # Este archivo
```

## Requisitos

**Para ejecución local/Docker:**
- Go 1.16 o superior
- Docker y Docker Compose (opcional)

**Para despliegue Serverless en AWS:**
- Cuenta de AWS
- AWS CLI v2 configurado con credenciales
- AWS SAM CLI instalado
- Terraform instalado (para ElastiCache)
- Go 1.16 o superior (para compilar)

## Instalación y Ejecución

### Método 1: Ejecución local (Base de datos en memoria)

Este modo es ideal para desarrollo y pruebas rápidas sin dependencias externas.

1.  Clonar el repositorio: `git clone ...`
2.  `cd go-challenge`
3.  Ejecutar (sin argumentos para usar repositorios en memoria):
    ```bash
    go run cmd/main.go
    ```
    La API estará disponible en `http://localhost:8080`.

**Simulación Local del Modo Lambda:**

Es posible iniciar la aplicación localmente para que utilice la lógica de inicialización de AWS (DynamoDB, Redis) pasando el argumento `aws`. **Importante:** Esto requiere que las tablas DynamoDB (`users`, `tweets`) y la instancia Redis (definida por `REDIS_ENDPOINT` en el código/entorno) existan y sean accesibles desde tu máquina local (o que uses herramientas como DynamoDB Local y un Redis local).

```bash
# Asegúrate de que REDIS_ENDPOINT esté configurado si ejecutas así
export REDIS_ENDPOINT="localhost:6379" # Ejemplo para Redis local
go run cmd/main.go aws
```

### Método 2: Ejecución con Docker (Base de datos en memoria)

1.  Clonar el repositorio y `cd go-challenge`.
2.  Ejecutar: `./scripts/automation.sh start`
    La API estará disponible en `http://localhost:8080`.

### Método 3: Despliegue Serverless en AWS (DynamoDB + ElastiCache)

Este método despliega la aplicación en AWS Lambda, API Gateway, DynamoDB y ElastiCache.

**Pasos:**

1.  **Desplegar ElastiCache (Redis) con Terraform:**
    *   Navegar a `infrastructure/aws/`.
    *   Revisar y **modificar** `elasticache.tf` para ajustar la región, VPC, subnets y security groups a tu entorno AWS.
    *   Ejecutar `terraform init`.
    *   Ejecutar `terraform apply`.
    *   Anotar las salidas `redis_primary_endpoint_address` y `redis_primary_endpoint_port`.

2.  **Desplegar la Aplicación con SAM:**
    *   Navegar a la raíz del proyecto (`cd ../../`).
    *   Revisar y **modificar** `infrastructure/aws/template.yaml`, específicamente la sección `VpcConfig` de `MicroblogApiFunction` para que coincida con la configuración de tu VPC y los security groups permitan la conexión Lambda -> ElastiCache.
    *   Ejecutar el script de despliegue, proporcionando el entorno deseado (e.g., `dev`) y los datos del endpoint de Redis obtenidos de Terraform:
        ```bash
        chmod +x scripts/deploy-serverless.sh
        ./scripts/deploy-serverless.sh <entorno> <redis_endpoint_address> [redis_endpoint_port]
        # Ejemplo:
        # ./scripts/deploy-serverless.sh dev my-redis-endpoint.xxxxx.cache.amazonaws.com
        ```
    *   El script compilará la aplicación, empaquetará el código y desplegará la stack de CloudFormation usando SAM.
    *   La salida del comando `sam deploy` incluirá el endpoint de la API Gateway.

## Scripts

- **`./scripts/deploy-serverless.sh <env> <redis_addr> [redis_port]`**: Compila y despliega la aplicación en AWS usando SAM CLI. Requiere las salidas de Terraform.
- **`./scripts/tests.sh`**: Ejecuta todas las pruebas unitarias y de integración (usando repositorios en memoria).
- **`./scripts/automation.sh [comando]`**: Script para manejar el entorno Docker (build, start, stop, logs, test, etc.).

## API REST

### Usuarios

- `POST /users` - Crear un nuevo usuario
- `GET /users` - Obtener todos los usuarios
- `GET /users/{id}` - Obtener un usuario específico
- `POST /users/follow` - Seguir a un usuario (requiere `User-ID` en header y `followed_id` en body)
- `POST /users/unfollow` - Dejar de seguir a un usuario (requiere `User-ID` en header y `followed_id` en body)

### Tweets

- `POST /tweets` - Crear un nuevo tweet (requiere `User-ID` en header)
- `GET /tweets` - Obtener todos los tweets
- `GET /tweets/{id}` - Obtener un tweet específico
- `GET /users/tweets` - Obtener tweets de un usuario específico (requiere `User-ID` en header)
- `GET /timeline` - Obtener timeline de un usuario (requiere `User-ID` en header)

## Autenticación

Para simplificar, la aplicación utiliza un encabezado `User-ID` para identificar al usuario que realiza la petición en todos los endpoints que lo requieren.

## Documentación Adicional

- **Arquitectura Serverless**: Ver `docs/serverless-architecture.md`.
- **Logging**: La aplicación utiliza el paquete estándar `log/slog` para el logging estructurado en formato JSON, ideal para el análisis en CloudWatch Logs.
- **API Spec**: Ver `docs/swagger.json`.
- **Decisiones/Asunciones**: Ver `business.txt`.