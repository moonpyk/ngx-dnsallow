package pkg

const (
	ExitSuccess              = 0  // ExitSuccess indicates a successful execution of the program.
	ExitGenerationError      = 1  // ExitGenerationError indicates an error occurred during the generation of the configuration.
	ExitNginxError           = 2  // ExitNginxError indicates an error occurred while interacting with nginx.
	ExitInvalidConfiguration = 78 // ExitInvalidConfiguration indicates that the configuration provided is invalid or incomplete.
)
