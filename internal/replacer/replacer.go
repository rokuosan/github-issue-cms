package replacer

type Replacer interface {
	Replace(string) (string, error)
}
