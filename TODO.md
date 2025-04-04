# TODO - Uala-Challenge-Golang

## Estructura del Proyecto
- [x] Crear documentación de arquitectura
- [x] Definir estructura de carpetas siguiendo Clean Architecture
- [x] Crear archivo README.md con instrucciones de ejecución

## Capa de Dominio
- [x] Definir entidades (User, Tweet)
- [x] Definir interfaces de repositorios
- [x] Definir reglas de negocio

## Capa de Aplicación
- [x] Implementar casos de uso para publicación de tweets
- [x] Implementar casos de uso para seguimiento de usuarios
- [x] Implementar casos de uso para obtener timeline

## Capa de Infraestructura
- [x] Implementar repositorios en memoria
- [x] Implementar API REST
- [x] Implementar configuración de la aplicación

## Testing
- [x] Escribir tests unitarios para la capa de dominio
- [x] Escribir tests unitarios para la capa de aplicación
- [x] Escribir tests de integración

## Documentación
- [x] Crear documento de diseño de arquitectura
- [x] Documentar código en inglés
- [x] Crear archivo de Assumptions (business.txt)

## Optimización
- [x] Optimizar para lecturas
- [x] Implementar estrategias de escalabilidad

## Despliegue
### Dockerización
- [x] Crear Dockerfile para la aplicación
- [x] Configurar Docker Compose para desarrollo local
- [x] Implementar estrategias de optimización de imágenes Docker
- [x] Crear scripts de automatización para build y deploy
- [x] Documentar proceso de despliegue con Docker

### Serverless (AWS)
- [x] Configurar AWS Lambda para la aplicación
- [x] Implementar API Gateway para exponer endpoints
- [x] Configurar DynamoDB para almacenamiento persistente
- [x] Implementar ElastiCache para optimización de lecturas
- [ ] Configurar CloudWatch para monitoreo y logs
- [x] Crear plantillas de CloudFormation/SAM para infraestructura
- [x] Implementar estrategia de despliegue continuo (CI/CD)
- [X] Documentar arquitectura serverless