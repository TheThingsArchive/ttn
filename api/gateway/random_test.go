package gateway

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/api"
	s "github.com/smartystreets/assertions"
)

func TestRandomizers(t *testing.T) {
	for name, msg := range map[string]interface{}{
		"RandomRxMetadata()":      RandomRxMetadata(),
		"RandomTxConfiguration()": RandomTxConfiguration(),
		"RandomLocation":          RandomLocation(),
		"RandomStatus":            RandomStatus(),
	} {
		t.Run(name, func(t *testing.T) {
			if v, ok := msg.(api.Validator); ok {
				t.Run("Validate", func(t *testing.T) {
					a := s.New(t)
					a.So(v.Validate(), s.ShouldBeNil)
				})
			}
		})
	}
}
