package godo

import (
	"context"
	"fmt"
	"net/http"
)

const microDropletCheckpointBasePath = "v2/microdroplets/checkpoints"

// MicroDropletCheckpointStatus represents the status of a MicroDroplet checkpoint.
type MicroDropletCheckpointStatus string

// Possible states for a MicroDroplet checkpoint.
const (
	MicroDropletCheckpointStatusUnknown   = MicroDropletCheckpointStatus("CHECKPOINT_UNKNOWN")
	MicroDropletCheckpointStatusCreating  = MicroDropletCheckpointStatus("CHECKPOINT_CREATING")
	MicroDropletCheckpointStatusAvailable = MicroDropletCheckpointStatus("CHECKPOINT_AVAILABLE")
	MicroDropletCheckpointStatusFailed    = MicroDropletCheckpointStatus("CHECKPOINT_FAILED")
	MicroDropletCheckpointStatusDeleted   = MicroDropletCheckpointStatus("CHECKPOINT_DELETED")
	// MicroDropletCheckpointStatusDeleting means deletion was requested and the
	// checkpoint's stored state is being released. The checkpoint stops being
	// returned once that finishes.
	MicroDropletCheckpointStatusDeleting = MicroDropletCheckpointStatus("CHECKPOINT_DELETING")
)

// MicroDropletCheckpointsService is an interface for interfacing with the
// MicroDroplet checkpoint endpoints of the DigitalOcean API.
//
// Checkpoints are their own collection rather than a sub-resource of a
// MicroDroplet: a checkpoint outlives the MicroDroplet it was captured from,
// and deleting that MicroDroplet does not delete or stop billing for it.
// See: https://docs.digitalocean.com/reference/api/api-reference/#tag/MicroDroplets
type MicroDropletCheckpointsService interface {
	List(ctx context.Context, opt *MicroDropletCheckpointListOptions) ([]MicroDropletCheckpoint, *Response, error)
	Get(ctx context.Context, id string) (*MicroDropletCheckpoint, *Response, error)
	Delete(ctx context.Context, id string) (*Response, error)
}

// MicroDropletCheckpointsServiceOp handles communication with the MicroDroplet
// checkpoint related methods of the DigitalOcean API.
type MicroDropletCheckpointsServiceOp struct {
	client *Client
}

var _ MicroDropletCheckpointsService = &MicroDropletCheckpointsServiceOp{}

// MicroDropletCheckpoint represents a checkpoint of a MicroDroplet (persisted
// memory + disk state), captured when the MicroDroplet is paused or by an
// explicit capture request.
type MicroDropletCheckpoint struct {
	ID             string                       `json:"id,omitempty"`
	MicroDropletID string                       `json:"micro_droplet_id,omitempty"`
	Status         MicroDropletCheckpointStatus `json:"status,omitempty"`
	Name           string                       `json:"name,omitempty"`
	MemoryBytes    uint64                       `json:"memory_bytes,omitempty"`
	DiskBytes      uint64                       `json:"disk_bytes,omitempty"`
	Created        string                       `json:"created_at,omitempty"`
}

// MicroDropletCheckpointListOptions holds the checkpoint-specific list filters
// alongside the shared pagination options.
type MicroDropletCheckpointListOptions struct {
	ListOptions

	// MicroDropletID narrows the list to the checkpoints captured from one
	// MicroDroplet. Leave it empty to list every checkpoint the team owns. It
	// still resolves after that MicroDroplet has been deleted.
	MicroDropletID string `url:"micro_droplet_id,omitempty"`
}

// String returns a human-readable description of a MicroDropletCheckpoint.
func (c MicroDropletCheckpoint) String() string {
	return Stringify(c)
}

// URN returns the MicroDropletCheckpoint ID in a valid DO API URN form.
func (c MicroDropletCheckpoint) URN() string {
	return ToURN("microdroplet_checkpoint", c.ID)
}

type microDropletCheckpointRoot struct {
	Checkpoint *MicroDropletCheckpoint `json:"checkpoint"`
}

type microDropletCheckpointsRoot struct {
	Checkpoints []MicroDropletCheckpoint `json:"checkpoints"`
	Links       *Links                   `json:"links"`
	Meta        *Meta                    `json:"meta"`
}

// List lists the team's MicroDroplet checkpoints, optionally narrowed to one
// MicroDroplet, with optional pagination.
func (s *MicroDropletCheckpointsServiceOp) List(ctx context.Context, opt *MicroDropletCheckpointListOptions) ([]MicroDropletCheckpoint, *Response, error) {
	path, err := addOptions(microDropletCheckpointBasePath, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(microDropletCheckpointsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}
	if m := root.Meta; m != nil {
		resp.Meta = m
	}

	return root.Checkpoints, resp, nil
}

// Get retrieves a MicroDroplet checkpoint by its ID. This is what callers poll
// after requesting a capture, and it resolves after the MicroDroplet the
// checkpoint was captured from has been deleted.
func (s *MicroDropletCheckpointsServiceOp) Get(ctx context.Context, id string) (*MicroDropletCheckpoint, *Response, error) {
	if id == "" {
		return nil, nil, NewArgError("id", "cannot be empty")
	}

	path := fmt.Sprintf("%s/%s", microDropletCheckpointBasePath, id)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(microDropletCheckpointRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Checkpoint, resp, nil
}

// Delete releases the state stored by a MicroDroplet checkpoint. The
// DigitalOcean API returns a 204 on success and does not include a response
// body.
//
// This is the only way to release that storage; deleting the MicroDroplet the
// checkpoint was captured from does not. Deletion is asynchronous: the
// checkpoint reports MicroDropletCheckpointStatusDeleting until its stored
// state has been released, then stops being returned. A MicroDroplet already
// restored from it is unaffected.
func (s *MicroDropletCheckpointsServiceOp) Delete(ctx context.Context, id string) (*Response, error) {
	if id == "" {
		return nil, NewArgError("id", "cannot be empty")
	}

	path := fmt.Sprintf("%s/%s", microDropletCheckpointBasePath, id)

	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}
