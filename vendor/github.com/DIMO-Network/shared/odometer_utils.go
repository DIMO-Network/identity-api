package shared

// IsOdometerValid encapsulates logic to decide whether to return odometer
func IsOdometerValid(odometer float64) bool {
	if odometer <= 100 {
		return false
	}
	if odometer == 65539 || odometer == 65538 || odometer == 983040 {
		return false
	}
	return true
}
