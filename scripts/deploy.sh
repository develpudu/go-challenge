#!/bin/bash

# Script para automatizar el despliegue de la aplicación en diferentes entornos
# Este script complementa al script de automatización general

# Colores para mejorar la legibilidad
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Directorio raíz del proyecto
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/.."
cd "$ROOT_DIR"

# Función para mostrar ayuda
show_help() {
    echo -e "${BLUE}Script de Despliegue para la Aplicación de Microblogging${NC}"
    echo ""
    echo "Uso: $0 [entorno] [opciones]"
    echo ""
    echo "Entornos disponibles:"
    echo "  dev          - Entorno de desarrollo"
    echo "  staging      - Entorno de pruebas"
    echo "  prod         - Entorno de producción"
    echo ""
    echo "Opciones:"
    echo "  --build      - Reconstruir la imagen antes del despliegue"
    echo "  --force      - Forzar el despliegue incluso si hay pruebas fallidas"
    echo "  --help       - Mostrar esta ayuda"
    echo ""
}

# Función para validar el entorno
validate_environment() {
    case "$1" in
        dev|staging|prod)
            return 0
            ;;
        *)
            echo -e "${RED}Entorno no válido: $1${NC}"
            show_help
            exit 1
            ;;
    esac
}

# Función para ejecutar pruebas antes del despliegue
run_pre_deploy_tests() {
    echo -e "${YELLOW}Ejecutando pruebas previas al despliegue...${NC}"
    
    # Ejecutar pruebas unitarias
    echo -e "${BLUE}Ejecutando pruebas unitarias...${NC}"
    go test -v ./domain/entity/... ./application/usecase/...
    UNIT_TESTS_RESULT=$?
    
    # Si estamos desplegando en producción, ejecutar también pruebas de integración
    if [ "$1" == "prod" ]; then
        echo -e "\n${BLUE}Ejecutando pruebas de integración...${NC}"
        go test -v ./integration/...
        INT_TESTS_RESULT=$?
    else
        INT_TESTS_RESULT=0
    fi
    
    # Verificar resultados de las pruebas
    if [ $UNIT_TESTS_RESULT -eq 0 ] && [ $INT_TESTS_RESULT -eq 0 ]; then
        echo -e "${GREEN}Todas las pruebas pasaron exitosamente.${NC}"
        return 0
    else
        echo -e "${RED}Algunas pruebas fallaron.${NC}"
        return 1
    fi
}

# Función para construir la imagen Docker
build_image() {
    echo -e "${YELLOW}Construyendo imagen Docker para entorno $1...${NC}"
    
    # Usar diferentes tags según el entorno
    case "$1" in
        dev)
            TAG="uala-microblog:dev"
            ;;
        staging)
            TAG="uala-microblog:staging"
            ;;
        prod)
            TAG="uala-microblog:latest"
            ;;
    esac
    
    # Construir la imagen
    docker build -t $TAG -f docker/Dockerfile --build-arg ENV=$1 .
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Imagen $TAG construida exitosamente.${NC}"
        return 0
    else
        echo -e "${RED}Error al construir la imagen $TAG.${NC}"
        return 1
    fi
}

# Función para desplegar en el entorno especificado
deploy_to_environment() {
    ENV=$1
    echo -e "${YELLOW}Desplegando en entorno: $ENV${NC}"
    
    case "$ENV" in
        dev)
            # Despliegue en entorno de desarrollo (local)
            docker-compose -f docker/docker-compose.yml up -d
            ;;
        staging)
            # Despliegue en entorno de staging
            # Aquí se podría agregar la lógica para desplegar en un servidor de staging
            echo -e "${BLUE}Simulando despliegue en servidor de staging...${NC}"
            echo -e "${GREEN}Despliegue en staging completado (simulación).${NC}"
            ;;
        prod)
            # Despliegue en entorno de producción
            # Aquí se podría agregar la lógica para desplegar en un servidor de producción
            echo -e "${BLUE}Simulando despliegue en servidor de producción...${NC}"
            echo -e "${GREEN}Despliegue en producción completado (simulación).${NC}"
            ;;
    esac
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Despliegue en $ENV completado exitosamente.${NC}"
        return 0
    else
        echo -e "${RED}Error durante el despliegue en $ENV.${NC}"
        return 1
    fi
}

# Variables para opciones
BUILD=false
FORCE=false
ENVIRONMENT=""

# Procesar argumentos
while [ "$1" != "" ]; do
    case $1 in
        dev|staging|prod)
            ENVIRONMENT=$1
            ;;
        --build)
            BUILD=true
            ;;
        --force)
            FORCE=true
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}Opción desconocida: $1${NC}"
            show_help
            exit 1
            ;;
    esac
    shift
done

# Verificar que se haya especificado un entorno
if [ -z "$ENVIRONMENT" ]; then
    echo -e "${RED}Debe especificar un entorno (dev, staging, prod).${NC}"
    show_help
    exit 1
fi

# Validar el entorno
validate_environment "$ENVIRONMENT"

# Ejecutar pruebas previas al despliegue
run_pre_deploy_tests "$ENVIRONMENT"
TESTS_RESULT=$?

# Si las pruebas fallaron y no se forzó el despliegue, abortar
if [ $TESTS_RESULT -ne 0 ] && [ "$FORCE" != "true" ]; then
    echo -e "${RED}Despliegue abortado debido a pruebas fallidas.${NC}"
    echo -e "${YELLOW}Use la opción --force para desplegar de todos modos.${NC}"
    exit 1
fi

# Construir la imagen si se solicitó
if [ "$BUILD" == "true" ]; then
    build_image "$ENVIRONMENT"
    if [ $? -ne 0 ] && [ "$FORCE" != "true" ]; then
        echo -e "${RED}Despliegue abortado debido a errores en la construcción de la imagen.${NC}"
        exit 1
    fi
fi

# Desplegar en el entorno especificado
deploy_to_environment "$ENVIRONMENT"

echo -e "\n${GREEN}Proceso de despliegue completado.${NC}"
exit 0