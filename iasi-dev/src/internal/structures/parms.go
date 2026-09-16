package structures

import "os"

type Parms struct {
	Verbose                int      // Nivel de detalle de la salida.
	All                    bool     // Ejecuta también las etapas anteriores.
	Checkpoints            bool     // Usa workflows intermedios como checkpoints.
	Debug                  bool     // Muestra mensajes temporales de depuración.
	PrepareOnly            bool     // Prepara parámetros, los registra y termina (-m, modo interno).
	DryRun                 bool     // Ejecuta el flujo sin lanzar comandos externos (-M, modo interno).
	Force                  bool     // Fuerza la operación cuando está soportado.
	Help                   bool     // Muestra la ayuda.
	Install                bool     // Instala el artefacto cuando proceda.
	Local                  bool     // Mantiene local la publicación cuando la operación lo soporta.
	Push                   bool     // Ejecuta solo la publicación remota cuando promote lo soporta.
	Tolerant               bool     // Continúa cuando una operación falla.
	Message                string   // Mensaje utilizado para el commit.
	Format                 string   // Formato o formatos de salida.
	Platforms              []string // Plataformas preparadas para build; por defecto windows y linux.
	Path                   string   // Directorio de trabajo solicitado con --path.
	LogDir                 string   // Directorio de logs solicitado con --log.
	Organization           string   // Organización GitHub explícita o deducida del workspace.
	Version                string   // Versión actual de la organización leída de GitHub.
	TargetVersion          string   // Versión solicitada por promote o restore.
	MaterializeDestination string   // Directorio destino solicitado por materialize, resuelto antes de --path.
	Subcommand             string   // Subcomando cuando command es workflow.
	LogFile                *os.File // Handle al fichero de log de la ejecución.
	RC                     *int     // Código de retorno acumulativo compartido.
	LastRC                 int      // Resultado de la última operación ejecutada; lo consumen los workflows.
	Targets                []string // Rutas de proyectos IASI descubiertos dentro del ámbito solicitado.
	TargetDetails          []Target // Metadatos de cada proyecto IASI descubierto.
	RequestedTargets       []string // Objetivos solicitados por el usuario.
	Exclusions             []string // Nombres excluidos durante el descubrimiento.
	Repos                  []string // Repositorios Git descubiertos.
	BlackList              []string // Repositorios que no deben procesarse en commit.
}
