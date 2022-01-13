package statem

type StateContext interface {
	StateCliet()
	PubsubClient()
	SearchClient()
}
