package transformation

type OrderType struct {
	value string
}

func NewOrderType(value string) (OrderType, error) {
	if value == "" {
		return OrderType{value: "STANDARD"}, nil
	}

	return OrderType{value: value}, nil
}

func (o OrderType) String() string {
	return o.value
}

func (o OrderType) Equals(other OrderType) bool {
	return o.value == other.value
}
