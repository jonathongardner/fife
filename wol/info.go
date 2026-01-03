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
	nextWol time.Time
	WolInt  int `yaml:"wolInt" json:"wolInt"` // minutes
}

// Setup return an error if the info has negative or missing mac. Sets default values
func (wi *Info) Setup() error {
	if err := wi.invalid(); err != nil {
		return err
	}

	// default wolInt to 60 minutes
	if wi.WolInt == 0 {
		wi.WolInt = 60
	}

	wi.nextWol = time.Now()
	wi.LastWol = wi.nextWol.Add(-1 * time.Duration(wi.WolInt) * time.Minute)

	return nil
}

// invalid return an error if the info has negative or missing mac
func (wi *Info) invalid() error {
	if wi.WolInt < 0 {
		return fmt.Errorf("wolInt cant be negative")
	}
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
	wi.nextWol = wi.LastWol.Add(time.Duration(wi.WolInt) * time.Minute)
	return nil
}

// ShouldWakeUp return true if the current time is after the next wol
func (wi *Info) ShouldWakeUp() bool {
	return time.Now().After(wi.nextWol)
}
