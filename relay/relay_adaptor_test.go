package relay

import (
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/constant"
)

func TestGetTaskAdaptorSupportsXAI(t *testing.T) {
	if adaptor := GetTaskAdaptor(constant.TaskPlatform(strconv.Itoa(constant.ChannelTypeXai))); adaptor == nil {
		t.Fatalf("expected xAI channel type to have a task adaptor")
	}
}
