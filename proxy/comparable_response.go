package main

type AvailabilityStatus int32

const (
	Available AvailabilityStatus = iota
	Degraded_freshness
	Degraded_json
	Degraded_http
	Unavailable
)

type ComparableResponse struct {
	body               []byte
	statusCode         int
	availabilityStatus AvailabilityStatus
}

func (me *ComparableResponse) Update(availabilityStatus AvailabilityStatus, statusCode int, body []byte) {

	if me.body == nil {
		// if me is empty assign
		me.statusCode = statusCode
		me.body = body
		me.availabilityStatus = availabilityStatus
	} else if availabilityStatus < me.availabilityStatus {
		// if other is better assign
		me.statusCode = statusCode
		me.body = body
		me.availabilityStatus = availabilityStatus
	}
}
