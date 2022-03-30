package placement

type Info struct {
	ID   string
	Flag bool
}

type Placement interface {
	Select(string) Info
	Append(Info)
	Remove(Info)
}

func Global() Placement {
	return globalPlacement
}
