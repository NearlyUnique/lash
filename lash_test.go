package lash

var testErrors []error

func init() {
	DefaultSession.OnError(func(e error) {
		testErrors = append(testErrors, e)
	})
}
