package artifacts

import "context"

type Executable interface {
	Execute(context.Context) error
}
