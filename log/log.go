package log

import (
	"log"

	"github.com/pkg/errors"
)

func FatalIfError(err error, message string) {
	if err != nil {
		log.Fatal(errors.Wrap(err, message))
	}
}
