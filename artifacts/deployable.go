package artifacts

type Deployable interface {
	Deploy(path string) error
}
