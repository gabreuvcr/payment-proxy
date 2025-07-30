package util

import (
	"math/rand"

	"github.com/gabreuvcr/proxy-payment/internal/model"
)

func RandomProcessor() model.Processor {
	if rand.Intn(2) == 0 {
		return model.ProcessorDefault
	}
	return model.ProcessorFallback
}
