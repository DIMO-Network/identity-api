package shared

import "time"

// CloudEvent represents an event according to the CloudEvents spec.
// See https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/spec.md
type CloudEvent[A any] struct {
	// ID is an identifier for the event. The combination of ID and Source must
	// be unique.
	ID string `json:"id"`
	// Source is a URI reference describing the producer of the event.
	Source string `json:"source"`
	// SpecVersion is the version of CloudEvents specification used. For now,
	// this is always "1.0".
	SpecVersion string `json:"specversion"`
	// Subject is an optional field identifying the subject of the event within
	// the context of the event producer.
	Subject string `json:"subject"`
	// Time is an optional field giving the time at which the event occurred.
	Time time.Time `json:"time"`
	// Type describes the type of event. It should generally be a reverse-DNS
	// name.
	Type string `json:"type"`
	// Data contains domain-specific information about the event.
	Data A `json:"data"`
}
