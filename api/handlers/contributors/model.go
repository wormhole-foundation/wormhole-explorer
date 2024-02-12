package contributors

type ContributorTotalValuesDTO struct {
	Contributor           string `json:"contributor"`
	TotalValueLocked      string `json:"total_value_locked,omitempty"`
	TotalValueSecured     string `json:"total_value_secured,omitempty"`
	TotalValueTransferred string `json:"total_value_transferred,omitempty"`
	TotalMessages         string `json:"total_messages"`
	LastDayMessages       string `json:"last_day_messages"`
	LastDayDiffPercentage string `json:"last_day_diff_percentage"`
}
