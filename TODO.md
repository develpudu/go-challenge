# TODO - Uala-Challenge-Golang

## Estructura del Proyecto
- [ ] Crear documentación de arquitectura
- [ ] Definir estructura de carpetas siguiendo Clean Architecture
- [ ] Crear archivo README.md con instrucciones de ejecución

## Capa de Dominio
- [ ] Definir entidades (User, Tweet)
- [ ] Definir interfaces de repositorios
- [ ] Definir reglas de negocio

## Capa de Aplicación
- [ ] Implementar casos de uso para publicación de tweets
- [ ] Implementar casos de uso para seguimiento de usuarios
- [ ] Implementar casos de uso para obtener timeline

## Capa de Infraestructura
- [ ] Implementar repositorios en memoria
- [ ] Implementar API REST
- [ ] Implementar configuración de la aplicación

## Testing
- [ ] Escribir tests unitarios para la capa de dominio
- [ ] Escribir tests unitarios para la capa de aplicación
- [ ] Escribir tests de integración

## Documentación
- [ ] Crear documento de diseño de arquitectura
- [ ] Documentar código en inglés
- [ ] Crear archivo de supuestos (business.txt)

## Optimización
- [ ] Optimizar para lecturas
- [ ] Implementar estrategias de escalabilidad

## Despliegue
### Dockerización
- [ ] Crear Dockerfile para la aplicación
- [ ] Configurar Docker Compose para desarrollo local
- [ ] Implementar estrategias de optimización de imágenes Docker
- [ ] Crear scripts de automatización para build y deploy
- [ ] Documentar proceso de despliegue con Docker

### Serverless (AWS)
- [ ] Configurar AWS Lambda para la aplicación
- [ ] Implementar API Gateway para exponer endpoints
- [ ] Configurar DynamoDB para almacenamiento persistente
- [ ] Implementar ElastiCache para optimización de lecturas
- [ ] Configurar CloudWatch para monitoreo y logs
- [ ] Crear plantillas de CloudFormation/SAM para infraestructura
- [ ] Implementar estrategia de despliegue continuo (CI/CD)
- [ ] Documentar arquitectura serverless