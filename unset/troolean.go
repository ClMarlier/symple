package unset

type troolean int

const (
	unset troolean = iota
	triFalse
	triTrue
)

func toTroolean(b bool) troolean {
	if b {
		return triTrue
	} else {
		return triFalse
	}
}
