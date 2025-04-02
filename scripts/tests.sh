#!/bin/bash

# Script para ejecutar todos los tests del proyecto

echo "Ejecutando tests unitarios de la capa de dominio..."
go test -v ./domain/entity/...

echo "\nEjecutando tests unitarios de la capa de aplicación..."
go test -v ./application/usecase/...

echo "\nEjecutando tests de integración..."
go test -v ./integration/...

echo "\nTodos los tests completados."