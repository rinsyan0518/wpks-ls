package portout

type PackwerkRunner interface {
	RunCheck(uri string) (string, error)
}
