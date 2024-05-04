package pocketic_test

import (
	"github.com/aviate-labs/agent-go/pocketic"
	"testing"
)

func Endpoints(t *testing.T) *pocketic.PocketIC {
	pic, err := pocketic.New(
		pocketic.WithLogger(new(testLogger)),
		pocketic.WithNNSSubnet(),
		pocketic.WithApplicationSubnet(),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("status", func(t *testing.T) {
		if err := pic.Status(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("blobstore", func(t *testing.T) {
		id, err := pic.UploadBlob([]byte{0, 1, 2, 3})
		if err != nil {
			t.Fatal(err)
		}
		bytes, err := pic.GetBlob(id)
		if err != nil {
			t.Fatal(err)
		}
		if len(bytes) != 4 {
			t.Fatalf("unexpected blob size: %d", len(bytes))
		}
	})

	return pic
}
