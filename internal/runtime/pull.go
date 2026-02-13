package runtime

import (
	"log/slog"

	logger "github.com/harsha3330/crun/internal/log"
)

func Pull(log *slog.Logger, stater logger.Console, image string) error {
	log.Info("Starting pull the image", "value", image)
	stater.Step("Pulling the image", "value", image)

	return nil
}
