package operator

type Operator interface {
	Judge(payload interface{})
}
