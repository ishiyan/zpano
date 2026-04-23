package core

// Descriptor classifies an indicator along multiple taxonomic dimensions to enable
// filtering and display in charting catalogs.
type Descriptor struct {
	// Identifier uniquely identifies the indicator.
	Identifier Identifier `json:"identifier"`

	// Family groups related indicators (e.g., by author or category).
	Family string `json:"family"`

	// Adaptivity classifies whether the indicator adapts its parameters.
	Adaptivity Adaptivity `json:"adaptivity"`

	// InputRequirement is the minimum input data type this indicator consumes.
	InputRequirement InputRequirement `json:"inputRequirement"`

	// VolumeUsage classifies how this indicator uses volume information.
	VolumeUsage VolumeUsage `json:"volumeUsage"`

	// Outputs classifies each output of this indicator.
	Outputs []OutputDescriptor `json:"outputs"`
}

// DescriptorOf returns the taxonomic descriptor for the given indicator identifier.
// The second return value is false if no descriptor is registered for the identifier.
func DescriptorOf(id Identifier) (Descriptor, bool) {
	d, ok := descriptors[id]

	return d, ok
}

// Descriptors returns a copy of the full descriptor registry.
func Descriptors() map[Identifier]Descriptor {
	out := make(map[Identifier]Descriptor, len(descriptors))
	for k, v := range descriptors {
		out[k] = v
	}

	return out
}
