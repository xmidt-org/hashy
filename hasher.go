package hashy

type Hasher interface {
	Get([]byte) (string, error)
}
