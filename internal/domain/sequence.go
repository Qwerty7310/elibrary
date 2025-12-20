package domain

type BarcodeSequence struct {
	Prefix    int    `json:"prefix"`
	LastValue int64  `json:"last_value"`
	UpdateAt  string `json:"update_at,omitempty"`
}
