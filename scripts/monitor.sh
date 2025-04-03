#!/bin/bash

# Script para monitorizar la aplicación y sus recursos
# Este script complementa a los scripts de automatización y despliegue

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
    echo -e "${BLUE}Script de Monitorización para la Aplicación de Microblogging${NC}"
    echo ""
    echo "Uso: $0 [comando]"
    echo ""
    echo "Comandos disponibles:"
    echo "  status       - Muestra el estado actual de la aplicación"
    echo "  resources    - Muestra el uso de recursos de los contenedores"
    echo "  logs         - Muestra los logs recientes de la aplicación"
    echo "  health       - Realiza una verificación de salud de la aplicación"
    echo "  report       - Genera un informe completo del estado del sistema"
    echo "  help         - Muestra esta ayuda"
    echo ""
}

# Función para verificar si Docker está en ejecución
check_docker_running() {
    docker info >/dev/null 2>&1
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Docker no está en ejecución.${NC}"
        echo -e "${YELLOW}Por favor, inicie Docker e intente nuevamente.${NC}"
        exit 1
    fi
}

# Función para verificar si la aplicación está en ejecución
check_app_running() {
    docker-compose -f docker/docker-compose.yml ps | grep "app" | grep "Up" >/dev/null 2>&1
    if [ $? -ne 0 ]; then
        echo -e "${RED}La aplicación no está en ejecución.${NC}"
        echo -e "${YELLOW}Use el comando 'scripts/automation.sh start' para iniciarla.${NC}"
        return 1
    else
        echo -e "${GREEN}La aplicación está en ejecución.${NC}"
        return 0
    fi
}

# Función para mostrar el estado actual de la aplicación
show_status() {
    echo -e "${YELLOW}Verificando estado de la aplicación...${NC}"
    
    # Verificar si la aplicación está en ejecución
    check_app_running
    APP_RUNNING=$?
    
    if [ $APP_RUNNING -eq 0 ]; then
        # Mostrar información detallada
        echo -e "\n${BLUE}Detalles de los contenedores:${NC}"
        docker-compose -f docker/docker-compose.yml ps
        
        echo -e "\n${BLUE}Tiempo de actividad:${NC}"
        docker ps --format "{{.Names}} - {{.Status}}" | grep "app"
        
        echo -e "\n${BLUE}Endpoint de la API:${NC}"
        echo -e "http://localhost:8080"
        
        # Verificar si la API responde
        echo -e "\n${BLUE}Verificando respuesta de la API...${NC}"
        curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/users >/dev/null 2>&1
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}La API está respondiendo correctamente.${NC}"
        else
            echo -e "${RED}La API no está respondiendo.${NC}"
        fi
    fi
}

# Función para mostrar el uso de recursos
show_resources() {
    echo -e "${YELLOW}Monitorizando uso de recursos...${NC}"
    
    # Verificar si la aplicación está en ejecución
    check_app_running
    if [ $? -ne 0 ]; then
        return 1
    fi
    
    echo -e "\n${BLUE}Uso de CPU y memoria:${NC}"
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"
    
    echo -e "\n${BLUE}Espacio en disco utilizado por Docker:${NC}"
    docker system df
}

# Función para mostrar logs recientes
show_logs() {
    echo -e "${YELLOW}Mostrando logs recientes de la aplicación...${NC}"
    
    # Verificar si la aplicación está en ejecución
    check_app_running
    if [ $? -ne 0 ]; then
        return 1
    fi
    
    # Número de líneas a mostrar (por defecto 50)
    LINES=${1:-50}
    
    echo -e "\n${BLUE}Últimas $LINES líneas de logs:${NC}"
    docker-compose -f docker/docker-compose.yml logs --tail=$LINES app
}

# Función para realizar una verificación de salud
check_health() {
    echo -e "${YELLOW}Realizando verificación de salud de la aplicación...${NC}"
    
    # Verificar si la aplicación está en ejecución
    check_app_running
    if [ $? -ne 0 ]; then
        return 1
    fi
    
    echo -e "\n${BLUE}Verificando endpoints principales:${NC}"
    
    # Array de endpoints a verificar
    ENDPOINTS=(
        "/users"
        "/tweets"
        "/timeline"
    )
    
    # Verificar cada endpoint
    for ENDPOINT in "${ENDPOINTS[@]}"; do
        echo -ne "Endpoint $ENDPOINT: "
        STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080$ENDPOINT 2>/dev/null)
        
        if [ "$STATUS" == "200" ]; then
            echo -e "${GREEN}OK (200)${NC}"
        else
            echo -e "${RED}ERROR ($STATUS)${NC}"
        fi
    done
    
    echo -e "\n${BLUE}Verificando rendimiento:${NC}"
    echo -e "Tiempo de respuesta promedio:"
    
    # Realizar 5 solicitudes y calcular el tiempo promedio
    TOTAL_TIME=0
    for i in {1..5}; do
        TIME=$(curl -s -o /dev/null -w "%{time_total}" http://localhost:8080/users 2>/dev/null)
        TOTAL_TIME=$(echo "$TOTAL_TIME + $TIME" | bc)
    done
    
    AVG_TIME=$(echo "scale=3; $TOTAL_TIME / 5" | bc)
    echo -e "${GREEN}$AVG_TIME segundos${NC}"
}

# Función para generar un informe completo
generate_report() {
    REPORT_FILE="monitoring_report_$(date +"%Y%m%d_%H%M%S").txt"
    
    echo -e "${YELLOW}Generando informe completo del sistema...${NC}"
    echo -e "El informe se guardará en: ${BLUE}$REPORT_FILE${NC}"
    
    # Crear el archivo de informe
    echo "=== INFORME DE MONITORIZACIÓN - $(date) ===" > $REPORT_FILE
    echo "" >> $REPORT_FILE
    
    # Verificar estado de la aplicación
    echo "=== ESTADO DE LA APLICACIÓN ===" >> $REPORT_FILE
    docker-compose -f docker/docker-compose.yml ps >> $REPORT_FILE 2>&1
    echo "" >> $REPORT_FILE
    
    # Información de recursos
    echo "=== USO DE RECURSOS ===" >> $REPORT_FILE
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}" >> $REPORT_FILE 2>&1
    echo "" >> $REPORT_FILE
    
    # Información del sistema Docker
    echo "=== INFORMACIÓN DEL SISTEMA DOCKER ===" >> $REPORT_FILE
    docker system df >> $REPORT_FILE 2>&1
    echo "" >> $REPORT_FILE
    
    # Logs recientes
    echo "=== LOGS RECIENTES (últimas 20 líneas) ===" >> $REPORT_FILE
    docker-compose -f docker/docker-compose.yml logs --tail=20 app >> $REPORT_FILE 2>&1
    echo "" >> $REPORT_FILE
    
    # Verificación de endpoints
    echo "=== VERIFICACIÓN DE ENDPOINTS ===" >> $REPORT_FILE
    ENDPOINTS=(
        "/users"
        "/tweets"
        "/timeline"
    )
    
    for ENDPOINT in "${ENDPOINTS[@]}"; do
        STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080$ENDPOINT 2>/dev/null)
        echo "Endpoint $ENDPOINT: $STATUS" >> $REPORT_FILE
    done
    
    echo "" >> $REPORT_FILE
    echo "=== FIN DEL INFORME ===" >> $REPORT_FILE
    
    echo -e "${GREEN}Informe generado exitosamente.${NC}"
    echo -e "Para ver el informe, ejecute: ${BLUE}cat $REPORT_FILE${NC}"
}

# Verificar si Docker está en ejecución
check_docker_running

# Procesar el comando proporcionado
case "$1" in
    status)
        show_status
        ;;
    resources)
        show_resources
        ;;
    logs)
        show_logs $2
        ;;
    health)
        check_health
        ;;
    report)
        generate_report
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