#!/bin/bash

# Script de automatización para la aplicación de microblogging
# Este script proporciona comandos para construir, desplegar, probar y gestionar la aplicación

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
    echo -e "${BLUE}Script de Automatización para la Aplicación de Microblogging${NC}"
    echo ""
    echo "Uso: $0 [comando]"
    echo ""
    echo "Comandos disponibles:"
    echo "  build         - Construye la imagen Docker de la aplicación"
    echo "  start         - Inicia los servicios con Docker Compose"
    echo "  stop          - Detiene los servicios"
    echo "  restart       - Reinicia los servicios"
    echo "  logs          - Muestra los logs de los servicios"
    echo "  test          - Ejecuta todas las pruebas"
    echo "  test:unit     - Ejecuta solo pruebas unitarias"
    echo "  test:int      - Ejecuta solo pruebas de integración"
    echo "  clean         - Limpia recursos no utilizados (imágenes, volúmenes)"
    echo "  status        - Muestra el estado de los servicios"
    echo "  help          - Muestra esta ayuda"
    echo ""
}

# Función para construir la imagen Docker
build_app() {
    echo -e "${YELLOW}Construyendo la imagen Docker de la aplicación...${NC}"
    docker build -t uala-microblog -f docker/Dockerfile .
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Imagen construida exitosamente.${NC}"
    else
        echo -e "${RED}Error al construir la imagen.${NC}"
        exit 1
    fi
}

# Función para iniciar los servicios
start_services() {
    echo -e "${YELLOW}Iniciando servicios con Docker Compose...${NC}"
    docker-compose -f docker/docker-compose.yml up -d
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Servicios iniciados exitosamente.${NC}"
        echo -e "La aplicación está disponible en ${BLUE}http://localhost:8080${NC}"
    else
        echo -e "${RED}Error al iniciar los servicios.${NC}"
        exit 1
    fi
}

# Función para detener los servicios
stop_services() {
    echo -e "${YELLOW}Deteniendo servicios...${NC}"
    docker-compose -f docker/docker-compose.yml down
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Servicios detenidos exitosamente.${NC}"
    else
        echo -e "${RED}Error al detener los servicios.${NC}"
        exit 1
    fi
}

# Función para reiniciar los servicios
restart_services() {
    stop_services
    start_services
}

# Función para mostrar logs
show_logs() {
    echo -e "${YELLOW}Mostrando logs de los servicios...${NC}"
    docker-compose -f docker/docker-compose.yml logs -f
}

# Función para ejecutar todas las pruebas
run_tests() {
    echo -e "${YELLOW}Ejecutando todas las pruebas...${NC}"
    
    # Ejecutar el script tests.sh que ya contiene todos los tests configurados
    bash "${ROOT_DIR}/scripts/tests.sh"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Todos los tests completados exitosamente.${NC}"
    else
        echo -e "${RED}Algunos tests fallaron.${NC}"
        exit 1
    fi
}

# Función para ejecutar pruebas unitarias
run_unit_tests() {
    echo -e "${YELLOW}Ejecutando pruebas unitarias...${NC}"
    go test -v ./domain/entity/... ./application/usecase/...
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Pruebas unitarias completadas exitosamente.${NC}"
    else
        echo -e "${RED}Algunas pruebas unitarias fallaron.${NC}"
        exit 1
    fi
}

# Función para ejecutar pruebas de integración
run_integration_tests() {
    echo -e "${YELLOW}Ejecutando pruebas de integración...${NC}"
    go test -v ./integration/...
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Pruebas de integración completadas exitosamente.${NC}"
    else
        echo -e "${RED}Algunas pruebas de integración fallaron.${NC}"
        exit 1
    fi
}

# Función para limpiar recursos no utilizados
clean_resources() {
    echo -e "${YELLOW}Limpiando recursos no utilizados...${NC}"
    
    # Detener contenedores si están en ejecución
    docker-compose -f docker/docker-compose.yml down 2>/dev/null
    
    # Eliminar imágenes no utilizadas
    echo -e "${BLUE}Eliminando imágenes no utilizadas...${NC}"
    docker image prune -af --filter "label=project=uala-microblog"
    
    # Eliminar volúmenes no utilizados
    echo -e "${BLUE}Eliminando volúmenes no utilizados...${NC}"
    docker volume prune -f
    
    echo -e "${GREEN}Limpieza completada.${NC}"
}

# Función para mostrar el estado de los servicios
show_status() {
    echo -e "${YELLOW}Estado de los servicios:${NC}"
    docker-compose -f docker/docker-compose.yml ps
}

# Procesar el comando proporcionado
case "$1" in
    build)
        build_app
        ;;
    start)
        start_services
        ;;
    stop)
        stop_services
        ;;
    restart)
        restart_services
        ;;
    logs)
        show_logs
        ;;
    test)
        run_tests
        ;;
    test:unit)
        run_unit_tests
        ;;
    test:int)
        run_integration_tests
        ;;
    clean)
        clean_resources
        ;;
    status)
        show_status
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo -e "${RED}Comando desconocido: $1${NC}"
        show_help
        exit 1
        ;;
esac

exit 0