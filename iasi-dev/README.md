# IASI DEV, SemVer 2.0.0

Paquete **completo** del modulo `iasi-dev`, preparado a partir de los fuentes proporcionados y de la ultima correccion. No es un parche.

## Estructura

- `src/`: fuentes de produccion, `go.mod`, `go.sum`, sin archivos `_test.go`.
- `tests/`: tests espejo de la estructura de paquetes y script `run.sh`.

**Importante:** Go necesita que los tests de funciones no exportadas esten en el mismo paquete durante la compilacion. `tests/run.sh` crea una copia temporal de `src/`, inserta en ella los tests correspondientes, ejecuta `go test ./...` y borra la copia al salir. **No se modifica `src/`.**

## Reemplazo

Sustituir el contenido del directorio `iasi-dev` por el de este ZIP (respaldar/eliminar primero los archivos obsoletos). No superponer sin borrar porque quedarian tests antiguos en `src/`.

## Ejecutar los tests (Git Bash)

```bash
cd /c/iasi-org-dev/iasi-dev-tools/iasi-dev
bash tests/run.sh
```

`cd src && go test ./...` **no ejecuta** los tests almacenados fuera de `src/`. Usar siempre el script anterior para la bateria completa.

La dependencia `go-toml/v2` se resolvera mediante la cache Go o la red disponible en la maquina. No se ha ejecutado `gofmt` ni se ha alterado ningun repositorio remoto. La variable VERSION permanece sin modificar.
