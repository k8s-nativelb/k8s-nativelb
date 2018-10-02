package tests

const (
	TestNamespace = "nativelb-tests-namespace"
)

func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
