#!/bin/bash

# Script para ejecutar todos los tests del proyecto

# Colores para mejorar la legibilidad
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Ejecutando tests unitarios de la capa de dominio...${NC}"
go test -v ./domain/entity/...

echo -e "\n${BLUE}Ejecutando tests unitarios de la capa de aplicación...${NC}"
go test -v ./application/usecase/...

echo -e "\n${BLUE}Ejecutando tests de integración...${NC}"
go test -v ./integration/...

echo -e "\n${GREEN}Todos los tests completados.${NC}"