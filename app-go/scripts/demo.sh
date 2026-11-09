#!/usr/bin/env bash
# Demostración reproducible del prototipo: levanta el servidor, conecta dos clientes
# (ariel y jorge) y muestra mensajes en tiempo real, historial y varios casos de error.
# Uso, desde la carpeta app-go:   bash scripts/demo.sh
set -u

ADDR="127.0.0.1:18080"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

echo "== Compilando =="
mkdir -p bin
go build -o bin/server ./cmd/server || exit 1
go build -o bin/client ./cmd/client || exit 1

# Se trabaja sobre una copia de los datos para no ensuciar app-go/data con mensajes de prueba.
cp data/users.json data/rooms.json "$WORK"/

echo "== Iniciando el servidor en $ADDR =="
./bin/server -addr "$ADDR" -data "$WORK" > "$WORK/server.log" 2>&1 &
SERVER_PID=$!
sleep 1

echo "== Clientes (dos procesos independientes) =="
{
  sleep 4
  echo "/rooms"
  echo "/history general 5"
  echo "/dmhistory ariel"
  echo "/say proyectos hola"
  echo "/join proyectos"
  echo "/who"
  sleep 1
  echo "/quit"
} | ./bin/client -addr "$ADDR" -user jorge > "$WORK/jorge.out" 2>&1 &
JORGE_PID=$!

{
  sleep 1
  echo "/room general"
  echo "hola equipo, el prototipo ya está corriendo"
  echo "/say general    "
  echo "/dm jorge nos vemos en la demo"
  echo "/dm fantasma hola"
  echo "/say nada hola"
  echo "/who"
  sleep 4
  echo "/quit"
} | ./bin/client -addr "$ADDR" -user ariel > "$WORK/ariel.out" 2>&1 &
ARIEL_PID=$!

wait "$JORGE_PID" "$ARIEL_PID"

echo "== Cuentas que no pueden entrar =="
printf '/quit\n' | ./bin/client -addr "$ADDR" -user fantasma 2>&1 | grep -v "^Conectado"
printf '/quit\n' | ./bin/client -addr "$ADDR" -user lucia 2>&1 | grep -v "^Conectado"

echo "== Error de conexión (servidor inexistente) =="
./bin/client -addr 127.0.0.1:1 2>&1 || true

kill -INT "$SERVER_PID" 2>/dev/null
wait "$SERVER_PID" 2>/dev/null

echo
echo "================ SALIDA DEL CLIENTE ariel ================"
cat "$WORK/ariel.out"
echo
echo "================ SALIDA DEL CLIENTE jorge ================"
cat "$WORK/jorge.out"
echo
echo "================ REGISTRO DEL SERVIDOR ================"
cat "$WORK/server.log"
echo
echo "================ ARCHIVO DE MENSAJES ================"
cat "$WORK/messages.jsonl"
