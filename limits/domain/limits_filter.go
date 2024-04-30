package domain

type LimitsFilter struct {
	LimitTypes []LimitType `json:"limit_types,omitempty"`
	Currency   *string     `json:"currency,omitempty"`
	Entities   Attributes  `json:"entities,omitempty"`
	Period     *PeriodType `json:"period,omitempty"`
	Timezone   *string     `json:"timezone,omitempty"`
	Limit      *uint64     `json:"limit,omitempty"`
	Offset     *uint64     `json:"offset,omitempty"`
}
