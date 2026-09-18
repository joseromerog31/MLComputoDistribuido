# MLComputoDistribuido

Sistema distribuido para ejecutar predicciones de Machine Learning mediante distintos servicios desplegados con **Docker**.

## Ejecución del proyecto

Construir las imágenes e iniciar todos los servicios en segundo plano:

```bash
docker compose up --build -d
```

## Verificar el estado del servicio

```bash
curl http://localhost:8080/health
```

## Ejecutar predicciones

### Ejemplo con 6 registros

```bash
curl -X POST \
http://localhost:8080/predict-batch \
-H "Content-Type: application/json" \
-d '{"limit":6}'
```

### Ejemplo con 300 registros

```bash
curl -X POST \
http://localhost:8080/predict-batch \
-H "Content-Type: application/json" \
-d '{"limit":300}'
```

## Prueba de tolerancia a fallos

Detener manualmente `worker-2`:

```bash
docker compose stop worker-2
```

Esperar algunos segundos:

```bash
sleep 7
```

Después se pueden realizar nuevamente las peticiones al endpoint para comprobar el comportamiento del sistema cuando uno de los workers no se encuentra disponible.

## Detener el proyecto

Para detener y eliminar los contenedores creados por Docker Compose:

```bash
docker compose down
```

## Tecnologías usadas

* Go
* PostgreSQL
* Docker
* Python
* Nginx
