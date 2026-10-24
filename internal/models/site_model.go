package models

import "time"

type Site struct {
	Id                 string
	isBookingAllowed   string
	siteName           string
	siteAddress        SiteAddress
	supervisorDetails  SupervisorDetails
	siteOpeningTime    time.Time
	siteClosingTime    time.Time
	totalCapacity      int16
	valetFee           float64
	siteInsuranceAddOn SiteInsuranceAddOn
	tippingAllowed     bool
	tippingDetails     TippingDetails
	parkingType        string
	driverCount        int16
	createdAt          time.Time
	updatedAt          time.Time
}

type SiteAddress struct {
	pinCode         int16
	city            string
	state           string
	addressLine1    string
	addressLine2    string
	siteCoordinates float64
}

type SupervisorDetails struct {
	supervisorMobileNumber string
	supervisorDob          string
	supervisorGender       string
	supervisorName         string
	siteImages             []string
}

type SiteInsuranceAddOn struct {
	insuranceAddOnAllowed bool
	insuranceAddOnPrice   float32
}

type TippingDetails struct {
	tippingAmount      []float32
	pickTippingAllowed bool
	dropTippingAllowed bool
}
