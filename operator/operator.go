package operator

import "net/http"

type Operator interface {
	Judge(payload interface{},header http.Header)
}
