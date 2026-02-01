package wol

import (
	"context"
	"fmt"
	"time"
)

// wolInfo info when sedning wol before a proxy request
type Info struct {
	Mac     string    `yaml:"mac" json:"mac"` // Mac adress, if set will do WOL
	LastWol time.Time `json:"lastWake"`
}

// Setup return an error if the info has negative or missing mac. Sets default values
func (wi *Info) Setup() error {
	if err := wi.invalid(); err != nil {
		return err
	}

	return nil
}

// invalid return an error if the info has negative or missing mac
func (wi *Info) invalid() error {
	if wi.Mac == "" {
		return fmt.Errorf("mac cant be blank")
	}
	return nil
}

// WakeUp make wol to mac. Updates LastWol and nextWol
func (wi *Info) WakeUp(ctx context.Context) error {
	if err := wi.invalid(); err != nil {
		return fmt.Errorf("invalid wol info: %w", err)
	}
	if err := Run(ctx, wi.Mac); err != nil {
		return fmt.Errorf("error running wake up: %w", err)
	}
	wi.LastWol = time.Now()
	return nil
}
