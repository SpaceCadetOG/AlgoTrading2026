package risk

type Decision struct {
	Approved bool
	Reason   string
}

func approve(reason string) Decision {
	return Decision{
		Approved: true,
		Reason:   reason,
	}
}

func reject(reason string) Decision {
	return Decision{
		Approved: false,
		Reason:   reason,
	}
}
