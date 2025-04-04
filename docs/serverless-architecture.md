# Arquitectura Serverless para la Aplicación de Microblogging

Este documento describe la arquitectura serverless implementada para la aplicación de microblogging en AWS.

## Visión General

La arquitectura serverless permite escalar automáticamente la aplicación según la demanda, reducir costos operativos y simplificar el mantenimiento. La aplicación se ha diseñado siguiendo los principios de Clean Architecture, lo que facilita la migración desde la implementación basada en contenedores a la arquitectura serverless.

## Componentes Principales

### AWS Lambda

AWS Lambda es el núcleo de nuestra arquitectura serverless. La aplicación se ejecuta como una función Lambda que procesa las solicitudes HTTP provenientes de API Gateway.

**Características implementadas:**
- Función Lambda con tiempo de ejecución Go 1.x
- Configuración de memoria y timeout optimizados
- Variables de entorno para configuración
- Integración con API Gateway mediante proxy

### API Gateway

API Gateway expone los endpoints de la API REST y enruta las solicitudes a la función Lambda.

**Características implementadas:**
- Configuración de CORS para permitir solicitudes desde cualquier origen
- Soporte para métodos HTTP (GET, POST, PUT, DELETE)
- Proxy para todas las rutas a la función Lambda
- Etapas de despliegue (dev, staging, prod)

### DynamoDB

DynamoDB proporciona almacenamiento persistente para los datos de la aplicación.

**Características implementadas:**
- Tabla de usuarios con índice secundario global para búsqueda por nombre de usuario
- Tabla de tweets con índice secundario global para búsqueda por ID de usuario y fecha de creación
- Modo de facturación bajo demanda (pay-per-request) para optimizar costos
- Implementación de repositorios que siguen las interfaces definidas en la capa de dominio

### ElastiCache (Redis)

ElastiCache se utiliza para optimizar las operaciones de lectura, especialmente para los timelines de usuarios.

**Características implementadas:**
- Cluster de Redis para almacenamiento en caché
- Servicio de caché para timelines con invalidación automática
- Estrategia de caché con tiempo de expiración
- Optimización para reducir la carga en DynamoDB

### CloudWatch

CloudWatch proporciona monitoreo y registro para la aplicación.

**Características implementadas:**
- Grupo de logs para la función Lambda
- Métricas personalizadas para monitorear el rendimiento
- Alarmas para notificar errores y problemas de rendimiento
- Dashboard para visualizar el estado de la aplicación

## Diagrama de Arquitectura

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│             │     │             │     │             │
│   Cliente   │────▶│ API Gateway │────▶│   Lambda    │
│             │     │             │     │             │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                                               │
                     ┌─────────────┐          │
                     │             │          │
                     │ ElastiCache │◀─────────┤
                     │   (Redis)   │          │
                     │             │          │
                     └─────────────┘          │
                                               │
                     ┌─────────────┐          │
                     │             │          │
                     │  DynamoDB   │◀─────────┘
                     │             │
                     └─────────────┘
```

## Estrategia de Despliegue

Se ha implementado una estrategia de despliegue continuo (CI/CD) utilizando AWS SAM (Serverless Application Model) y scripts de automatización.

**Proceso de despliegue:**
1. Ejecución de pruebas automatizadas
2. Validación de la plantilla SAM
3. Compilación de la aplicación para el entorno Lambda
4. Empaquetado de la aplicación
5. Despliegue en el entorno seleccionado (dev, staging, prod)

## Optimizaciones

### Optimización para Lecturas

La aplicación está optimizada para operaciones de lectura, que son las más frecuentes en una plataforma de microblogging:

- **Caché de timelines**: