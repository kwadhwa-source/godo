package godo

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestMicroDropletCheckpoints_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/checkpoints", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"checkpoints": [
				{"id": "chk-1", "micro_droplet_id": "aaa-111", "status": "CHECKPOINT_AVAILABLE", "memory_bytes": 1024, "disk_bytes": 2048},
				{"id": "chk-2", "micro_droplet_id": "bbb-222", "status": "CHECKPOINT_CREATING"}
			],
			"meta": {"total": 2}
		}`)
	})

	checkpoints, resp, err := client.MicroDropletCheckpoints.List(ctx, nil)
	if err != nil {
		t.Fatalf("MicroDropletCheckpoints.List returned error: %v", err)
	}

	expected := []MicroDropletCheckpoint{
		{ID: "chk-1", MicroDropletID: "aaa-111", Status: MicroDropletCheckpointStatusAvailable, MemoryBytes: 1024, DiskBytes: 2048},
		{ID: "chk-2", MicroDropletID: "bbb-222", Status: MicroDropletCheckpointStatusCreating},
	}
	if !reflect.DeepEqual(checkpoints, expected) {
		t.Errorf("MicroDropletCheckpoints.List returned %+v, expected %+v", checkpoints, expected)
	}

	if resp.Meta == nil || resp.Meta.Total != 2 {
		t.Errorf("MicroDropletCheckpoints.List Meta not propagated: %+v", resp.Meta)
	}
}

// The MicroDroplet is a query filter, so it rides alongside pagination rather
// than changing the path.
func TestMicroDropletCheckpoints_ListByMicroDroplet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/checkpoints", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{"micro_droplet_id": "aaa-111", "page": "2"})
		fmt.Fprint(w, `{"checkpoints": [{"id": "chk-1", "micro_droplet_id": "aaa-111"}]}`)
	})

	checkpoints, _, err := client.MicroDropletCheckpoints.List(ctx, &MicroDropletCheckpointListOptions{
		ListOptions:    ListOptions{Page: 2},
		MicroDropletID: "aaa-111",
	})
	if err != nil {
		t.Fatalf("MicroDropletCheckpoints.List returned error: %v", err)
	}

	if len(checkpoints) != 1 || checkpoints[0].ID != "chk-1" {
		t.Errorf("MicroDropletCheckpoints.List returned %+v", checkpoints)
	}
}

func TestMicroDropletCheckpoints_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/checkpoints/chk-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"checkpoint": {"id": "chk-1", "micro_droplet_id": "aaa-111", "status": "CHECKPOINT_AVAILABLE"}}`)
	})

	checkpoint, _, err := client.MicroDropletCheckpoints.Get(ctx, "chk-1")
	if err != nil {
		t.Fatalf("MicroDropletCheckpoints.Get returned error: %v", err)
	}

	expected := &MicroDropletCheckpoint{ID: "chk-1", MicroDropletID: "aaa-111", Status: MicroDropletCheckpointStatusAvailable}
	if !reflect.DeepEqual(checkpoint, expected) {
		t.Errorf("MicroDropletCheckpoints.Get returned %+v, expected %+v", checkpoint, expected)
	}
}

func TestMicroDropletCheckpoints_Get_EmptyID(t *testing.T) {
	_, _, err := (&MicroDropletCheckpointsServiceOp{}).Get(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDropletCheckpoints_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/checkpoints/chk-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.MicroDropletCheckpoints.Delete(ctx, "chk-1"); err != nil {
		t.Fatalf("MicroDropletCheckpoints.Delete returned error: %v", err)
	}
}

func TestMicroDropletCheckpoints_Delete_EmptyID(t *testing.T) {
	_, err := (&MicroDropletCheckpointsServiceOp{}).Delete(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDropletCheckpoint_URN(t *testing.T) {
	c := MicroDropletCheckpoint{ID: "chk-1"}
	if want, got := "do:microdroplet_checkpoint:chk-1", c.URN(); want != got {
		t.Errorf("MicroDropletCheckpoint.URN returned %q, expected %q", got, want)
	}
}
