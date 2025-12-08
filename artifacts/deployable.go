package artifacts

type Deployable interface {
	WriteScript(path string) error
}
