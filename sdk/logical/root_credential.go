// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logical

import (
	"github.com/robfig/cron/v3"
	"time"
)

// RotationOptions is an embeddable struct to capture common lease
// settings between a Secret and Auth
type RotationOptions struct {
	// Schedule holds the info for the framework.Schedule
	Schedule *RootSchedule
}

// RootCredential represents the secret part of a response.
type RootCredential struct {
	RotationOptions

	// RotationID is the ID returned to the user to manage this secret.
	// This is generated by Vault core. Any set value will be ignored.
	// For requests, this will always be blank.
	RotationID string `sentinel:""`
	Path       string
	Name       string
}

func (s *RootCredential) Validate() error {
	// TODO: validation?
	return nil
}

// GetRootCredential initializes a root credential structure based on the passed in rotation_schedule or ttl
// If rotation schedule is empty, the included spec schedule would be nil
// NextVaultRotation and LastVaultRotation are set to zero value; it's the responsibility of callers to set these
// values appropriately
func GetRootCredential(rotationSchedule, path, credentialName string, rotationWindow int, ttl int) (*RootCredential, error) {
	var cronSc *cron.SpecSchedule
	if rotationSchedule != "" {
		var err error
		cronSc, err = DefaultScheduler.Parse(rotationSchedule)
		if err != nil {
			return nil, err
		}
	}

	rs := &RootSchedule{
		Schedule:         cronSc,
		RotationSchedule: rotationSchedule,
		RotationWindow:   time.Duration(rotationWindow) * time.Second,
		TTL:              time.Duration(ttl) * time.Second,
		// TODO
		// decide if next rotation should be set here
		// or when we actually push item into queue
		NextVaultRotation: time.Time{},
		LastVaultRotation: time.Time{},
	}

	return &RootCredential{
		RotationOptions: RotationOptions{
			Schedule: rs,
		},
		// Figure out how to get mount info
		Path: path,
		Name: credentialName,
	}, nil
}
